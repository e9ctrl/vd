package api

import (
	"fmt"
	"io"
	"net/http"
	"time"
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

func (c *client) GetGlobalDelay(typ string) (time.Duration, error) {
	resp, err := http.Get("http://" + c.url + "/delay/" + typ)
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

func (c *client) SetGlobalDelay(typ, value string) error {
	resp, err := http.Post("http://"+c.url+"/delay/"+typ+"/"+value, "text/plain", nil)
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

func (c *client) GetParamDelay(param, typ string) (time.Duration, error) {
	resp, err := http.Get("http://" + c.url + "/delay/" + typ + "/" + param)
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

func (c *client) SetParamDelay(param, typ, value string) error {
	resp, err := http.Post("http://"+c.url+"/delay/"+typ+"/"+param+"/"+value, "text/plain", nil)
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
