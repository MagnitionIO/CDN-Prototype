package main

import (
	"cdn-prototype/internal/client"
	"fmt"
	"os"
	"runtime"
	"strings"
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
		ServerAddr string `long:"server-addr" description:"specifying a server address for prometheus" default:":9090"`
		L1Addrs    string `long:"l1-addrs" description:"specifying L1 addresses to get object" default:":8080"`
		L2Addrs    string `long:"l2-addrs" description:"specifying L2 addresses to get object" default:":8080"`
		L1LB       string `long:"l1-lb" description:"specifying L1 LB type, rand or hash"`
		TraceFile  string `long:"trace-file" description:"specifying a file name of wiki trace records"`
		LogLevel   string `long:"log-level" description:"specifying log level (info, debug, warn, error)" default:"info"`
		CPUs       int    `long:"cpus" description:"specify the number of CPUs to be used" default:"1"`
		Refs       int    `long:"io-refs" description:"specify the number of IO-Refs to run" default:"-1"`
		IOps       int64  `long:"iops" description:"specify the number of IO per second" default:"1"`
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

	L1Addrs := strings.Split(opts.L1Addrs, ",")
	L2Addrs := strings.Split(opts.L2Addrs, ",")
	var L1LB = client.StringToLB[opts.L1LB]
	var L2LB = client.StringToLB["hash"]

	s := &client.Server{
		Addr:     opts.ServerAddr,
		L1Addrs:  L1Addrs,
		L2Addrs:  L2Addrs,
		L1LB:     L1LB,
		L2LB:     L2LB,
		WikiFile: opts.TraceFile,
		CPUs:     opts.CPUs,
		Logger:   &logger,
		LogLevel: logLevel,
		IOrefs:   opts.Refs,
		IOps:     opts.IOps,
	}

	if err := s.Serve(); err != nil {
		logger.Err(err).Msg("Fail to serve")
	}
}
