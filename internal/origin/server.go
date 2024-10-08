package origin

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
)

type Server struct {
	Addr     string
	Port     uint16
	Logger   *zerolog.Logger
	LogLevel zerolog.Level

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

	s.Logger.Info().
		Str("addr", s.Addr).Uint16("port", s.Port).
		Msg("Start origin server")

	return s.ec.Start(fmt.Sprintf("%s:%d", s.Addr, s.Port))
}

func (s *Server) setHandlers() {
	s.ec.GET("/origin", s.sayHello)
	s.ec.GET("/origin/objects/:id", s.getObject)
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
	ctx.Response().Header().Set("X-Cache-L2-Store", l2_store_header)

	return ctx.String(http.StatusOK, strings.Repeat("*", objSize))
}
