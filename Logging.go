package main

import (
    "os"
    "github.com/op/go-logging"
)

// Logger setup
func SetupLogging(errorLevel logging.Level) {
    // Setup the format for logging
    var format = logging.MustStringFormatter(`%{color}[%{level:.4s}] %{time:01/02/2006 15:04:05} ▶ %{color:reset} %{message}`)

    // Configure the logging backend
    backend := logging.NewLogBackend(os.Stdout, "", 0)
    backendFormatter := logging.NewBackendFormatter(backend, format)
    backendLevel := logging.AddModuleLevel(backendFormatter)
    backendLevel.SetLevel(errorLevel, "")

    logging.SetBackend(backendLevel)
}

