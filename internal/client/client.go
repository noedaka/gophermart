package client

import (
	"encoding/json"
	"fmt"
	"gophermart/internal/model"
	"net/http"
	"strconv"
	"time"
)

type Client struct {
	BaseURL    string
	HttpClient *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		BaseURL: baseURL,
		HttpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *Client) GetOrderInfo(orderNumber string) (*model.AccrualResponse, error) {
	url := fmt.Sprintf("%s/api/orders/%s", c.BaseURL, orderNumber)

	resp, err := c.HttpClient.Get(url)
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
		delay, _ := strconv.Atoi(retryAfter)
		if delay == 0 {
			delay = 60
		}
		time.Sleep(time.Duration(delay) * time.Second)
		// Рекурсивно повторяем запрос после паузы
		return c.GetOrderInfo(orderNumber)

	default:
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
}
