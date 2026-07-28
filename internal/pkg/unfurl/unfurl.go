// Package unfurl fetches a web page and extracts its Open Graph / Twitter
// metadata (image, title, price) so the wishlist can preview a product from
// just its store link. Best-effort: sites that block bots or render only via
// JS simply return empty fields.
package unfurl

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

// ErrInvalidURL is returned for a missing or non-http(s) URL.
var ErrInvalidURL = errors.New("url inválida")

// Metadata is the subset of a page's metadata the wishlist cares about.
type Metadata struct {
	ImageURL string `json:"imageUrl"`
	Title    string `json:"title"`
	Price    string `json:"price"`
}

const browserUA = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0 Safari/537.36"

var (
	metaTagRe = regexp.MustCompile(`(?is)<meta\b[^>]*>`)
	titleTagRe = regexp.MustCompile(`(?is)<title[^>]*>(.*?)</title>`)
)

// Fetch downloads the page at rawURL and pulls its metadata. Returns
// ErrInvalidURL for a bad URL; network/parse failures yield empty metadata
// with a nil error (the caller treats "not found" as a non-error).
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
	req.Header.Set("Accept", "text/html,application/xhtml+xml")
	req.Header.Set("Accept-Language", "pt-BR,pt;q=0.9,en;q=0.8")

	resp, err := (&http.Client{Timeout: 10 * time.Second}).Do(req)
	if err != nil {
		return Metadata{}, nil
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return Metadata{}, nil
	}

	// Cap the read so a huge/streaming response can't exhaust memory.
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 2<<20)) // 2 MB
	html := string(body)

	image := metaContent(html, "og:image", "og:image:url", "og:image:secure_url", "twitter:image", "twitter:image:src")
	if image != "" {
		if abs, err := u.Parse(strings.TrimSpace(image)); err == nil {
			image = abs.String()
		}
	}

	title := metaContent(html, "og:title", "twitter:title")
	if title == "" {
		if m := titleTagRe.FindStringSubmatch(html); m != nil {
			title = strings.TrimSpace(m[1])
		}
	}

	price := metaContent(html, "product:price:amount", "og:price:amount", "twitter:data1")

	return Metadata{ImageURL: image, Title: strings.TrimSpace(title), Price: strings.TrimSpace(price)}, nil
}

// metaContent returns the content of the first <meta> whose property or name
// matches any of the given keys (case-insensitive).
func metaContent(html string, keys ...string) string {
	for _, tag := range metaTagRe.FindAllString(html, -1) {
		key := attrValue(tag, "property")
		if key == "" {
			key = attrValue(tag, "name")
		}
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

// attrValue extracts an HTML attribute's value (single or double quoted).
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
