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
	Addr   string
	Port   uint16
	Logger *zerolog.Logger

	ec *echo.Echo
}

func (s *Server) Serve() error {
	if s.Port == 0 {
		return errors.New("missing port value")
	}

	s.ec = echo.New()
	if s.Logger == nil {
		log := zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}).
			With().Timestamp().Logger()
		s.Logger = &log
	}
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

	objSize := 0
	objId := ctx.Param("id")
	if size, err := strconv.Atoi(ctx.QueryParam("size")); err != nil {
		log.Debug().Err(err).Msg("Invalid size of object")
		return ctx.String(http.StatusBadRequest, "Usage: /origin/objects/:id?size=<integer>")
	} else {
		objSize = size
	}

	log.Debug().Str("objId", objId).Int("objSize", objSize).Msg("Try to get object")

	// Set Cache-Control header
	ctx.Response().Header().Set("Cache-Control", "public, max-age=3600")

	return ctx.String(http.StatusOK, strings.Repeat("*", objSize))
}
