package api

import (
	"fmt"
	"io"
	"net/http"
)

type client struct {
	url string
}

func NewClient(url string) *client {
	return &client{
		url: url,
	}
}

func (c *client) GetParameter(param string) (string, error) {
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

func (c *client) SetParameter(param, value string) error {
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
