package auth

import (
	"encoding/json"
	"fmt"
	"github.com/go-resty/resty/v2"
	log "github.com/sirupsen/logrus"
	"net/http"
	"time"
)

type Client struct {
	client    *resty.Client
	name      string
	password  string
	token     string
	expiresAt time.Time
}

func NewClient(authURL, name, password string) *Client {
	client := resty.New()
	client.SetBaseURL(authURL)

	return &Client{
		client:    client,
		name:      name,
		password:  password,
		token:     "",
		expiresAt: time.Now(),
	}
}

func (c *Client) GetToken() (string, error) {
	if c.token != "" && time.Now().Before(c.expiresAt) {
		return c.token, nil
	}

	resp, err := c.client.R().
		SetBody(map[string]string{
			"name":     c.name,
			"password": c.password,
		}).Post("/login")

	if err != nil {
		return "", fmt.Errorf("c.GetToken - Post: %v", err)
	}

	if resp.IsError() {
		log.Errorf("c.GetToken - Post: %v", resp.Status())

		// try to register new user
		err = c.register()
		if err != nil {
			log.Fatalf("c.GetToken -> c.register: %v", err) // looks like wrong credentials were provided
		}
		return "", ErrTryAgain
	}

	var response struct {
		Token     string `json:"token"`
		ExpiresAt int64  `json:"expires_at"`
	}

	if err = json.Unmarshal(resp.Body(), &response); err != nil {
		return "", fmt.Errorf("c.GetToken -> json.Unmarshal: %v", err)
	}

	c.token = response.Token
	c.expiresAt = time.Unix(response.ExpiresAt, 0)
	return c.token, nil
}

func (c *Client) register() error {
	resp, err := c.client.R().
		SetBody(map[string]string{
			"name":     c.name,
			"password": c.password,
		}).Post("/register")

	if err != nil {
		return fmt.Errorf("c.register: %v", err)
	}

	if resp.IsError() {
		if resp.StatusCode() == http.StatusConflict {
			return ErrAlreadyExists
		}
		return fmt.Errorf("c.register: %v", resp.Status())
	}

	return nil
}
