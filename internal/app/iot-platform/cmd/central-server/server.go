package central_server

import (
	"context"
	"github.com/labstack/echo/v4"
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
	myDB, err := database.FirstSetup("./centralDB")
	if err != nil {
		logrus.Fatalf("failed to setup db: %s", err.Error())
	}

	err = myDB.AutoMigrate(&model.User{}, &model.Node{})
	if err != nil {
		logrus.Fatalf("error on creating tables: %s", err.Error())
	}

	userRepo := model.SQLUserRepo{DB: myDB}
	nodeRepo := model.SQLNodeRepo{DB: myDB}
	userHandler := handler.NewUserHandler(cfg, userRepo)
	nodeHandler := handler.NewNodeHandler(nodeRepo)

	e := router.New(cfg.CentralServer)

	user := e.Group("/user")
	user.POST("/register", userHandler.Register)
	user.File("/register", "./web/static/register.html")
	user.POST("/login", userHandler.Login)
	user.File("/login", "./web/static/login.html")

	node := user.Group("/node", middleware.JWTWithConfig(
		middleware.JWTConfig{
			ErrorHandlerWithContext: func(err error, e echo.Context) error {
				return e.File("./web/static/login.html")
			},
			SigningKey:  []byte(cfg.JWT.Secret),
			TokenLookup: "cookie:token",
		}))

	node.POST("", nodeHandler.Register)
	node.File("", "./web/static/node.html")

	node.GET("/:type", nodeHandler.GetNodes)
	node.POST("/auth", nodeHandler.Authorize)
	node.POST("/update", nodeHandler.UpdateNodes)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	go func() {
		if err := e.Start(cfg.CentralServer.Address); err != nil {
			logrus.Fatalf("failed to start iot platform central server: %s", err.Error())
		}
	}()

	logrus.Info("iot platform central server started!")

	s := <-sig

	logrus.Infof("signal %s received", s)

	ctx, cancel := context.WithTimeout(context.Background(), cfg.CentralServer.GracefulTimeout)
	defer cancel()

	e.Server.SetKeepAlivesEnabled(false)

	if err := e.Shutdown(ctx); err != nil {
		logrus.Errorf("failed to shutdown iot platform central central-server: %s", err.Error())
	}
}

// Register registers central-server command for iot-platform binary.
func Register(root *cobra.Command, cfg config.Config) {
	runServer := &cobra.Command{
		Use:   "central-server",
		Short: "central server for iot platform",
		Run: func(cmd *cobra.Command, args []string) {
			main(cfg)
		},
	}

	root.AddCommand(runServer)
}
