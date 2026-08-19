package asciiapi

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"botwa_go_ascii/config"
)

type Client struct {
	baseURL    string
	apiKey     string
	webURL     string
	httpClient *http.Client
}

func NewClient(cfg *config.Config) *Client {
	baseURL := strings.TrimRight(cfg.AsciiAPIURL, "/")
	return &Client{
		baseURL: baseURL,
		apiKey:  cfg.AsciiAPIKey,
		webURL:  strings.TrimRight(cfg.AsciiWebURL, "/"),
		httpClient: &http.Client{
			Timeout: 12 * time.Second,
		},
	}
}

func (c *Client) WebURL() string {
	return c.webURL
}

func (c *Client) doRequest(method, endpoint string, body io.Reader) ([]byte, error) {
	url := fmt.Sprintf("%s/%s", c.baseURL, strings.TrimLeft(endpoint, "/"))

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	if c.apiKey != "" {
		req.Header.Set("x-api-key", c.apiKey)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request error: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("server returned HTTP %d: %s", resp.StatusCode, string(respBytes))
	}

	return respBytes, nil
}

// GetSchedules retrieves the practical schedules
func (c *Client) GetSchedules() ([]ScheduleItem, error) {
	// Try oRPC route path: /classes/getSchedules
	data, err := c.doRequest(http.MethodGet, "/classes/getSchedules", nil)
	if err != nil {
		// Fallback attempt to alternate path if needed
		data, err = c.doRequest(http.MethodGet, "/jadwal", nil)
		if err != nil {
			return nil, err
		}
	}

	var schedules []ScheduleItem
	if err := json.Unmarshal(data, &schedules); err != nil {
		return nil, fmt.Errorf("failed to parse schedules json: %w", err)
	}

	return schedules, nil
}

// GetClasses retrieves all registered classes
func (c *Client) GetClasses() ([]ClassItem, error) {
	data, err := c.doRequest(http.MethodGet, "/classes/getClasses", nil)
	if err != nil {
		return nil, err
	}

	var classes []ClassItem
	if err := json.Unmarshal(data, &classes); err != nil {
		return nil, fmt.Errorf("failed to parse classes json: %w", err)
	}

	return classes, nil
}

// GetAllModul retrieves all practical modules
func (c *Client) GetAllModul() ([]ModulItem, error) {
	data, err := c.doRequest(http.MethodGet, "/modul/getAll", nil)
	if err != nil {
		return nil, err
	}

	var moduls []ModulItem
	if err := json.Unmarshal(data, &moduls); err != nil {
		return nil, fmt.Errorf("failed to parse modul json: %w", err)
	}

	return moduls, nil
}

// GetAllAnnouncements retrieves announcements
func (c *Client) GetAllAnnouncements() ([]AnnouncementItem, error) {
	data, err := c.doRequest(http.MethodGet, "/announcement/getAll", nil)
	if err != nil {
		return nil, err
	}

	var announcements []AnnouncementItem
	if err := json.Unmarshal(data, &announcements); err != nil {
		return nil, fmt.Errorf("failed to parse announcements json: %w", err)
	}

	return announcements, nil
}

// GetAllBerita retrieves portal news
func (c *Client) GetAllBerita() ([]BeritaItem, error) {
	data, err := c.doRequest(http.MethodGet, "/berita", nil)
	if err != nil {
		return nil, err
	}

	var beritaList []BeritaItem
	if err := json.Unmarshal(data, &beritaList); err != nil {
		return nil, fmt.Errorf("failed to parse berita json: %w", err)
	}

	return beritaList, nil
}

// GetAllContacts retrieves lab contacts & assistants
func (c *Client) GetAllContacts() ([]ContactItem, error) {
	data, err := c.doRequest(http.MethodGet, "/contactPraktikum/getAll", nil)
	if err != nil {
		return nil, err
	}

	var contacts []ContactItem
	if err := json.Unmarshal(data, &contacts); err != nil {
		return nil, fmt.Errorf("failed to parse contacts json: %w", err)
	}

	return contacts, nil
}

// GetAllAturan retrieves lab rules & conduct
func (c *Client) GetAllAturan() ([]AturanItem, error) {
	data, err := c.doRequest(http.MethodGet, "/aturan/getAll", nil)
	if err != nil {
		return nil, err
	}

	var aturan []AturanItem
	if err := json.Unmarshal(data, &aturan); err != nil {
		return nil, fmt.Errorf("failed to parse aturan json: %w", err)
	}

	return aturan, nil
}

// GetAllFAQ retrieves FAQs
func (c *Client) GetAllFAQ() ([]FAQItem, error) {
	data, err := c.doRequest(http.MethodGet, "/faq/getAll", nil)
	if err != nil {
		return nil, err
	}

	var faqs []FAQItem
	if err := json.Unmarshal(data, &faqs); err != nil {
		return nil, fmt.Errorf("failed to parse faq json: %w", err)
	}

	return faqs, nil
}

// Ping checks connectivity with the web backend
func (c *Client) Ping() (time.Duration, error) {
	start := time.Now()
	_, err := c.doRequest(http.MethodGet, "/openapi.json", nil)
	if err != nil {
		return 0, err
	}
	return time.Since(start), nil
}
