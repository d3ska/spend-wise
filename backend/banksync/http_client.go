package banksync

import (
	"bytes"
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/google/uuid"
)

// HTTPClient implements the Client interface using the Enable Banking API.
type HTTPClient struct {
	baseURL       string
	applicationID string
	privateKey    *rsa.PrivateKey
	httpClient    *http.Client
}

// NewHTTPClient creates a new Enable Banking API client.
func NewHTTPClient(baseURL, applicationID, keyPath string) (*HTTPClient, error) {
	key, err := loadPrivateKey(keyPath)
	if err != nil {
		return nil, fmt.Errorf("loading private key: %w", err)
	}

	return &HTTPClient{
		baseURL:       baseURL,
		applicationID: applicationID,
		privateKey:    key,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

func loadPrivateKey(path string) (*rsa.PrivateKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading key file: %w", err)
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("failed to parse PEM block")
	}

	switch block.Type {
	case "RSA PRIVATE KEY":
		return x509.ParsePKCS1PrivateKey(block.Bytes)
	case "PRIVATE KEY":
		key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, err
		}
		rsaKey, ok := key.(*rsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("PKCS8 key is not RSA")
		}
		return rsaKey, nil
	default:
		return nil, fmt.Errorf("unsupported PEM type: %s", block.Type)
	}
}

func (c *HTTPClient) generateJWT() (string, error) {
	header, err := json.Marshal(map[string]string{
		"alg": "RS256",
		"typ": "JWT",
		"kid": c.applicationID,
	})
	if err != nil {
		return "", err
	}

	now := time.Now().Unix()
	payload, err := json.Marshal(map[string]any{
		"iss": "enablebanking.com",
		"aud": "api.enablebanking.com",
		"iat": now,
		"exp": now + 3600,
	})
	if err != nil {
		return "", err
	}

	encodedHeader := base64.RawURLEncoding.EncodeToString(header)
	encodedPayload := base64.RawURLEncoding.EncodeToString(payload)
	signingInput := encodedHeader + "." + encodedPayload

	hash := sha256.Sum256([]byte(signingInput))
	signature, err := rsa.SignPKCS1v15(rand.Reader, c.privateKey, crypto.SHA256, hash[:])
	if err != nil {
		return "", fmt.Errorf("signing JWT: %w", err)
	}

	encodedSignature := base64.RawURLEncoding.EncodeToString(signature)
	return signingInput + "." + encodedSignature, nil
}

// apiError wraps API error responses.
type apiError struct {
	StatusCode int
	Body       string
}

func (e *apiError) Error() string {
	return fmt.Sprintf("Enable Banking API error (status %d): %s", e.StatusCode, e.Body)
}

func (c *HTTPClient) doRequest(ctx context.Context, method, path string, reqBody, respBody any) error {
	slog.Debug("EnableBanking: sending request", "method", method, "path", path)

	var bodyReader io.Reader
	if reqBody != nil {
		data, err := json.Marshal(reqBody)
		if err != nil {
			return fmt.Errorf("marshaling request: %w", err)
		}
		slog.Debug("EnableBanking: request body", "size_bytes", len(data))
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, bodyReader)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	jwt, err := c.generateJWT()
	if err != nil {
		return fmt.Errorf("generating JWT: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+jwt)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		slog.Error("EnableBanking: request failed", "method", method, "path", path, "error", err)
		return fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	// Limit response body to 10 MB to prevent unbounded memory consumption
	// from a misbehaving upstream API.
	const maxResponseBody = 10 << 20 // 10 MB
	respData, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBody))
	if err != nil {
		return fmt.Errorf("reading response: %w", err)
	}

	slog.Debug("EnableBanking: response received", "method", method, "path", path, "status", resp.StatusCode, "size_bytes", len(respData))

	if resp.StatusCode >= 400 {
		slog.Warn("EnableBanking: API error", "method", method, "path", path, "status", resp.StatusCode)
		return &apiError{StatusCode: resp.StatusCode, Body: string(respData)}
	}

	if respBody != nil {
		if err := json.Unmarshal(respData, respBody); err != nil {
			slog.Error("EnableBanking: failed to decode response", "method", method, "path", path, "error", err)
			return fmt.Errorf("decoding response: %w", err)
		}
	}

	return nil
}

// aspspsResponse wraps the Enable Banking ASPSPs endpoint response.
type aspspsResponse struct {
	ASPSPs []ASPSP `json:"aspsps"`
}

// GetASPSPs returns available banking institutions for a country.
func (c *HTTPClient) GetASPSPs(ctx context.Context, country string) ([]ASPSP, error) {
	path := "/aspsps?country=" + url.QueryEscape(country)

	var resp aspspsResponse
	if err := c.doRequest(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, fmt.Errorf("getting ASPSPs: %w", err)
	}
	return resp.ASPSPs, nil
}

// StartAuth initiates an authorization flow with a bank.
func (c *HTTPClient) StartAuth(ctx context.Context, aspspName, country, redirectURL string, validUntil time.Time) (AuthResult, error) {
	body := map[string]any{
		"access": map[string]string{
			"valid_until": validUntil.Format(time.RFC3339),
		},
		"aspsp": map[string]string{
			"name":    aspspName,
			"country": country,
		},
		"state":        uuid.NewString(),
		"redirect_url": redirectURL,
		"psu_type":     "personal",
	}

	var result AuthResult
	if err := c.doRequest(ctx, http.MethodPost, "/auth", body, &result); err != nil {
		return AuthResult{}, fmt.Errorf("starting auth: %w", err)
	}
	return result, nil
}

// CreateSession exchanges an authorization code for a session with accounts.
func (c *HTTPClient) CreateSession(ctx context.Context, code string) (Session, error) {
	body := map[string]string{
		"code": code,
	}

	var session Session
	if err := c.doRequest(ctx, http.MethodPost, "/sessions", body, &session); err != nil {
		return Session{}, fmt.Errorf("creating session: %w", err)
	}
	return session, nil
}

// transactionsResponse is the Enable Banking transactions endpoint response.
type transactionsResponse struct {
	Transactions    []Transaction `json:"transactions"`
	ContinuationKey string        `json:"continuation_key"`
}

// GetAccountTransactions fetches booked transactions for an account within a date range.
func (c *HTTPClient) GetAccountTransactions(ctx context.Context, accountUID string, dateFrom, dateTo time.Time) ([]Transaction, error) {
	basePath := fmt.Sprintf("/accounts/%s/transactions?date_from=%s&date_to=%s&strategy=longest",
		url.PathEscape(accountUID),
		url.QueryEscape(dateFrom.Format("2006-01-02")),
		url.QueryEscape(dateTo.Format("2006-01-02")),
	)

	var allTransactions []Transaction
	path := basePath

	for {
		var resp transactionsResponse
		if err := c.doRequest(ctx, http.MethodGet, path, nil, &resp); err != nil {
			return nil, fmt.Errorf("getting transactions: %w", err)
		}

		for _, tx := range resp.Transactions {
			if tx.Status == "BOOK" {
				allTransactions = append(allTransactions, tx)
			}
		}

		if resp.ContinuationKey == "" {
			break
		}
		path = basePath + "&continuation_key=" + url.QueryEscape(resp.ContinuationKey)
	}

	return allTransactions, nil
}
