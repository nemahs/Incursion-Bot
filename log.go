package main

import (
	"log"
	"os"
)

var (
	Info *log.Logger
	Warning *log.Logger
	Error *log.Logger
)

func init() {
	Info = log.New(os.Stdout, "INFO: ", log.LstdFlags|log.Lshortfile)
	Warning = log.New(os.Stdout, "WARN: ", log.LstdFlags|log.Lshortfile)
	Error = log.New(os.Stdout, "ERROR: ", log.LstdFlags|log.Lshortfile)
}