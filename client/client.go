// /home/krylon/go/src/github.com/blicero/mistwetter/client/client.go
// -*- mode: go; coding: utf-8; -*-
// Created on 08. 05. 2023 by Benjamin Walkenhorst
// (c) 2023 Benjamin Walkenhorst
// Time-stamp: <2023-05-09 09:54:30 krylon>

// Package client implements the interaction with the DWD web API and the
// processing of the JSON data.
package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"sync"
	"time"

	"github.com/blicero/mistwetter/common"
	"github.com/blicero/mistwetter/logdomain"
)

// The interval between checks is deliberately short for now to make testing
// easier. Once we reach a stable state, we need to increase this interval to
// something like 10 or 15 minutes.
const (
	warnURL       = "https://www.dwd.de/DWD/warnungen/warnapp/json/warnings.json"
	checkInterval = time.Second * 30
)

// The response from the DWD's web service looks like this:
// warnWetter.loadWarnings({"time":1627052765000,"warnings":{},"vorabInformation":{},"copyright":"Copyright Deutscher Wetterdienst"});

var respPattern = regexp.MustCompile(`^warnWetter[.]loadWarnings\((.*)\);`)

// Client implements the communication with the DWD's web service and the handling of the response.
type Client struct {
	lastStamp    int64
	active       bool
	lock         sync.RWMutex
	log          *log.Logger
	client       http.Client
	locations    []*regexp.Regexp
	WarnQueue    chan Warning
	refreshQueue chan int
	stopQueue    chan int
}

// New creates a new Client. If proxy is a non-empty string, it is used as the
// URL of the proxy server to use for accessing the DWD's web service.
func New(proxy string, locations ...string) (*Client, error) {
	var (
		err error
		c   = new(Client)
	)

	if c.log, err = common.GetLogger(logdomain.Client); err != nil {
		return nil, err
	}

	c.WarnQueue = make(chan Warning, 8)
	c.stopQueue = make(chan int)
	c.refreshQueue = make(chan int)
	c.client.Timeout = time.Second * 90

	if proxy != "" {
		var u *url.URL
		if u, err = url.Parse(proxy); err != nil {
			c.log.Printf("[ERROR] Cannot parse proxy URL %q: %s\n",
				proxy,
				err.Error())
			return nil, err
		}

		var pfunc = func(r *http.Request) (*url.URL, error) { return u, nil }

		switch t := c.client.Transport.(type) {
		case *http.Transport:
			t.Proxy = pfunc
		default:
			err = fmt.Errorf("Unexpected type for HTTP Client Transport: %T",
				c.client.Transport)
			fmt.Fprintf(os.Stderr, "%s\n", err.Error())
			return nil, err
		}

	}

	c.locations = make([]*regexp.Regexp, 0, len(locations))

	for _, l := range locations {
		var r *regexp.Regexp

		c.log.Printf("[DEBUG] Add regexp %s\n", l)

		if r, err = regexp.Compile(l); err != nil {
			c.log.Printf("[ERROR] Cannot compile Regexp %q: %s\n",
				l,
				err.Error())
			return nil, err
		}

		c.locations = append(c.locations, r)
	}

	c.log.Printf("[DEBUG] Client has %d regular expressions for matching locations\n",
		len(c.locations))

	return c, nil
} // func New(proxy string) (*Client, error)

// FetchWarning fetches the warning data from the DWD's web service.
func (c *Client) FetchWarning() ([]byte, error) {
	var (
		err error
		res *http.Response
		buf bytes.Buffer
	)

	if res, err = c.client.Get(warnURL); err != nil {
		c.log.Printf("[ERROR] Failed to fetch %q: %s\n",
			warnURL,
			err.Error())
	}

	defer res.Body.Close() // nolint: errcheck

	if res.StatusCode != 200 {
		c.log.Printf("[DEBUG] Response for %q: %s\n",
			warnURL,
			res.Status)
		return nil, fmt.Errorf("HTTP Request to %q failed: %s",
			warnURL,
			res.Status)
	} else if _, err = io.Copy(&buf, res.Body); err != nil {
		c.log.Printf("[ERROR] Cannot read response Body for %q: %s\n",
			warnURL,
			err.Error())
		return nil, err
	}

	var (
		body  = buf.Bytes()
		match [][]byte
	)

	if match = respPattern.FindSubmatch(body); match == nil {
		err = fmt.Errorf("Cannot parse response from %q: %q",
			warnURL,
			body)
		c.log.Printf("[ERROR] %s\n", err.Error())
		return nil, err
	}

	var data = match[1]

	return data, nil
} // func (c *Client) FetchWarning() ([]byte, error)

// ProcessWarnings parses the warnings returned by the DWD's web service and
// returns a list of all the warnings that are relevant to us.
func (c *Client) ProcessWarnings(raw []byte) ([]Warning, error) {
	var (
		err  error
		info WeatherInfo
	)

	if err = json.Unmarshal(raw, &info); err != nil {
		c.log.Printf("[ERROR] Cannot parse JSON data: %s\n%s\n",
			err.Error(),
			raw)
		return nil, err
	}

	var list = make([]Warning, 0, len(c.locations)*2)

	for id, i := range info.Warnings {
	W_ITEM:
		for _, w := range i {
			for _, l := range c.locations {
				if l.MatchString(w.Location) {
					w.ID = id
					list = append(list, w)
					continue W_ITEM
				}
			}
		}
	}

	for id, i := range info.PrelimWarnings {
	V_ITEM:
		for _, w := range i {
			for _, l := range c.locations {
				if l.MatchString(w.Location) {
					w.ID = id
					list = append(list, w)
					continue V_ITEM
				}
			}
		}
	}

	return list, nil
} // func (c *Client) ProcessWarnings(raw []byte) ([]Warning, error)

// GetWarnings loads the current warnings from the DWD and returns all warnings
// matching its list of locations.
func (c *Client) GetWarnings() ([]Warning, error) {
	var (
		err      error
		rawData  []byte
		warnings []Warning
	)

	if rawData, err = c.FetchWarning(); err != nil {
		c.log.Printf("[ERROR] Failed to fetch data from DWD: %s\n",
			err.Error())
		return nil, err
	} else if warnings, err = c.ProcessWarnings(rawData); err != nil {
		c.log.Printf("[ERROR] Failed to process Warnings: %s\n",
			err.Error())
		return nil, err
	}

	return warnings, nil
} // func (c *Client) GetWarnings() ([]Warning, error)

// IsActive returns the current state of the Client.
func (c *Client) IsActive() bool {
	c.lock.RLock()
	var state = c.active
	c.lock.RUnlock()
	return state
} // func (c *Client) IsActive() bool

// Start creates a new goroutine running the Client's fetch loop.
func (c *Client) Start() {
	c.lock.Lock()
	c.active = true
	go c.Loop()
	c.lock.Unlock()
} // func (c *Client) Start()

// Stop terminates the Client's fetch loop.
func (c *Client) Stop() {
	c.lock.Lock()
	c.active = false
	c.lock.Unlock()
	c.stopQueue <- 1
} // func (c *Client) Stop()

// Refresh forces the Client to fetch the latest warnings from the DWD.
func (c *Client) Refresh() {
	c.refreshQueue <- 1
} // func (c *Client) Refresh()

// Loop is the Client's loop, intended to be run in a separate goroutine.
// It regularly checks the DWD's warnings and feeds them to the WarnQueue
func (c *Client) Loop() {
	var ticker = time.NewTicker(checkInterval)

	defer ticker.Stop()
	defer close(c.WarnQueue)

	for c.IsActive() {
		select {
		case <-ticker.C:
			c.log.Println("[DEBUG] Check warnings")
			c.checkWarnings()
		case <-c.refreshQueue:
			c.log.Println("[DEBUG] User requested refresh")
			c.checkWarnings()
		case <-c.stopQueue:
			return
		}
	}
} // func (c *Client) Loop()

func (c *Client) checkWarnings() {
	var (
		err  error
		raw  []byte
		info WeatherInfo
	)

	if raw, err = c.FetchWarning(); err != nil {
		c.log.Printf("[ERROR] Cannot fetch warnings from DWD: %s\n",
			err.Error())
		return
	} else if err = json.Unmarshal(raw, &info); err != nil {
		c.log.Printf("[ERROR] Cannot parse JSON: %s\n",
			err.Error())
		return
	} else if info.Time <= c.lastStamp {
		c.log.Printf("[DEBUG] Data at %s has already been processed.\n",
			info.TimeStamp().Format(common.TimestampFormatMinute))
		return
	}

	c.lastStamp = info.Time

	c.log.Printf("[DEBUG] Process %d Warnings\n", len(info.Warnings))

	for id, i := range info.Warnings {
	W_ITEM:
		for _, w := range i {
			for _, l := range c.locations {
				if l.MatchString(w.Location) {
					w.ID = id
					c.WarnQueue <- w
					continue W_ITEM
				}
			}
		}
	}

	c.log.Printf("[DEBUG] Process %d preliminary Warnings\n", len(info.PrelimWarnings))

	for id, i := range info.PrelimWarnings {
	V_ITEM:
		for _, w := range i {
			for _, l := range c.locations {
				if l.MatchString(w.Location) {
					w.ID = id
					c.WarnQueue <- w
					continue V_ITEM
				}
			}
		}
	}
} // func (c *Client) checkWarnings()
