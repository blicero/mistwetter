// /home/krylon/go/src/github.com/blicero/mistwetter/client/data.go
// -*- mode: go; coding: utf-8; -*-
// Created on 09. 05. 2023 by Benjamin Walkenhorst
// (c) 2023 Benjamin Walkenhorst
// Time-stamp: <2023-05-09 09:53:12 krylon>

package client

import (
	"fmt"
	"time"
)

// Warning represents a weather warning for a specific location and time.
type Warning struct {
	ID            int64
	Location      string `json:"regionName"`
	Start         int64  `json:"start"`
	End           int64  `json:"end"`
	Type          int64  `json:"type"`
	State         string `json:"state"`
	Level         int    `json:"level"`
	Description   string `json:"description"`
	Event         string `json:"event"`
	Headline      string `json:"headline"`
	Instruction   string `json:"instruction"`
	StateShort    string `json:"stateShort"`
	AltitudeStart int64  `json:"altitudeStart"`
	AltitudeEnd   int64  `json:"altitudeEnd"`
	t1            time.Time
	t2            time.Time
	uid           string
}

// TimeStart returns the Warning's Start time.
func (w *Warning) TimeStart() time.Time {
	if w.t1.IsZero() {
		w.t1 = time.Unix(w.Start/1000, 0)
	}

	return w.t1
} // func (w *Warning) TimeStart() time.Time

// TimeEnd returns the Warning's End time.
func (w *Warning) TimeEnd() time.Time {
	if w.t2.IsZero() {
		w.t2 = time.Unix(w.End/1000, 0)
	}

	return w.t2
} // func (w *Warning) TimeEnd() time.Time

// Period returns the timespan the warnings is issued for, as a 2-element array.
// Index 0 is the starting time, index 1 the end.
func (w *Warning) Period() [2]time.Time {
	return [2]time.Time{
		w.TimeStart(),
		w.TimeEnd(),
	}
} // func (w *Warning) Period() [2]time.Time

// GetUniqueID returns a string value used for comparing the Warning
// to determine if it is unique.
func (w *Warning) GetUniqueID() string {
	if w.uid == "" {
		w.uid = fmt.Sprintf("%s/%s",
			w.Location,
			w.Event)
	}

	return w.uid
} // func (w *Warning) GetUniqueID() string

// WarningList is a helper type used for sorting Warnings.
type WarningList []Warning

func (wl WarningList) Len() int      { return len(wl) }
func (wl WarningList) Swap(i, j int) { wl[i], wl[j] = wl[j], wl[i] }
func (wl WarningList) Less(i, j int) bool {
	var w1, w2 *Warning
	w1 = &wl[i]
	w2 = &wl[j]

	if w1.Location == w2.Location {
		return w1.Start < w2.Start
	}

	return w1.Location < w2.Location
} // func (wl WarningList) Less(i, j int) bool

// WeatherInfo represetns an aggregate of warnings issued by the DWD at a given time.
type WeatherInfo struct {
	Time           int64               `json:"time"`
	Warnings       map[int64][]Warning `json:"warnings"`
	PrelimWarnings map[int64][]Warning `json:"vorabInformation"`
	Copyright      string              `json:"copyright"`
}

// TimeStamp returns the time the warnings were last updated.
func (w *WeatherInfo) TimeStamp() time.Time {
	return time.Unix(w.Time/1000, 0)
} // func (w *WeatherInfo) TimeStamp() time.Time
