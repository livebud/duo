package wd

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/tebeka/selenium"
	"github.com/tebeka/selenium/chrome"
)

func Dial(url string) (*Browser, error) {
	if !strings.HasPrefix(url, "http://") {
		url = "http://" + url
	}
	var opts []selenium.ServiceOption
	opts = append(opts, selenium.Output(&logFilter{os.Stderr}))
	service, err := selenium.NewChromeDriverService("chromedriver", 4444, opts...)
	if err != nil {
		return nil, fmt.Errorf("error starting the ChromeDriver server: %w", err)
	}
	caps := selenium.Capabilities{}
	chromeCaps := chrome.Capabilities{
		Args: []string{
			"--headless",
			"--disable-gpu",
			"--no-sandbox",
			"--disable-dev-shm-usage",
			"--disable-background-timer-throttling",
			"--disable-backgrounding-occluded-windows",
			"--disable-renderer-backgrounding",
		},
	}
	caps.AddChrome(chromeCaps)
	wd, err := selenium.NewRemote(caps, "")
	if err != nil {
		service.Stop()
		return nil, fmt.Errorf("error creating new browser: %w", err)
	}
	return &Browser{wd, service, url}, nil
}

type Browser struct {
	wd  selenium.WebDriver
	svc *selenium.Service
	url string
}

func (b *Browser) Get(path string) (*Response, error) {
	if err := b.wd.Get(b.url + path); err != nil {
		return nil, err
	}
	if err := b.Ready(5 * time.Second); err != nil {
		return nil, err
	}
	html, err := b.wd.PageSource()
	if err != nil {
		return nil, err
	}
	return &Response{http.StatusOK, bytes.NewBufferString(html)}, nil
}

func (b *Browser) GetHtml(path string, selector ...string) (string, error) {
	res, err := b.Get(path)
	if err != nil {
		return "", err
	}
	if len(selector) == 0 {
		return res.body.String(), nil
	}
	sel, err := res.Find(strings.Join(selector, " "))
	if err != nil {
		return "", err
	}
	return sel.Html()
}

func (b *Browser) Click(selector string) error {
	element, err := b.wd.FindElement(selenium.ByCSSSelector, selector)
	if err != nil {
		return err
	}
	return element.Click()
}

func (b *Browser) GetAttribute(selector, attr string) (string, error) {
	element, err := b.wd.FindElement(selenium.ByCSSSelector, selector)
	if err != nil {
		return "", err
	}
	return element.GetAttribute(attr)
}

func (b *Browser) ComputedStyle(selector, property string) (string, error) {
	element, err := b.wd.FindElement(selenium.ByCSSSelector, selector)
	if err != nil {
		return "", err
	}
	return element.CSSProperty(property)
}

// Ready uses the window.prerenderReady check to determine if the page is ready.
func (b *Browser) Ready(timeout time.Duration) error {
	return b.WaitFor(timeout, "window.prerenderReady == true")
}

func (b *Browser) WaitFor(timeout time.Duration, expr string) error {
	return b.wd.WaitWithTimeout(func(d selenium.WebDriver) (bool, error) {
		value, err := b.wd.ExecuteScript("return "+expr, nil)
		if err != nil {
			return false, err
		}
		boolValue, ok := value.(bool)
		if !ok {
			return false, fmt.Errorf("expected bool, got %T", value)
		}
		return boolValue, nil
	}, timeout)
}

func (b *Browser) Close() {
	b.wd.Quit()
	b.svc.Stop()
}

type Response struct {
	status int
	body   *bytes.Buffer
}

func (r *Response) Status() int {
	return r.status
}

func (r *Response) Body() io.Reader {
	return bytes.NewReader(r.body.Bytes())
}

func (r *Response) Contains(substr string) error {
	if !strings.Contains(r.body.String(), substr) {
		return fmt.Errorf("expected %q to contain %q", r.body.String(), substr)
	}
	return nil
}

func (r *Response) NotContains(substr string) error {
	if strings.Contains(r.body.String(), substr) {
		return fmt.Errorf("expected %q to not contain %q", r.body.String(), substr)
	}
	return nil
}

func (r *Response) Find(selector string) (*goquery.Selection, error) {
	doc, err := goquery.NewDocumentFromReader(r.Body())
	if err != nil {
		return nil, err
	}
	return doc.Find(selector), nil
}
