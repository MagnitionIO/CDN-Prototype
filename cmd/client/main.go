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

func main() {
	var opts struct {
		ServerAddr  string `long:"server-addr" description:"specifying a server address for prometheus" default:":9090"`
		OriginAddrs string `long:"origin-addrs" description:"specifying origin addresses to get object" default:":8080"`
		WikiFile    string `long:"wiki-file" description:"specifying a file name of wiki trace records"`
		LogLevel    string `long:"log-level" description:"specifying log level (info, debug, warn, error)" default:"info"`
		CPUs        int    `long:"cpus" description:"specify the number of CPUs to be used" default:"1"`
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

	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}).
		With().Timestamp().Logger()

	switch opts.LogLevel {
	case "debug":
		logger = logger.Level(zerolog.DebugLevel)
	case "error":
		logger = logger.Level(zerolog.ErrorLevel)
	case "warn":
		logger = logger.Level(zerolog.WarnLevel)
	default:
		logger = logger.Level(zerolog.InfoLevel)
	}

	originAddrs := strings.Split(opts.OriginAddrs, ",")

	s := &client.Server{
		Addr:        opts.ServerAddr,
		OriginAddrs: originAddrs,
		WikiFile:    opts.WikiFile,
		CPUs:        opts.CPUs,
		Logger:      &logger,
	}

	if err := s.Serve(); err != nil {
		logger.Err(err).Msg("Fail to serve")
	}
}
