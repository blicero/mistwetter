// /home/krylon/go/src/github.com/blicero/mistwetter/gui/gui.go
// -*- mode: go; coding: utf-8; -*-
// Created on 09. 05. 2023 by Benjamin Walkenhorst
// (c) 2023 Benjamin Walkenhorst
// Time-stamp: <2023-05-09 18:20:53 krylon>

// Package gui provides the graphical user interface.
package gui

import (
	"log"

	"github.com/blicero/mistwetter/client"
)

const ()

type GUI struct {
	log *log.Logger
	c   *client.Client
}
