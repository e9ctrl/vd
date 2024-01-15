package api

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

// Structure with client configuration.
type Client struct {
	url string
}

// Create new Client isntance with configurable HTTP address.
func NewClient(url string) *Client {
	return &Client{
		url: url,
	}
}

// Get given parameter name from the simulator server via exposed REST API with HTTP GET query.
func (c *Client) GetParameter(param string) (string, error) {
	resp, err := http.Get("http://" + c.url + "/" + param)
	if err != nil {
		return "", err
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API error %s", body)
	}
	return string(body), nil
}

// Set given parameter with a specified value via exposed REST API with HTTP POST query.
func (c *Client) SetParameter(param, value string) error {
	resp, err := http.Post("http://"+c.url+"/"+param+"/"+value, "text/plain", nil)
	if err != nil {
		return err
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API error %s", body)
	}

	return nil
}

// Get command delay value via exposed REST API with HTTP Get query.
func (c *Client) GetCommandDelay(commandName string) (time.Duration, error) {
	resp, err := http.Get("http://" + c.url + "/delay/" + commandName)
	if err != nil {
		return 0, err
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("API error %s", body)
	}

	return time.ParseDuration(string(body))
}

// Set command delay via exposed REST aPI with HTTP Post query.
func (c *Client) SetCommandDelay(commandName, value string) error {
	resp, err := http.Post("http://"+c.url+"/delay/"+commandName+"/"+value, "text/plain", nil)
	if err != nil {
		return err
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API error %s", body)
	}

	return nil
}

// Get mismatch string (message that is returned when ) via exposed REST API with Get query.
func (c *Client) GetMismatch() (string, error) {
	resp, err := http.Get("http://" + c.url + "/mismatch")
	if err != nil {
		return "", err
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API error %s", body)
	}
	return string(body), nil
}

// Set new mismatch message via exposed REST API with POST query.
func (c *Client) SetMismatch(value string) error {
	resp, err := http.Post("http://"+c.url+"/mismatch/"+value, "text/plain", nil)
	if err != nil {
		return err
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API error %s", body)
	}

	return nil
}

// Method to trigger returning parameter value on the TCP server side, uses HTTP Post query.
func (c *Client) Trigger(param string) error {
	resp, err := http.Post("http://"+c.url+"/trigger/"+param, "text/plain", nil)
	if err != nil {
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API error %s", body)
	}
	return nil
}
