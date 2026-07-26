// Package pluggy is an outbound adapter for the Pluggy Open Finance API
// (https://docs.pluggy.ai): it authenticates with the client credentials,
// caches the short-lived API key, and reads a connection's accounts and
// transactions so the application layer can mirror them into the user's data.
package pluggy

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

const defaultBaseURL = "https://api.pluggy.ai"

// Client talks to the Pluggy REST API on behalf of the configured
// application. It is safe for concurrent use.
type Client struct {
	clientID     string
	clientSecret string
	baseURL      string
	http         *http.Client

	mu     sync.Mutex
	apiKey string
	expiry time.Time
}

// NewClient builds a Pluggy client from the application's credentials.
func NewClient(clientID, clientSecret string) *Client {
	return &Client{
		clientID:     clientID,
		clientSecret: clientSecret,
		baseURL:      defaultBaseURL,
		http:         &http.Client{Timeout: 30 * time.Second},
	}
}

// Configured reports whether Pluggy credentials were supplied.
func (c *Client) Configured() bool {
	return c.clientID != "" && c.clientSecret != ""
}

// Account mirrors a Pluggy account (a bank account, credit card, etc.).
type Account struct {
	ID            string  `json:"id"`
	ItemID        string  `json:"itemId"`
	Type          string  `json:"type"`    // BANK, CREDIT
	Subtype       string  `json:"subtype"` // CHECKING_ACCOUNT, SAVINGS_ACCOUNT, CREDIT_CARD...
	Name          string  `json:"name"`
	MarketingName string  `json:"marketingName"`
	Number        string  `json:"number"`
	Balance       float64 `json:"balance"`
	CurrencyCode  string  `json:"currencyCode"`
}

// Transaction mirrors a Pluggy transaction. Amount is negative for money
// leaving the account (a debit) and positive for money coming in (a credit);
// Type ("DEBIT"/"CREDIT") says the same thing explicitly.
type Transaction struct {
	ID          string  `json:"id"`
	Description string  `json:"description"`
	Amount      float64 `json:"amount"`
	Date        string  `json:"date"`
	Category    string  `json:"category"`
	CategoryID  string  `json:"categoryId"`
	Type        string  `json:"type"`
}

// Webhook is a Pluggy webhook registration (per application/client).
type Webhook struct {
	ID    string `json:"id"`
	URL   string `json:"url"`
	Event string `json:"event"`
}

// ListWebhooks returns the webhooks registered for this application.
func (c *Client) ListWebhooks(ctx context.Context) ([]Webhook, error) {
	var out struct {
		Results []Webhook `json:"results"`
	}
	if err := c.get(ctx, c.baseURL+"/webhooks", &out); err != nil {
		return nil, err
	}
	return out.Results, nil
}

// CreateWebhook registers a webhook URL for the given event (e.g.
// "item/updated"), so Pluggy notifies us when a connection refreshes.
func (c *Client) CreateWebhook(ctx context.Context, webhookURL, event string) (*Webhook, error) {
	var out Webhook
	body := map[string]string{"url": webhookURL, "event": event}
	if err := c.post(ctx, c.baseURL+"/webhooks", body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Item is a Pluggy connection to a financial institution, carrying the
// connector (the bank/institution) it links to.
type Item struct {
	ID        string `json:"id"`
	Status    string `json:"status"`
	Connector struct {
		ID           int    `json:"id"`
		Name         string `json:"name"`
		ImageURL     string `json:"imageUrl"`
		PrimaryColor string `json:"primaryColor"`
	} `json:"connector"`
}

// GetItem fetches a single connection, used to name synced accounts after
// the institution (e.g. "Itaú") instead of Pluggy's generic account name.
func (c *Client) GetItem(ctx context.Context, itemID string) (*Item, error) {
	var out Item
	u := fmt.Sprintf("%s/items/%s", c.baseURL, url.PathEscape(itemID))
	if err := c.get(ctx, u, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListAccounts returns every account under a Pluggy connection (item).
func (c *Client) ListAccounts(ctx context.Context, itemID string) ([]Account, error) {
	var out struct {
		Results []Account `json:"results"`
	}
	u := fmt.Sprintf("%s/accounts?itemId=%s", c.baseURL, url.QueryEscape(itemID))
	if err := c.get(ctx, u, &out); err != nil {
		return nil, err
	}
	return out.Results, nil
}

// ListTransactions returns every transaction for an account, following
// Pluggy's cursor pagination (the newer /v2 endpoint) to the end.
func (c *Client) ListTransactions(ctx context.Context, accountID string) ([]Transaction, error) {
	base := fmt.Sprintf("%s/v2/transactions?accountId=%s", c.baseURL, url.QueryEscape(accountID))
	var all []Transaction
	next := ""
	// Bound the loop so a misbehaving cursor can never spin forever.
	for page := 0; page < 200; page++ {
		u := base
		if next != "" {
			if strings.HasPrefix(next, "http") {
				u = next
			} else {
				u = base + "&cursor=" + url.QueryEscape(next)
			}
		}
		var out struct {
			Results []Transaction `json:"results"`
			Next    string        `json:"next"`
		}
		if err := c.get(ctx, u, &out); err != nil {
			return nil, err
		}
		all = append(all, out.Results...)
		if strings.TrimSpace(out.Next) == "" {
			break
		}
		next = out.Next
	}
	return all, nil
}

// apiKeyValue returns a valid API key, authenticating (and caching the key)
// when there isn't a fresh one. Keys last ~2h; we refresh a little early.
func (c *Client) apiKeyValue(ctx context.Context) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.apiKey != "" && time.Now().Before(c.expiry) {
		return c.apiKey, nil
	}
	payload, _ := json.Marshal(map[string]string{"clientId": c.clientID, "clientSecret": c.clientSecret})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/auth", bytes.NewReader(payload))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("erro ao autenticar no Pluggy: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		return "", fmt.Errorf("autenticação Pluggy falhou (%d): %s", resp.StatusCode, readSnippet(resp.Body))
	}
	var out struct {
		APIKey string `json:"apiKey"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", fmt.Errorf("erro ao ler resposta de autenticação do Pluggy: %w", err)
	}
	if out.APIKey == "" {
		return "", fmt.Errorf("Pluggy retornou uma apiKey vazia")
	}
	c.apiKey = out.APIKey
	c.expiry = time.Now().Add(100 * time.Minute)
	return c.apiKey, nil
}

// get performs an authenticated GET and decodes the JSON body into out.
func (c *Client) get(ctx context.Context, urlStr string, out any) error {
	key, err := c.apiKeyValue(ctx)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlStr, nil)
	if err != nil {
		return err
	}
	req.Header.Set("X-API-KEY", key)
	req.Header.Set("Accept", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("erro ao chamar o Pluggy: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		return fmt.Errorf("Pluggy retornou %d: %s", resp.StatusCode, readSnippet(resp.Body))
	}
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("erro ao decodificar resposta do Pluggy: %w", err)
	}
	return nil
}

// post performs an authenticated POST with a JSON body and decodes the JSON
// response into out (out may be nil to ignore the body).
func (c *Client) post(ctx context.Context, urlStr string, body any, out any) error {
	key, err := c.apiKeyValue(ctx)
	if err != nil {
		return err
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, urlStr, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("X-API-KEY", key)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("erro ao chamar o Pluggy: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		return fmt.Errorf("Pluggy retornou %d: %s", resp.StatusCode, readSnippet(resp.Body))
	}
	if out != nil {
		if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
			return fmt.Errorf("erro ao decodificar resposta do Pluggy: %w", err)
		}
	}
	return nil
}

func readSnippet(r io.Reader) string {
	b, _ := io.ReadAll(io.LimitReader(r, 512))
	return strings.TrimSpace(string(b))
}
