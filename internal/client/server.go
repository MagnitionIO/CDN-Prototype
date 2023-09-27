package client

import (
	"cdn-prototype/internal/origin"
	"context"
	"encoding/csv"
	"errors"
	"io"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/labstack/echo"
	"github.com/panjf2000/ants/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	"golang.org/x/time/rate"
)

type Server struct {
	Addr        string // for prometheus
	OriginAddrs []string
	Logger      *zerolog.Logger
	WikiFile    string
	CPUs        int // 0 => serial execution; positive => multiple goroutinues

	client      *origin.Client
	metrics     *metrics
	promHandler http.Handler
	ec          *echo.Echo
}

func (s *Server) Serve() error {
	if len(s.Addr) == 0 {
		return errors.New("missing addr value")
	}

	if s.CPUs < 0 {
		return errors.New("invalid CPUs value")
	}

	s.metrics = newMetrics()
	prometheus.MustRegister(s.metrics.requests)
	prometheus.MustRegister(s.metrics.latency)

	logFile, err := os.OpenFile("/var/log/cdn_client.log", os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return err // or however you want to handle this error
	}
	defer logFile.Close()

	s.ec = echo.New()
	if s.Logger == nil {
		log := zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}).
			With().Timestamp().Logger()
		s.Logger = &log
	}

	log := zerolog.New(logFile).With().Timestamp().Logger()
	s.Logger = &log

	s.setHandlers()

	s.client = &origin.Client{
		// Endpoint: "http://" + s.getOriginAddr(),
		Client: &http.Client{
			Transport: &http.Transport{
				MaxIdleConns:    s.CPUs,
				MaxConnsPerHost: s.CPUs,
			},
		},
	}

	go s.loadWikiTrace()

	s.Logger.Info().
		Str("addr", s.Addr).
		Msg("Start client server")

	return s.ec.Start(s.Addr)
}

func (s *Server) setHandlers() {
	s.promHandler = promhttp.Handler()
	s.ec.GET("/client/metrics", func(ctx echo.Context) error {
		s.promHandler.ServeHTTP(ctx.Response(), ctx.Request())
		return nil
	})
}

func (s *Server) loadWikiTrace() {
	t := time.Now()

	ctx := context.TODO()
	log := s.Logger.With().Str("wiki", s.WikiFile).Logger()

	// Initialize rate limiter
	limiter := rate.NewLimiter(rate.Every(time.Second), 1) // 1 requests per second

	if len(s.WikiFile) == 0 {
		log.Info().Msg("Skip loading wiki trace: empty file name")
		return
	}

	file, err := os.Open(s.WikiFile)
	if err != nil {
		log.Error().Err(err).Msg("Fail to open wiki file")
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Comma = ' '

	pool, err := ants.NewPool(2)
	if err != nil {
		log.Error().Err(err).Msg("Fail to create pool")
	}
	defer pool.Release()

	var wg sync.WaitGroup
	cnt := 0
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}

		if err != nil {
			log.Error().Err(err).Msg("Fail to read")
			return
		}

		seq, id, size, err := s.parse(record)
		if err != nil {
			log.Warn().Err(err).Strs("record", record).Msg("Skip invalid record")
			continue
		}

		// Wait for rate limiter
		limiter.Wait(ctx)

		wg.Add(1)
		ants.Submit(func() {
			defer wg.Done()
			s.getObject(ctx, seq, id, size)
		})
		cnt++
	}

	wg.Wait()

	log.Info().
		Int("count", cnt).
		Str("used", time.Since(t).String()).
		Msg("Get all objects")
}

func (s *Server) parse(record []string) (seq, id, size int, err error) {
	seq, err = strconv.Atoi(record[0])
	if err != nil {
		return
	}

	id, err = strconv.Atoi(record[1])
	if err != nil {
		return
	}

	size, err = strconv.Atoi(record[2])
	if err != nil {
		return
	}

	return
}

func (s *Server) getObject(ctx context.Context, seq, id, size int) {
	start := time.Now() // Record the start time here

	endpoint := "http://" + s.getOriginAddr()

	log := s.Logger.With().Int("seq", seq).
		Int("id", id).
		Int("size", size).
		Str("endpoint", endpoint).
		Logger()

	resp, err := s.client.GetObject(ctx, id, size, endpoint)
	// log.Debug().Str("resp", fmt.Sprintf("%+v", resp.StringResponse.Response)).Msg("Get object")

	if err != nil {
		log.Err(err).Msg("Fail to get object")
		return
	}
	defer resp.Body.Close()

	latency := time.Since(start) // Calculate the total latency here

	xcache := resp.Header.Get("X-Cache")

	if len(xcache) == 0 {
		xcache = "none"
	}
	s.metrics.requests.WithLabelValues(xcache).Inc()
	s.metrics.latency.Observe(float64(latency.Milliseconds()))

	log.Debug().Str("status", xcache).Dur("latency", latency).Msg("Get object")
}

var originIndex = 0
var originMutex = sync.Mutex{}

func (s *Server) getOriginAddr() string {
	// Load balancing: round-robin
	originMutex.Lock()
	addr := s.OriginAddrs[originIndex%len(s.OriginAddrs)]
	originIndex++
	originMutex.Unlock()
	return addr
}
