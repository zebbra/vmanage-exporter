package vmanage

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/go-resty/resty/v2"
	"io"
	"math"
	"math/big"
	"net/http"
	"net/http/cookiejar"
	"net/url"
)

type Client struct {
	BaseURL         string
	Username        string
	Password        string
	Session         *http.Cookie
	Token           string
	TLSClientConfig *tls.Config
}

type FetchOptions interface {
	Params() url.Values
}

func (c *Client) Login() error {
	if c.Session != nil {
		_ = c.Logout()
	}

	jar, _ := cookiejar.New(nil)
	client := &http.Client{
		Jar: jar,
		Transport: &http.Transport{
			TLSClientConfig: c.TLSClientConfig,
		},
	}

	loginResp, err := client.PostForm(
		c.BaseURL+"/j_security_check",
		url.Values{"j_username": {c.Username}, "j_password": {c.Password}},
	)

	if err != nil {
		return err
	}

	defer loginResp.Body.Close()

	for _, cookie := range loginResp.Cookies() {
		if cookie.Name == "JSESSIONID" {
			c.Session = cookie
		}
	}

	if loginResp.StatusCode != http.StatusOK || c.Session == nil {
		return errors.New("Login error")
	}

	// fetch token
	tokenResp, err := client.Get(c.BaseURL + "/dataservice/client/token")

	if err != nil {
		return fmt.Errorf("Error fetching token: %w", err)
	}

	defer tokenResp.Body.Close()

	t, err := io.ReadAll(tokenResp.Body)

	if err != nil {
		return err
	}

	c.Token = string(t)
	return nil
}

func (c *Client) Request() (*resty.Request, error) {
	if c.Session == nil || c.Token == "" {
		if err := c.Login(); err != nil {
			return nil, fmt.Errorf("Login failed: %w", err)
		}
	}

	rc := resty.New()
	rc.DisableWarn = true
	rc.SetCookie(c.Session)
	rc.SetHeader("X-XSRF-TOKEN", c.Token)
	rc.SetTLSClientConfig(c.TLSClientConfig)

	return rc.R(), nil
}

func (c *Client) Fetch(ctx context.Context, endpoint string, options FetchOptions, results interface{}) (interface{}, error) {
	r, err := c.Request()

	if err != nil {
		return nil, err
	}

	r.SetContext(ctx)

	endpoint = c.BaseURL + endpoint

	if options != nil {
		endpoint += "?" + options.Params().Encode()
	}

	resp, err := r.SetResult(results).Get(endpoint)

	if err != nil {
		return nil, err
	}

	if resp.IsError() {
		return nil, fmt.Errorf("%s: %s", resp.Status(), resp.String())
	}

	return resp.Result(), nil
}

func (c *Client) Logout() error {
	if c.Session == nil {
		c.Token = ""
		return nil
	}

	r, err := c.Request()

	if err != nil {
		return err
	}

	rnd, _ := rand.Int(rand.Reader, big.NewInt(int64(math.Pow10(9))))
	resp, err := r.Get(fmt.Sprintf("%s/logout?nocache=%s", c.BaseURL, rnd))

	c.Token = ""
	c.Session = nil

	// If the http response code is 302 redirect with location header
	// https://{vmanage-ip-address}/welcome.html?nocache=, the session has been invalidated.
	// Otherwise, an error occurred in the session invalidation process.
	if resp.RawResponse.Request.URL.Path != "/welcome.html" {
		return errors.New("Logout did not return redirect to welcome.html")
	}

	return nil
}

func NewClient(baseURL string, username string, password string) *Client {
	return &Client{
		BaseURL:         baseURL,
		Username:        username,
		Password:        password,
		TLSClientConfig: &tls.Config{},
	}
}
