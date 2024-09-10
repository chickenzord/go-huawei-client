package hn8010ts

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/chickenzord/go-huawei-client/pkg/js"
)

type Client struct {
	jar *cookiejar.Jar
	h   *http.Client
	m   sync.Mutex

	baseURL   string
	userAgent string
	username  string
	password  string
}

// newClient
// Create a new client.
func newClient(baseURL, username, password string) *Client {
	jar, err := cookiejar.New(nil)
	if err != nil {
		panic(err)
	}

	u, err := url.Parse(baseURL)
	if err != nil {
		panic(err)
	}

	jar.SetCookies(u, []*http.Cookie{
		{
			Name:  "Cookie",
			Value: "body:Language:english:id=-1",
		},
	})

	return &Client{
		userAgent: "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:109.0) Gecko/20100101 Firefox/115.0",
		baseURL:   baseURL,
		username:  username,
		password:  password,

		jar: jar,
		h: &http.Client{
			Jar:       jar,
			Timeout:   5 * time.Second,
			Transport: http.DefaultTransport,
		},
		m: sync.Mutex{},
	}
}

// NewClient
// Create a new client.
func NewClient(cfg Config) *Client {
	return newClient(cfg.URL, cfg.Username, cfg.Password)
}

// GetHardwareToken
// Get the generated random number to be used in authentication
func (c *Client) GetHardwareToken() (string, error) {
	req, err := http.NewRequest(http.MethodPost, c.baseURL+"/asp/GetRandCount.asp", nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Referer", c.baseURL)

	res, err := c.h.Do(req)
	if err != nil {
		return "", err
	}

	token, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	rawToken := strings.TrimSpace(string(token))

	return rawToken[len(rawToken)-48:], nil
}

func (c *Client) Validate() error {
	if c.baseURL == "" {
		return fmt.Errorf("URL is not set")
	}

	if c.username == "" {
		return fmt.Errorf("username is not set")
	}

	if c.password == "" {
		return fmt.Errorf("password is not set")
	}

	return nil
}

// Login
// Authenticate using saved username/password.
// Authentication cookies will be persisted in the lifetime of Client.
func (c *Client) Login() error {
	c.m.Lock()
	defer c.m.Unlock()

	if err := c.Validate(); err != nil {
		return fmt.Errorf("config error: %w", err)
	}

	hwToken, err := c.GetHardwareToken()
	if err != nil {
		return err
	}

	params := url.Values{}
	params.Set("UserName", c.username)
	params.Set("PassWord", base64.StdEncoding.EncodeToString([]byte(c.password)))
	params.Set("Language", "english")
	params.Set("x.X_HW_Token", hwToken)

	req, err := http.NewRequest(http.MethodPost, c.baseURL+"/login.cgi", strings.NewReader(params.Encode()))
	if err != nil {
		return err
	}

	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Origin", c.baseURL)
	req.Header.Set("Referer", c.baseURL+"/")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Content-Length", strconv.Itoa(len(params.Encode())))
	req.Header.Set("Pragma", "no-cache")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Accept-Encoding", "gzip, deflate")

	res, err := c.h.Do(req)
	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusOK {
		resBody, err := io.ReadAll(res.Body)
		if err != nil {
			return err
		}

		return fmt.Errorf("http %d: %s", res.StatusCode, string(resBody))
	}

	if len(res.Cookies()) == 0 {
		return fmt.Errorf("login failed")
	}

	return nil
}

// Logout
// End Client's session and clear authentication cookies.
func (c *Client) Logout() error {
	c.m.Lock()
	defer c.m.Unlock()

	if err := c.Validate(); err != nil {
		return fmt.Errorf("config error: %w", err)
	}

	hwToken, err := c.GetHardwareToken()
	if err != nil {
		return err
	}

	params := url.Values{}
	params.Set("x.X_HW_Token", hwToken)

	req, err := http.NewRequest(http.MethodPost, c.baseURL+"/logout.cgi?RequestFile=html/logout.html", strings.NewReader(params.Encode()))
	if err != nil {
		return err
	}

	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Origin", c.baseURL)
	req.Header.Set("Referer", c.baseURL+"/index.asp")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Content-Length", strconv.Itoa(len(params.Encode())))
	req.Header.Set("Pragma", "no-cache")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Accept-Encoding", "gzip, deflate")

	res, err := c.h.Do(req)
	if err != nil {
		if errors.Is(err, io.EOF) {
			return nil
		}

		return err
	}

	if res.StatusCode != http.StatusOK {
		resBody, err := io.ReadAll(res.Body)
		if err != nil {
			return err
		}

		return fmt.Errorf("http %d: %s", res.StatusCode, string(resBody))
	}

	return nil
}

// Session
// Run the fnSession function wrapped in Login and Logout
func (c *Client) Session(fnSession func(c *Client) error) error {
	if err := c.Login(); err != nil {
		return err
	}
	defer c.Logout()

	if err := fnSession(c); err != nil {
		return err
	}

	return nil
}

// ListUserDevices
// Get all user devices. Client must be authenticated.
func (c *Client) ListUserDevices() ([]UserDevice, error) {
	c.m.Lock()
	defer c.m.Unlock()

	req, err := http.NewRequest(http.MethodPost, c.baseURL+"/html/bbsp/common/GetLanUserDevInfo.asp", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Referer", c.baseURL+"/html/bbsp/userdevinfo/userdevinfo.asp")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")

	res, err := c.h.Do(req)
	if err != nil {
		return nil, err
	}

	jsPayload, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	s := js.Script{
		Name:    "userdevinfo.asp.js",
		Content: string(jsPayload),
	}

	var devices []*UserDevice

	if err := s.EvalJSON(&devices, "GetUserDevInfoList"); err != nil {
		return nil, err
	}

	var result []UserDevice

	for _, dev := range devices {
		if dev != nil {
			result = append(result, *dev)
		}
	}

	return result, nil
}

// GetResourceUsage
// Get current resource usages. Client must be authenticated.
func (c *Client) GetResourceUsage() (*ResourceUsage, error) {
	c.m.Lock()
	defer c.m.Unlock()

	req, err := http.NewRequest(http.MethodGet, c.baseURL+"/html/ssmp/deviceinfo/deviceinfo.asp", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Referer", c.baseURL)

	res, err := c.h.Do(req)
	if err != nil {
		return nil, err
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, err
	}

	scriptContent := doc.Find("script:not([src])").First().Text()
	if scriptContent == "" {
		return nil, fmt.Errorf("cannot find the script")
	}

	s := js.Script{
		Name:    "deviceinfo.asp.js",
		Content: scriptContent + ResourceUsageFuncScript,
	}

	var usage ResourceUsage

	if err := s.EvalJSON(&usage, ResourceUsageFuncName); err != nil {
		return nil, err
	}

	return &usage, nil
}

// GetResourceUsage
// Get current resource usages. Client must be authenticated.
func (c *Client) GetOpticInfo() (*OpticInfo, error) {
	c.m.Lock()
	defer c.m.Unlock()

	req, err := http.NewRequest(http.MethodGet, c.baseURL+"/html/amp/opticinfo/opticinfo.asp", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Referer", c.baseURL)

	res, err := c.h.Do(req)
	if err != nil {
		return nil, err
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, err
	}

	scriptContent := doc.Find("script:not([src])").First().Text()
	if scriptContent == "" {
		return nil, fmt.Errorf("cannot find the script")
	}

	s := js.Script{
		Name:    "opticinfo.asp.js",
		Content: scriptContent + OpticInfoFuncScript,
	}

	var opticInfo OpticInfo

	if err := s.EvalJSON(&opticInfo, OpticInfoFuncName); err != nil {
		return nil, err
	}

	return &opticInfo, nil
}
