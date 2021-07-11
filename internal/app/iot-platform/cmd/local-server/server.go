package local_server

import (
	"context"
	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo/v4/middleware"
	"github.com/saman2000hoseini/iot-platform/internal/app/iot-platform/config"
	"github.com/saman2000hoseini/iot-platform/internal/app/iot-platform/handler"
	"github.com/saman2000hoseini/iot-platform/internal/app/iot-platform/model"
	"github.com/saman2000hoseini/iot-platform/internal/app/iot-platform/router"
	"github.com/saman2000hoseini/iot-platform/internal/pkg/database"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
	"os/signal"
	"syscall"
)

func main(cfg config.Config) {
	myDB, err := database.FirstSetup("./localDB")
	if err != nil {
		logrus.Fatalf("failed to setup db: %s", err.Error())
	}

	err = myDB.AutoMigrate(&model.SensorData{}, &model.SensorThreshold{})
	if err != nil {
		logrus.Fatalf("error on creating tables: %s", err.Error())
	}

	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["username"] = "admin"
	claims["exp"] = 0
	t, err := token.SignedString([]byte(cfg.JWT.Secret))
	if err != nil {
		logrus.Fatalf("failed to create token: %s", err.Error())
	}

	thresholdRepo := model.SQLSensorThresholdRepo{DB: myDB}
	thresholdHandler := handler.NewSensorThresholdHandler(thresholdRepo)
	dataRepo := model.SQLSensorDataRepo{DB: myDB}
	dataHandler := handler.NewSensorDataHandler(dataRepo, t, thresholdRepo)

	e := router.New(cfg.LocalServer)

	e.File("", "./web/static/index.html")

	e.POST("/nodes/threshold", thresholdHandler.Submit)
	e.File("/nodes/threshold", "./web/static/threshold.html")
	e.POST("/nodes/data", dataHandler.Submit, middleware.JWTWithConfig(
		middleware.JWTConfig{
			ErrorHandlerWithContext: dataHandler.Authorize,
			SigningKey:              []byte(cfg.JWT.Secret),
			TokenLookup:             "cookie:token",
		}))
	e.GET("/nodes/:type", dataHandler.Get)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	go func() {
		if err := e.Start(cfg.LocalServer.Address); err != nil {
			logrus.Fatalf("failed to start iot platform local server: %s", err.Error())
		}
	}()

	logrus.Info("iot platform local server started!")

	s := <-sig

	logrus.Infof("signal %s received", s)

	ctx, cancel := context.WithTimeout(context.Background(), cfg.LocalServer.GracefulTimeout)
	defer cancel()

	e.Server.SetKeepAlivesEnabled(false)

	if err := e.Shutdown(ctx); err != nil {
		logrus.Errorf("failed to shutdown iot platform local server: %s", err.Error())
	}
}

// Register registers central-server command for iot-platform binary.
func Register(root *cobra.Command, cfg config.Config) {
	runServer := &cobra.Command{
		Use:   "local-server",
		Short: "local server for iot platform",
		Run: func(cmd *cobra.Command, args []string) {
			main(cfg)
		},
	}

	root.AddCommand(runServer)
}
