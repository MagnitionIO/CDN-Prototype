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
	"strings"
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

type CacheStats struct {
	lock        sync.Mutex // lock to guard stats variables
	References  uint64
	Misses      uint64
	ByteRefs    uint64
	ByteMisses  uint64
	Latency     uint64
	MissLatency uint64
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
	IOrefs   int
	IOps     int64
	LogLevel zerolog.Level

	client      *origin.Client
	metrics     *metrics
	promHandler http.Handler
	ec          *echo.Echo

	overallStats *CacheStats
	L1Stats      map[int]*CacheStats
	L2Stats      map[int]*CacheStats
}

func (s *Server) Serve() error {
	zerolog.TimestampFunc = func() time.Time {
		return time.Now().In(time.Local)
	}

	if len(s.Addr) == 0 {
		return errors.New("missing addr value")
	}

	if s.CPUs < 0 {
		return errors.New("invalid CPUs value")
	}

	if s.IOps == 0 {
		s.IOps = 1
	}

	s.metrics = newMetrics()
	prometheus.MustRegister(s.metrics.requests)
	prometheus.MustRegister(s.metrics.latency)
	s.L1Stats = make(map[int]*CacheStats)
	s.L2Stats = make(map[int]*CacheStats)

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

	log := zerolog.New(logFile).
		Level(s.LogLevel).
		With().
		Timestamp().
		Logger()
	s.Logger = &log

	s.Logger.Info().
		Str("Stats-Addr:", s.Addr).
		Str("L1-LB:", LBStrings[s.L1LB]).
		Str("L2-LB:", LBStrings[s.L2LB]).
		Msg("Client Parameters")

	s.overallStats = &CacheStats{}
	for i, _ := range s.L1Addrs {
		s.L1Stats[i] = &CacheStats{}
	}

	for i, _ := range s.L2Addrs {
		s.L2Stats[i] = &CacheStats{}
	}

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
	start_timer := time.Now()

	ctx := context.TODO()
	log := s.Logger.With().Str("wiki", s.WikiFile).Logger()

	// Initialize rate limiter
	limiter := rate.NewLimiter(rate.Every(time.Second/time.Duration(s.IOps)), 1) // 1 requests per second

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

		if (s.IOrefs > 0) && (cnt >= s.IOrefs) {
			break
		}

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

		if time.Since(start_timer) >= time.Duration(900)*time.Second {
			log.Info().
				Int("#################### Stats Peek", cnt).
				Msg(" Refs #########################")
			s.showStats()
			log.Info().
				Msg("######################################")
			// Reset the timer
			start_timer = time.Now()
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

	s.showStats()
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
	var nextL1Index = 0
	var hash32 = murmur3.Sum32([]byte(id))
	nextL2Index := int(hash32 % uint32(len(s.L2Addrs)))

	switch s.L1LB {
	case Rand:
		nextL1Index = rand.Intn(len(s.L1Addrs))
	case Hash:
		nextL1Index = int(hash32 % uint32(len(s.L1Addrs)))
	default:
	}

	addr := s.L1Addrs[nextL1Index]

	endpoint := "http://" + addr

	headers := map[string]string{
		"X-Cache-L1-Store":  "True",
		"X-Cache-L2-Store":  "True",
		"X-Cache-L2-Server": strconv.FormatUint(uint64(nextL2Index), 10),
	}

	s.Logger.Debug().
		Int("seq", seq).
		Str("id", id).
		Int("size", size).
		Int("L1:", nextL1Index).
		Int("L2:", nextL2Index).
		Str("endpoint", endpoint).
		Msg("Sending Request")

	start := time.Now() // Record the start time here
	resp, err := s.client.GetObject(ctx, id, size, endpoint, headers)
	latency := time.Since(start) // Calculate the total latency here

	log := s.Logger.With().Int("seq", seq).
		Str("id", id).
		Int("size", size).
		Int("L1:", nextL1Index).
		Int("L2:", nextL2Index).
		Str("endpoint", endpoint).
		Logger()

	if err != nil {
		log.Err(err).Msg("Fail to get object")
		return
	}
	defer resp.Body.Close()

	xcache_status := resp.Header.Get("X-Cache-Status")
	xcache_node := resp.Header.Get("X-Cache-Node")

	if xcache_status != "HIT" && xcache_status != "MISS" {
		log.Error().Str("status", xcache_status).Msg("Unkown status received")
	}
	s.metrics.requests.WithLabelValues(xcache_status).Inc()
	s.metrics.latency.Observe(float64(latency.Milliseconds()))

	overallStats := s.overallStats
	overallStats.lock.Lock()
	overallStats.References += 1
	overallStats.ByteRefs += uint64(size)
	overallStats.Latency += uint64(latency)
	if xcache_status == "MISS" {
		overallStats.Misses += 1
		overallStats.ByteMisses += uint64(size)
		overallStats.MissLatency += uint64(latency)
	}
	overallStats.lock.Unlock()

	L1Stats := s.L1Stats[nextL1Index]
	L1Stats.lock.Lock()
	L1Stats.References += 1
	L1Stats.ByteRefs += uint64(size)
	L1Stats.Latency += uint64(latency)
	if xcache_status == "MISS" || strings.Contains(xcache_node, "L2") {
		L1Stats.Misses += 1
		L1Stats.ByteMisses += uint64(size)
		L1Stats.MissLatency += uint64(latency)
	}
	L1Stats.lock.Unlock()

	if xcache_status == "MISS" || strings.Contains(xcache_node, "L2") {
		L2Stats := s.L2Stats[nextL2Index]
		L2Stats.lock.Lock()
		L2Stats.References += 1
		L2Stats.ByteRefs += uint64(size)
		L2Stats.Latency += uint64(latency)
		if xcache_status == "MISS" {
			L2Stats.Misses += 1
			L2Stats.ByteMisses += uint64(size)
			L2Stats.MissLatency += uint64(latency)
		}
		L2Stats.lock.Unlock()
	}

	log.Debug().Str("status", xcache_status).Str("Node", xcache_node).Dur("latency", time.Duration(latency.Nanoseconds())).Msg("Received Response")
}

func (s *Server) showStats() {
	overall := s.overallStats

	overall.lock.Lock()
	s.Logger.Info().
		Uint64("Refs", overall.References).
		Uint64("ByteRefs", overall.ByteRefs).
		Uint64("Misses", overall.Misses).
		Uint64("ByteMisses", overall.ByteMisses).
		Uint64("Hits", overall.References-overall.Misses).
		Uint64("ByteHits", overall.ByteRefs-overall.ByteMisses).
		Float64("MissRatio", float64(overall.Misses)/float64(overall.References)).
		Float64("ByteMissRatio", float64(overall.ByteMisses)/float64(overall.ByteRefs)).
		Float64("AvgLatency", float64(overall.Latency)/float64(overall.References)).
		Float64("AvgMissLatency", float64(overall.MissLatency)/float64(overall.Misses)).
		Float64("AvgHitLatency", (float64(overall.Latency)-float64(overall.MissLatency))/(float64(overall.References)-float64(overall.Misses))).
		Msg("Overall Stats")
	overall.lock.Unlock()

	for i, stats := range s.L1Stats {
		stats.lock.Lock()
		s.Logger.Info().
			Uint64("Refs", stats.References).
			Uint64("ByteRefs", stats.ByteRefs).
			Uint64("Misses", stats.Misses).
			Uint64("ByteMisses", stats.ByteMisses).
			Uint64("Hits", stats.References-stats.Misses).
			Uint64("ByteHits", stats.ByteRefs-stats.ByteMisses).
			Float64("MissRatio", float64(stats.Misses)/float64(stats.References)).
			Float64("ByteMissRatio", float64(stats.ByteMisses)/float64(stats.ByteRefs)).
			Float64("AvgLatency", float64(stats.Latency)/float64(stats.References)).
			Float64("AvgMissLatency", float64(stats.MissLatency)/float64(stats.Misses)).
			Float64("AvgHitLatency", (float64(stats.Latency)-float64(stats.MissLatency))/(float64(stats.References)-float64(stats.Misses))).
			Msg("Stats Cache L1-" + string(rune(i)))
		stats.lock.Unlock()
	}

	for i, stats := range s.L2Stats {
		stats.lock.Lock()
		s.Logger.Info().
			Uint64("Refs", stats.References).
			Uint64("ByteRefs", stats.ByteRefs).
			Uint64("Misses", stats.Misses).
			Uint64("ByteMisses", stats.ByteMisses).
			Uint64("Hits", stats.References-stats.Misses).
			Uint64("ByteHits", stats.ByteRefs-stats.ByteMisses).
			Float64("MissRatio", float64(stats.Misses)/float64(stats.References)).
			Float64("ByteMissRatio", float64(stats.ByteMisses)/float64(stats.ByteRefs)).
			Float64("AvgLatency", float64(stats.Latency)/float64(stats.References)).
			Float64("AvgMissLatency", float64(stats.MissLatency)/float64(stats.Misses)).
			Float64("AvgHitLatency", (float64(stats.Latency)-float64(stats.MissLatency))/(float64(stats.References)-float64(stats.Misses))).
			Msg("Stats Cache L2-" + string(rune(i)))
		stats.lock.Unlock()
	}

}

// var originIndex = 0
// var originMutex = sync.Mutex{}

// func (s *Server) getOriginAddr(id string) string {
// 	// Load balancing: round-robin
// 	var nextL1Index = 0
// 	switch s.L1LB {
// 	case Rand:
// 		nextL1Index = rand.Int() % len(s.L1Addrs)
// 	case Hash:
// 		hash32 := murmur3.Sum32([]byte(id))
// 		nextL1Index = hash32 % len(s.L1Addrs)
// 	default:
// 	}
// 	//	originMutex.Lock()
// 	addr := s.L1Addrs[nextL1Index]
// 	//	originIndex++
// 	//	originMutex.Unlock()
// 	return addr
// }
