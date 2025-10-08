package client

import (
	"encoding/json"
	"fmt"
	"gophermart/internal/model"
	"net/http"
	"strconv"
	"time"
)

type TooManyRequestsError struct {
	RetryAfter time.Duration
}

func (e *TooManyRequestsError) Error() string {
	return fmt.Sprintf("too many requests, retry after %v", e.RetryAfter)
}

type Client struct {
	BaseURL    string
	HTTPClient *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *Client) GetOrderInfo(orderNumber string) (*model.AccrualResponse, error) {
	url := fmt.Sprintf("%s/api/orders/%s", c.BaseURL, orderNumber)

	resp, err := c.HTTPClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		var order model.AccrualResponse
		if err := json.NewDecoder(resp.Body).Decode(&order); err != nil {
			return nil, err
		}
		return &order, nil

	case http.StatusNoContent:
		return nil, fmt.Errorf("order not found")

	case http.StatusTooManyRequests:
		retryAfter := resp.Header.Get("Retry-After")
		delay, err := strconv.Atoi(retryAfter)
		if err != nil || delay == 0 {
			delay = 60
		}
		return nil, &TooManyRequestsError{RetryAfter: time.Duration(delay) * time.Second}

	default:
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
}