package proxmox

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"main-service/internal/model"
)

type Client struct {
	httpClient *http.Client
	host string
	node string
	token string
}

type vmListResponse struct {
	Data []model.ProxmoxVM `json:"data"`
}

type taskResponse struct {
	Data string `json:"data"`
}


func NewClient(host, node, token string) *Client {
return &Client{
		httpClient: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
			Timeout: 15 * time.Second,
		},
		host:  host,
		node:  node,
		token: token,
	}
}

func (c *Client) ListVMs(ctx context.Context) ([]model.ProxmoxVM, error) {
	apiURL := fmt.Sprintf("%s/api2/json/nodes/%s/qemu", c.host, c.node)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "PVEAPIToken="+c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("proxmox hatası (%d): %s", resp.StatusCode, string(body))
	}

	var res vmListResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, err
	}

	return res.Data, nil

}

func (c *Client) VMAction(ctx context.Context, vmid int, action string) (string, error) {
	apiURL := fmt.Sprintf("%s/api2/json/nodes/%s/qemu/%d/status/%s", c.host, c.node, vmid, action)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "PVEAPIToken="+c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("proxmox hatası (%d): %s", resp.StatusCode, string(body))
	}

	var task taskResponse
	if err := json.NewDecoder(resp.Body).Decode(&task); err != nil {
		return "", err
	}

	return task.Data, nil
}