// /home/krylon/go/src/github.com/blicero/dwd/data/02_json_test.go
// -*- mode: go; coding: utf-8; -*-
// Created on 24. 07. 2021 by Benjamin Walkenhorst
// (c) 2021 Benjamin Walkenhorst
// Time-stamp: <2023-05-09 09:52:34 krylon>

package client

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"testing"
)

func TestParseJSON(t *testing.T) {
	type parseTestCase struct {
		path            string
		expectError     bool
		expectLocations []string
	}

	var cases = []parseTestCase{
		parseTestCase{
			path: "testdata/warnings.1.json",
			expectLocations: []string{
				"Starnberger See",
				"Kreis Trier-Saarburg und Stadt Trier",
			},
		},
		parseTestCase{
			path: "testdata/warnings.2.json",
			expectLocations: []string{
				"Schwarzwald-Baar-Kreis",
				"Kreis Trier-Saarburg und Stadt Trier",
				"Kreis Garmisch-Partenkirchen",
			},
		},
		parseTestCase{
			path: "testdata/warnings.3.json",
			expectLocations: []string{
				"Kreis Ludwigsburg",
				"Kreis Vechta",
				"Kreis und Stadt Heilbronn",
				"Elbe von Hamburg bis Cuxhaven",
				"Helgoland",
				"Kreis Berchtesgadener Land",
			},
		},
		parseTestCase{
			path: "testdata/warnings.4.json",
			expectLocations: []string{
				"Kreis und Stadt Passau",
				"Altmarkkreis Salzwedel",
				"Kreis Rostock - Küste",
				"Kreis Altötting",
			},
		},
	}

	for _, c := range cases {
		var (
			err  error
			fh   *os.File
			buf  bytes.Buffer
			info WeatherInfo
		)

		if fh, err = os.Open(c.path); err != nil {
			t.Errorf("Error opening %s: %s",
				c.path,
				err.Error())
			continue
		}

		defer fh.Close() // nolint: errcheck

		if _, err = io.Copy(&buf, fh); err != nil {
			t.Errorf("Error reading %s: %s",
				c.path,
				err.Error())
			continue
		} else if err = json.Unmarshal(buf.Bytes(), &info); err != nil {
			if !c.expectError {
				t.Errorf("Error parsing %s: %s",
					c.path,
					err.Error())
			}
			continue
		}

		var locations = make(map[string]int, len(c.expectLocations))
		for _, l := range c.expectLocations {
			locations[l] = 0
		}

		for _, l := range info.Warnings {
			for _, i := range l {
				if _, ok := locations[i.Location]; ok {
					locations[i.Location]++
				}
			}
		}

		for _, l := range info.PrelimWarnings {
			for _, i := range l {
				if _, ok := locations[i.Location]; ok {
					locations[i.Location]++
				}
			}
		}

		for l, cnt := range locations {
			if cnt == 0 {
				t.Errorf("No hits for Location %s",
					l)
			}
		}
	}

} // func TestParseJSON(t *testing.T)
