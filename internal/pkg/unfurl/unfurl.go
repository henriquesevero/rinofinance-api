package unfurl

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

var ErrInvalidURL = errors.New("url inválida")

type Metadata struct {
	ImageURL string `json:"imageUrl"`
	Title    string `json:"title"`
	Price    string `json:"price"`
}

const browserUA = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0 Safari/537.36"

var (
	metaTagRe  = regexp.MustCompile(`(?is)<meta\b[^>]*>`)
	linkTagRe  = regexp.MustCompile(`(?is)<link\b[^>]*>`)
	titleTagRe = regexp.MustCompile(`(?is)<title[^>]*>(.*?)</title>`)
	ldJSONRe   = regexp.MustCompile(`(?is)<script[^>]+type\s*=\s*["']application/ld\+json["'][^>]*>(.*?)</script>`)

	amazonImgRe = regexp.MustCompile(`https://[a-z0-9.\-]*media-amazon\.com/images/I/[A-Za-z0-9._%\-]+\.(?:jpg|jpeg|png)`)
)

func Fetch(ctx context.Context, rawURL string) (Metadata, error) {
	rawURL = strings.TrimSpace(rawURL)
	u, err := url.Parse(rawURL)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return Metadata{}, ErrInvalidURL
	}

	ctx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return Metadata{}, nil
	}
	req.Header.Set("User-Agent", browserUA)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "pt-BR,pt;q=0.9,en;q=0.8")
	req.Header.Set("Sec-Fetch-Dest", "document")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-Site", "none")
	req.Header.Set("Upgrade-Insecure-Requests", "1")

	resp, err := (&http.Client{Timeout: 10 * time.Second}).Do(req)
	if err != nil {
		return Metadata{}, nil
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode >= 400 {
		return Metadata{}, nil
	}

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 3<<20))
	html := string(body)

	base := resp.Request.URL

	image := firstNonEmpty(
		metaContent(html, "og:image", "og:image:url", "og:image:secure_url", "twitter:image", "twitter:image:src"),
		metaContent(html, "image"),
		jsonLDImage(html),
		linkHref(html, "image_src"),
		amazonImgRe.FindString(html),
		linkHref(html, "apple-touch-icon"),
		linkHref(html, "apple-touch-icon-precomposed"),
	)
	image = absURL(base, image)

	title := metaContent(html, "og:title", "twitter:title")
	if title == "" {
		if m := titleTagRe.FindStringSubmatch(html); m != nil {
			title = strings.TrimSpace(m[1])
		}
	}

	price := metaContent(html, "product:price:amount", "og:price:amount", "twitter:data1")

	return Metadata{ImageURL: image, Title: strings.TrimSpace(title), Price: strings.TrimSpace(price)}, nil
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

func absURL(base *url.URL, raw string) string {
	if raw == "" || base == nil {
		return raw
	}
	if abs, err := base.Parse(raw); err == nil {
		return abs.String()
	}
	return raw
}

func metaContent(html string, keys ...string) string {
	for _, tag := range metaTagRe.FindAllString(html, -1) {
		key := firstNonEmpty(attrValue(tag, "property"), attrValue(tag, "name"), attrValue(tag, "itemprop"))
		if key == "" {
			continue
		}
		for _, want := range keys {
			if strings.EqualFold(key, want) {
				if c := attrValue(tag, "content"); c != "" {
					return c
				}
			}
		}
	}
	return ""
}

func linkHref(html, rel string) string {
	for _, tag := range linkTagRe.FindAllString(html, -1) {
		for _, r := range strings.Fields(attrValue(tag, "rel")) {
			if strings.EqualFold(r, rel) {
				if h := attrValue(tag, "href"); h != "" {
					return h
				}
			}
		}
	}
	return ""
}

func jsonLDImage(html string) string {
	for _, m := range ldJSONRe.FindAllStringSubmatch(html, -1) {
		var data any
		if err := json.Unmarshal([]byte(strings.TrimSpace(m[1])), &data); err != nil {
			continue
		}
		if img := findImage(data); img != "" {
			return img
		}
	}
	return ""
}

func findImage(v any) string {
	switch t := v.(type) {
	case map[string]any:
		if img, ok := t["image"]; ok {
			if s := imageURL(img); s != "" {
				return s
			}
		}
		for _, val := range t {
			if s := findImage(val); s != "" {
				return s
			}
		}
	case []any:
		for _, val := range t {
			if s := findImage(val); s != "" {
				return s
			}
		}
	}
	return ""
}

func imageURL(img any) string {
	switch t := img.(type) {
	case string:
		if strings.HasPrefix(t, "http") {
			return t
		}
	case []any:
		for _, e := range t {
			if s := imageURL(e); s != "" {
				return s
			}
		}
	case map[string]any:
		for _, key := range []string{"url", "contentUrl"} {
			if s, ok := t[key].(string); ok && strings.HasPrefix(s, "http") {
				return s
			}
		}
	}
	return ""
}

func attrValue(tag, attr string) string {
	re := regexp.MustCompile(`(?is)\b` + attr + `\s*=\s*("([^"]*)"|'([^']*)')`)
	m := re.FindStringSubmatch(tag)
	if m == nil {
		return ""
	}
	if m[2] != "" {
		return m[2]
	}
	return m[3]
}
