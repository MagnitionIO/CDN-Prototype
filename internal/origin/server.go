package origin

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
)

type Server struct {
	Addr     string
	Port     uint16
	Logger   *zerolog.Logger
	LogLevel zerolog.Level
	lock     sync.Mutex
	l2_map   map[string]int

	ec *echo.Echo
}

func (s *Server) Serve() error {
	if s.Port == 0 {
		return errors.New("missing port value")
	}

	zerolog.TimestampFunc = func() time.Time {
		return time.Now().In(time.Local)
	}

	s.ec = echo.New()

	logFile, err := os.OpenFile("/var/log/origin.log", os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return err // or however you want to handle this error
	}
	defer logFile.Close()

	log := zerolog.New(logFile).
		Level(s.LogLevel).
		With().Timestamp().Logger()

	s.Logger = &log
	s.setHandlers()
	s.l2_map = make(map[string]int)

	s.Logger.Info().
		Str("addr", s.Addr).Uint16("port", s.Port).
		Msg("Start origin server")

	return s.ec.Start(fmt.Sprintf("%s:%d", s.Addr, s.Port))
}

func (s *Server) setHandlers() {
	s.ec.GET("/origin", s.sayHello)
	s.ec.GET("/origin/objects/:id", s.getObject)
	s.ec.GET("/origin/objects/enforce-l2/:id", s.enforceL2Object)
}

func (s *Server) sayHello(ctx echo.Context) error {
	return ctx.String(http.StatusOK, "Hello, I'm origin.")
}

func (s *Server) getObject(ctx echo.Context) error {
	log := s.Logger.With().Str("path", ctx.Path()).Logger()

	l1_store_header := ctx.Request().Header.Get("X-Cache-L1-Store")
	l2_store_header := ctx.Request().Header.Get("X-Cache-L2-Store")

	objSize := 0
	objId := ctx.Param("id")
	if size, err := strconv.Atoi(ctx.QueryParam("size")); err != nil {
		log.Debug().Err(err).Msg("Invalid size of object")
		return ctx.String(http.StatusBadRequest, "Usage: /origin/objects/:id?size=<integer>")
	} else {
		objSize = size
	}

	log.Debug().Str("objId", objId).Int("objSize", objSize).Str("X-Cache-L1-Store", l1_store_header).Str("X-Cache-L2-Store", l2_store_header).Msg("Try to get object")

	// Set Cache-Control header
	ctx.Response().Header().Set("Cache-Control", "public, max-age=172800")
	ctx.Response().Header().Set("X-Cache-Status", "MISS")
	ctx.Response().Header().Set("X-Cache-Node", "ORIGIN")
	ctx.Response().Header().Set("X-Cache-L1-Store", l1_store_header)
	s.lock.Lock()
	if _, found := s.l2_map[objId]; found {
		ctx.Response().Header().Set("X-Cache-L2-Store", "True")
		s.l2_map[objId]--
		if s.l2_map[objId] == 0 {
			delete(s.l2_map, objId)
		}
	} else {
		ctx.Response().Header().Set("X-Cache-L2-Store", l2_store_header)
	}
	s.lock.Unlock()

	return ctx.String(http.StatusOK, strings.Repeat("*", objSize))
}

func (s *Server) enforceL2Object(ctx echo.Context) error {
	log := s.Logger.With().Str("path", ctx.Path()).Logger()

	l1_store_header := ctx.Request().Header.Get("X-Cache-L1-Store")
	l2_store_header := ctx.Request().Header.Get("X-Cache-L2-Store")

	objSize := 0
	objId := ctx.Param("id")
	if size, err := strconv.Atoi(ctx.QueryParam("size")); err != nil {
		log.Debug().Err(err).Msg("Invalid size of object")
		return ctx.String(http.StatusBadRequest, "Usage: /origin/objects/enforce-l2/:id?size=<integer>")
	} else {
		objSize = size
	}

	log.Debug().Str("objId", objId).Int("objSize", objSize).Str("X-Cache-L1-Store", l1_store_header).Str("X-Cache-L2-Store", l2_store_header).Msg("Enforce L2 object Start")

	ctx.Response().Header().Set("X-Cache-L2-Store", "True")

	s.lock.Lock()
	if _, found := s.l2_map[objId]; found {
		s.l2_map[objId]++
	} else {
		s.l2_map[objId] = 1
	}
	s.lock.Unlock()

	log.Debug().Str("objId", objId).Int("objSize", objSize).Str("X-Cache-L1-Store", l1_store_header).Str("X-Cache-L2-Store", l2_store_header).Msg("Enforce L2 object End")

	return ctx.String(http.StatusOK, "Enforced L2 Object")
}
