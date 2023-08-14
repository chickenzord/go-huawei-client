package eg8145v5

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/chickenzord/go-huawei-client/pkg/js"
	"github.com/rs/zerolog"
)

type Client struct {
	jar *cookiejar.Jar
	h   *http.Client
	log zerolog.Logger

	baseURL   string
	userAgent string
	username  string
	password  string
}

func NewClient(baseURL, username, password string) *Client {
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

	log := zerolog.New(os.Stdout).Level(zerolog.DebugLevel)

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
		log: log,
	}
}

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

func (c *Client) Login() error {
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
		return fmt.Errorf("not set-cookies found")
	}

	return nil
}

func (c *Client) Logout() error {
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

func (c *Client) WithSession(fnSession func(c *Client) error) error {
	if err := c.Login(); err != nil {
		return err
	}
	defer c.Logout()

	if err := fnSession(c); err != nil {
		return err
	}

	return nil
}

func (c *Client) ListUserDevices() ([]UserDevice, error) {
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

	if err := s.EvalJSON("GetUserDevInfoList()", &devices); err != nil {
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
