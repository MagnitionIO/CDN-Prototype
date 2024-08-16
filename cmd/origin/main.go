package main

import (
	"cdn-prototype/internal/origin"
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/jessevdk/go-flags"
	"github.com/rs/zerolog"
)

func getLogLevel(level string) zerolog.Level {
	switch level {
	case "debug":
		return zerolog.DebugLevel
	case "info":
		return zerolog.InfoLevel
	case "warn":
		return zerolog.WarnLevel
	case "error":
		return zerolog.ErrorLevel
	case "fatal":
		return zerolog.FatalLevel
	case "panic":
		return zerolog.PanicLevel
	default:
		return zerolog.InfoLevel
	}
}

func main() {
	var opts struct {
		ServerAddr string `long:"server-addr" description:"specifying a server address"`
		ServerPort uint16 `long:"server-port" description:"specifying a server port" default:"8080"`
		LogLevel   string `long:"log-level" description:"specifying log level (info, debug, warn, error)" default:"info"`
		CPUs       int    `long:"cpus" description:"specify the number of CPUs to be used"`
	}

	_, err := flags.Parse(&opts)
	if err != nil {
		panic(fmt.Errorf("fail to parse flags: %s", err))
	}

	if opts.CPUs == 0 {
		runtime.GOMAXPROCS(runtime.NumCPU())
	} else {
		if runtime.NumCPU() < opts.CPUs {
			panic(fmt.Errorf("Only %d CPUs are available but %d is specified", runtime.NumCPU(), opts.CPUs))
		}
		runtime.GOMAXPROCS(opts.CPUs)
	}

	logLevel := getLogLevel(opts.LogLevel)
	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}).
		Level(logLevel).
		With().
		Timestamp().
		Logger()

	s := origin.Server{
		Addr:     opts.ServerAddr,
		Port:     opts.ServerPort,
		Logger:   &logger,
		LogLevel: logLevel,
	}

	if err := s.Serve(); err != nil {
		logger.Err(err).Msg("Fail to serve")
	}
}
