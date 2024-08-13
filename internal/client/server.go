package client

import (
	"cdn-prototype/internal/origin"
	"context"
	"encoding/csv"
	"errors"
	"io"
	"math/rand"
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
	"github.com/spaolacci/murmur3"
	"golang.org/x/time/rate"
)

type LB_TYPE int

const (
	Unknown LB_TYPE = iota
	Hash
	Rand
)

var LBStrings = map[LB_TYPE]string{
	Hash: "hash",
	Rand: "rand",
}

var StringToLB = map[string]LB_TYPE{
	"hash": Hash,
	"rand": Rand,
}

type Server struct {
	Addr     string // for prometheus
	L1Addrs  []string
	L2Addrs  []string
	L1LB     LB_TYPE
	L2LB     LB_TYPE
	Logger   *zerolog.Logger
	WikiFile string
	CPUs     int // 0 => serial execution; positive => multiple goroutinues

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

	s.Logger.Info().
		Str("Stats-Addr:", s.Addr).
		Str("L1-LB:", LBStrings[s.L1LB]).
		Str("L2-LB:", LBStrings[s.L2LB]).
		Msg("Client Parameters")

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
	limiter := rate.NewLimiter(rate.Every(time.Second*5), 1) // 1 requests per second

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

func (s *Server) parse(record []string) (seq int, id string, size int, err error) {
	seq, err = strconv.Atoi(record[0])
	if err != nil {
		return
	}

	id = record[1]
	if err != nil {
		return
	}

	size, err = strconv.Atoi(record[2])
	if err != nil {
		return
	}

	return
}

func (s *Server) getObject(ctx context.Context, seq int, id string, size int) {
	start := time.Now() // Record the start time here

	var nextIndex = 0
	var hash32 = murmur3.Sum32([]byte(id))
	nextL2Index := int(hash32 % uint32(len(s.L2Addrs)))
	switch s.L1LB {
	case Rand:
		nextIndex = rand.Int() % len(s.L1Addrs)
	case Hash:
		nextIndex = int(hash32 % uint32(len(s.L1Addrs)))
	default:
	}
	//	originMutex.Lock()
	addr := s.L1Addrs[nextIndex]

	endpoint := "http://" + addr

	log := s.Logger.With().Int("seq", seq).
		Str("id", id).
		Int("size", size).
		Int("L1:", nextIndex).
		Int("L2:", nextL2Index).
		Str("endpoint", endpoint).
		Logger()

	headers := map[string]string{
		"X-Cache-L1-Store":  "False",
		"X-Cache-L2-Store":  "False",
		"X-Cache-L2-Server": strconv.FormatUint(uint64(nextL2Index), 10),
	}

	resp, err := s.client.GetObject(ctx, id, size, endpoint, headers)
	// log.Debug().Str("resp", fmt.Sprintf("%+v", resp.StringResponse.Response)).Msg("Get object")

	if err != nil {
		log.Err(err).Msg("Fail to get object")
		return
	}
	defer resp.Body.Close()

	latency := time.Since(start) // Calculate the total latency here

	xcache_status := resp.Header.Get("X-Cache-Status")
	xcache_node := resp.Header.Get("X-Cache-Node")

	if len(xcache_status) == 0 {
		xcache_status = "none"
	}
	s.metrics.requests.WithLabelValues(xcache_status).Inc()
	s.metrics.latency.Observe(float64(latency.Milliseconds()))

	log.Debug().Str("status", xcache_status).Str("Node", xcache_node).Dur("latency", latency).Msg("Get object")
}

// var originIndex = 0
// var originMutex = sync.Mutex{}

// func (s *Server) getOriginAddr(id string) string {
// 	// Load balancing: round-robin
// 	var nextIndex = 0
// 	switch s.L1LB {
// 	case Rand:
// 		nextIndex = rand.Int() % len(s.L1Addrs)
// 	case Hash:
// 		hash32 := murmur3.Sum32([]byte(id))
// 		nextIndex = hash32 % len(s.L1Addrs)
// 	default:
// 	}
// 	//	originMutex.Lock()
// 	addr := s.L1Addrs[nextIndex]
// 	//	originIndex++
// 	//	originMutex.Unlock()
// 	return addr
// }
