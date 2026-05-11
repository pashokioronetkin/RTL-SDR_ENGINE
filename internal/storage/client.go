package storage

import (
	"RTL-SDR/engine/internal/models"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Client struct {
	baseURL string
	httpCl  *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpCl:  &http.Client{Timeout: 10 * time.Second},
	}
}

// SaveDetectionData отправляет данные в storage API
func (c *Client) SaveDetectionData(data map[string]interface{}) error {
	url := c.baseURL + "/api/detections"
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	resp, err := c.httpCl.Post(url, "application/json", bytes.NewReader(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("storage API error: %d", resp.StatusCode)
	}
	return nil
}

// GetHistory получает последние записи из storage
func (c *Client) GetHistory(limit int) ([]models.Session, error) {
	url := fmt.Sprintf("%s/api/history?limit=%d", c.baseURL, limit)
	resp, err := c.httpCl.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("storage API error: %d", resp.StatusCode)
	}
	var sessions []models.Session
	if err := json.NewDecoder(resp.Body).Decode(&sessions); err != nil {
		return nil, err
	}
	return sessions, nil
}
