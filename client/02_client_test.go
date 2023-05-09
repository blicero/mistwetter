// /home/krylon/go/src/github.com/blicero/dwd/data/01_client_test.go
// -*- mode: go; coding: utf-8; -*-
// Created on 23. 07. 2021 by Benjamin Walkenhorst
// (c) 2021 Benjamin Walkenhorst
// Time-stamp: <2023-05-09 17:54:11 krylon>

package client

import (
	"bytes"
	"io"
	"os"
	"testing"
)

var c *Client

func TestClientCreate(t *testing.T) {
	var err error

	if c, err = New("Bielefeld", "Berchtesgaden", "Diepholz"); err != nil {
		c = nil
		t.Fatalf("Failed to create Client: %s",
			err.Error())
	}
} // func TestClientCreate(t *testing.T)

func TestClientFetch(t *testing.T) {
	if c == nil {
		t.SkipNow()
	}

	var (
		data []byte
		err  error
	)

	if data, err = c.FetchWarning(); err != nil {
		t.Errorf("Failed to fetch weather warnings data: %s", err.Error())
	} else if data == nil {
		t.Error("Client returned no error, but data is nil")
	}
} // func TestClientFetch(t *testing.T)

func TestClientProcess(t *testing.T) {
	if c == nil {
		t.SkipNow()
	}

	const path = "testdata/warnings.3.json"

	var (
		err error
		fh  *os.File
		buf bytes.Buffer
	)

	if fh, err = os.Open(path); err != nil {
		t.Fatalf("Cannot open %s for reading: %s",
			path,
			err.Error())
	}

	defer fh.Close()

	if _, err = io.Copy(&buf, fh); err != nil {
		t.Fatalf("Error reading %s: %s",
			path,
			err.Error())
	}

	var info []Warning

	if info, err = c.ProcessWarnings(buf.Bytes()); err != nil {
		t.Fatalf("Error processing response data from %s: %s",
			path,
			err.Error())
	} else if len(info) == 0 {
		t.Fatal("ProcessWarnings returned empty result list")
	}
} // func TestClientProcess(t *testing.T)
