// Copyright 2015 Casey Marshall
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/juju/persistent-cookiejar"
	"github.com/koofr/go-httplogger"
	"golang.org/x/crypto/ssh/terminal"
	"golang.org/x/net/publicsuffix"
	"gopkg.in/errgo.v1"
)

func onRedirect(req *http.Request, via []*http.Request) error {
	return nil
}

// homeDir returns the OS-specific home path as specified in the environment.
func homeDir() string {
	if runtime.GOOS == "windows" {
		return filepath.Join(os.Getenv("HOMEDRIVE"), os.Getenv("HOMEPATH"))
	}
	return os.Getenv("HOME")
}

// cookieFile returns the cookie filename to use for persisting cookie data.
//
// The following names will be used in decending order of preference:
//	- the value of the $USSO_COOKIES environment variable.
//	- $HOME/.usso-cookies
func cookieFile() string {
	if f := os.Getenv("USSO_COOKIES"); f != "" {
		return f
	}
	return filepath.Join(homeDir(), ".usso-cookies")
}

func main() {
	loginURL := "https://login.ubuntu.com"
	if len(os.Args) > 1 {
		loginURL = os.Args[1]
	}

	client, err := newClient()
	if err != nil {
		die(errgo.NoteMask(err, "failed to create client"))
	}

	die(client.login(loginURL))
}

func die(err error) {
	if err != nil {
		log.Fatal(errgo.Details(err))
	}
	os.Exit(0)
}

type client struct {
	*http.Client
	Jar *cookiejar.Jar
}

func newClient() (*client, error) {
	jar, err := cookiejar.New(&cookiejar.Options{
		PublicSuffixList: publicsuffix.List,
		Filename:         cookieFile(),
	})
	if err != nil {
		return nil, errgo.NoteMask(err, "failed to create cookiejar")
	}
	defer jar.Save()

	cl := &http.Client{
		Jar: jar,
	}
	if os.Getenv("USSO_DEBUG") == "1" {
		cl.Transport = httplogger.New(http.DefaultTransport)
	}
	return &client{Client: cl, Jar: jar}, nil
}

var ErrAlreadyLoggedIn = errgo.New("already logged in")

func (c *client) login(urlStr string) error {
	defer c.Jar.Save()

	resp, err := c.doLogin(urlStr)
	if errgo.Cause(err) == ErrAlreadyLoggedIn {
		log.Println("already logged in")
		return nil
	} else if err != nil {
		return errgo.NoteMask(err, "login failed")
	}
	result, err := ioutil.ReadAll(resp.Body)
	if clErr := resp.Body.Close(); clErr != nil {
		log.Println("warning: failed to close response:", clErr)
	}
	if err != nil {
		return errgo.NoteMask(err, "failed to read response")
	}

	err = c.loginResponse(resp.Request.URL, result)
	if err != nil {
		return errgo.NoteMask(err, "login failed")
	}

	return nil
}

func (c *client) doLogin(urlStr string) (*http.Response, error) {
	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return nil, errgo.NoteMask(err, "failed to create request")
	}
	resp, err := c.Do(req)
	if err != nil {
		return nil, errgo.Notef(err, "failed to request %q", urlStr)
	}
	referer := resp.Request.URL

	if referer.Host != "login.ubuntu.com" {
		return resp, nil
	}

	result, err := ioutil.ReadAll(resp.Body)
	if clErr := resp.Body.Close(); clErr != nil {
		log.Println("warning: failed to close response:", clErr)
	}
	if err != nil {
		return nil, errgo.NoteMask(err, "failed to read response")
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewBuffer(result))
	if err != nil {
		return nil, errgo.NoteMask(err, "failed to parse login page")
	}

	if doc.Find("#logout-link").Length() > 0 {
		// If we can log out, we must be logged in!
		return nil, ErrAlreadyLoggedIn
	}

	form := doc.Find("#login-form")
	if form.Length() == 0 {
		return nil, errgo.NoteMask(err, "failed to locate form")
	}
	action, ok := form.Attr("action")
	if !ok {
		return nil, errgo.NoteMask(err, "failed to locate submit action")
	}
	token, ok := form.Find(`[name="csrfmiddlewaretoken"]`).Attr("value")
	if !ok {
		return nil, errgo.NoteMask(err, "failed to locate hidden form field")
	}
	postURL := *resp.Request.URL
	postURL.Path = action
	postURL.RawQuery = ""
	postURL.Fragment = ""

	email, err := inputEmail()
	if err != nil {
		return nil, errgo.NoteMask(err, "input failed: email")
	}
	password, err := inputPassword()
	if err != nil {
		return nil, errgo.NoteMask(err, "input failed: password")
	}
	req, err = http.NewRequest("POST", postURL.String(), strings.NewReader(url.Values{
		"csrfmiddlewaretoken":   []string{token},
		"email":                 []string{email},
		"password":              []string{password},
		"openid.usernamesecret": []string{""},
	}.Encode()))
	if err != nil {
		return nil, errgo.NoteMask(err, "failed to create login form submit request")
	}
	req.Header.Add("Referer", referer.String())
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	return c.Do(req)
}

func inputEmail() (string, error) {
	if email := os.Getenv("USSO_EMAIL"); email != "" {
		return email, nil
	}
	fmt.Print("Email: ")
	return bufio.NewReader(os.Stdin).ReadString('\n')
}

func inputPassword() (string, error) {
	if password := os.Getenv("USSO_PASSWORD"); password != "" {
		return password, nil
	}
	return inputSecret("Password: ")
}

func inputTwoFactor() (string, error) {
	return inputSecret("Two-factor Auth: ")
}

func inputSecret(prompt string) (string, error) {
	state, err := terminal.MakeRaw(0)
	if err != nil {
		return "", errgo.Mask(err, errgo.Any)
	}
	defer terminal.Restore(0, state)
	term := terminal.NewTerminal(os.Stdout, ">")
	secret, err := term.ReadPassword(prompt)
	if err != nil {
		return "", errgo.Mask(err, errgo.Any)
	}
	return secret, nil
}

func (c *client) loginResponse(referer *url.URL, result []byte) error {
	var doc *goquery.Document
	var err error
	for {
		doc, err = goquery.NewDocumentFromReader(bytes.NewBuffer(result))
		if err != nil {
			return errgo.NoteMask(err, "failed to parse response")
		}

		twoFactorInput := doc.Find("#id_oath_token")
		if twoFactorInput.Length() == 0 {
			break
		}
		resp, err := c.doTwoFactor(referer, doc)
		if err != nil {
			return errgo.Mask(err, errgo.Any)
		}
		referer = resp.Request.URL
		result, err = ioutil.ReadAll(resp.Body)
		if clErr := resp.Body.Close(); clErr != nil {
			log.Println("warning: failed to close response:", clErr)
		}
		if err != nil {
			return errgo.NoteMask(err, "failed to read response")
		}
	}
	if referer.Host != "login.ubuntu.com" {
		// We've been redirected away from login.ubuntu.com, so we've probably
		// logged in...
		return nil
	}
	if doc.Find("#logout-link").Length() > 0 {
		// If we can log out, we must be logged in!
		return nil
	}
	return errgo.New("failed to confirm login")
}

func (c *client) doTwoFactor(referer *url.URL, doc *goquery.Document) (*http.Response, error) {
	form := doc.Find("#login-form")
	if form.Length() == 0 {
		return nil, errgo.New("failed to locate form")
	}

	action, ok := form.Attr("action")
	if !ok {
		return nil, errgo.New("failed to locate submit action")
	}
	token, ok := form.Find(`[name="csrfmiddlewaretoken"]`).Attr("value")
	if !ok {
		return nil, errgo.New("failed to locate hidden form field")
	}
	postURL := *referer
	if action != "" {
		postURL.Path = action
	}
	postURL.RawQuery = ""
	postURL.Fragment = ""

	twoFactorAuth, err := inputTwoFactor()
	if err != nil {
		return nil, errgo.NoteMask(err, "input failed: two-factor auth")
	}
	req, err := http.NewRequest("POST", postURL.String(), strings.NewReader(url.Values{
		"csrfmiddlewaretoken":   []string{token},
		"oath_token":            []string{twoFactorAuth},
		"openid.usernamesecret": []string{""},
	}.Encode()))
	if err != nil {
		return nil, errgo.NoteMask(err, "failed to create login form submit request")
	}
	req.Header.Add("Referer", referer.String())
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	return c.Do(req)
}
