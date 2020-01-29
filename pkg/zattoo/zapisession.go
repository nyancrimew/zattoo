package zattoo

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/allegro/bigcache"
	"github.com/valyala/fastjson"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"time"
)

const (
	ZapiUrl        = "https://zattoo.com"
	ZapiUUID       = "57fcdb5a-ec0d-4b1f-80f3-e69ef461f08b"
	ZapiAppVersion = "3.2004.2"
)

var appTokenRe = regexp.MustCompile(`window\.appToken\s*=\s'(.*?)'`)

type ZapiSession struct {
	user        string
	password    string
	accountData *fastjson.Value
	client      *http.Client
	cache       *bigcache.BigCache
}

func NewZapiSession(user, password string) (*ZapiSession, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}
	cache, err := bigcache.NewBigCache(bigcache.DefaultConfig(48 * time.Hour))
	if err != nil {
		return nil, err
	}
	session := ZapiSession{
		user:     user,
		password: password,
		client: &http.Client{
			Jar: jar,
		},
		cache:      cache,
	}
	return &session, session.RenewSession()
}

func (z *ZapiSession) ExecZapiCall(api string, params *url.Values, flags... bool) (*fastjson.Value, error) {
	fmt.Println(api)
	var resp *http.Response
	var err error
	if params == nil {
		var req *http.Request
		req, err = http.NewRequest(http.MethodGet, ZapiUrl + api, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Add("Accept", "application/json")
		resp, err = z.client.Do(req)
	} else {
		var req *http.Request
		req, err = http.NewRequest(http.MethodPost, ZapiUrl + api, bytes.NewBufferString(params.Encode()))
		if err != nil {
			return nil, err
		}
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Add("Accept", "application/json")
		resp, err = z.client.Do(req)
	}
	if err != nil || resp.StatusCode < 200 || resp.StatusCode >= 400 {
		if resp != nil {
			body,_ := ioutil.ReadAll(resp.Body)
		fmt.Println(string(body))

		}
		// Renew session and try again
		flagSessionContext := len(flags) > 0 && flags[0]
		flagRenewed := len(flags) > 1 && flags[1]
		if (resp != nil && resp.StatusCode != 503 && resp.StatusCode != 422) || (!flagSessionContext && !flagRenewed) {
			err = z.RenewSession()
			if err != nil {
				return nil, err
			}
			return z.ExecZapiCall(api, params, flagSessionContext, true)
		}
		return nil, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	values, err := fastjson.ParseBytes(body)
	if err != nil {
		// Renew session and try again
		flagSessionContext := len(flags) > 0 && flags[0]
		flagRenewed := len(flags) > 1 && flags[1]
		if !flagSessionContext && !flagRenewed {
			err = z.RenewSession()
			if err != nil {
				return nil, err
			}
			return z.ExecZapiCall(api, params, flagSessionContext, true)
		}
		return nil, err
	}
	if !values.Exists("success") || values.GetBool("success") {
		return values, nil
	}
	// Renew session and try again
	flagSessionContext := len(flags) > 0 && flags[0]
	flagRenewed := len(flags) > 1 && flags[1]
	fmt.Println(string(body))
	if !flagSessionContext && !flagRenewed {
		err = z.RenewSession()
		if err != nil {
			return nil, err
		}
		return z.ExecZapiCall(api, params, flagSessionContext, true)
	}
	fmt.Println(string(body))
	return nil, errors.New("api call failed")
}

func (z *ZapiSession) fetchAppToken() (string, error) {
	resp, err := z.client.Get(ZapiUrl)
	if err != nil {
		return "", err
	}
	html, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	matches := appTokenRe.FindSubmatch(html)
	if len(matches) > 1 {
		return string(matches[1]), nil
	}
	return "", errors.New("failed to retrieve new appToken")
}

func (z *ZapiSession) RenewSession() error {
	z.client.Jar, _ = cookiejar.New(&cookiejar.Options{})
	err := z.announce()
	if err != nil {
		return err
	}
	return z.login()
}

func (z *ZapiSession) announce() error {
	token, err := z.fetchAppToken()
	if err != nil {
		return err
	}
	_, err = z.ExecZapiCall("/zapi/v2/session/hello", &url.Values{
		"client_app_token": {token},
		"uuid":             {ZapiUUID},
		"lang":             {"en"},
		"app_version":      {ZapiAppVersion},
		"format":           {"json"},
	}, true)
	return err
}

func (z *ZapiSession) login() error {
	var err error
	z.accountData, err = z.ExecZapiCall("/zapi/v2/account/login", &url.Values{
		"login":    {z.user},
		"password": {z.password},
		"remember": {"True"},
	}, true)
	return err
}
