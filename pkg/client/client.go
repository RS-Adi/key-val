package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type Client struct {
	baseURL string
	http    *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		http:    &http.Client{},
	}
}

type SetRequest struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type GetResponse struct {
	Value string `json:"value"`
}

func (c *Client) Set(key, value string) error {
	reqBody, err := json.Marshal(SetRequest{Key: key, Value: value})
	if err != nil {
		return err
	}

	resp, err := c.http.Post(fmt.Sprintf("%s/set", c.baseURL), "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned status: %d", resp.StatusCode)
	}
	return nil
}

func (c *Client) Get(key string) (string, error) {
	resp, err := c.http.Get(fmt.Sprintf("%s/get?key=%s", c.baseURL, key))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return "", fmt.Errorf("key not found")
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("server returned status: %d", resp.StatusCode)
	}

	var getResp GetResponse
	if err := json.NewDecoder(resp.Body).Decode(&getResp); err != nil {
		return "", err
	}
	return getResp.Value, nil
}

func (c *Client) Delete(key string) error {
	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/delete?key=%s", c.baseURL, key), nil)
	if err != nil {
		return err
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned status: %d", resp.StatusCode)
	}
	return nil
}
