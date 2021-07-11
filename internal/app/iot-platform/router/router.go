package router

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/saman2000hoseini/iot-platform/internal/app/iot-platform/config"
	"github.com/sirupsen/logrus"
)

// New creates a new application router.
func New(cfg config.Server) *echo.Echo {
	e := echo.New()

	debug := logrus.IsLevelEnabled(logrus.DebugLevel)

	e.Debug = debug

	e.HideBanner = true

	if !debug {
		e.HidePort = true
	}

	e.Server.ReadTimeout = cfg.ReadTimeout
	e.Server.WriteTimeout = cfg.WriteTimeout

	recoverConfig := middleware.DefaultRecoverConfig
	recoverConfig.DisablePrintStack = !debug
	e.Use(middleware.RecoverWithConfig(recoverConfig))

	e.Use(middleware.CORS())

	return e
}
