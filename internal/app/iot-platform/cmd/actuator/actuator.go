package actuator

import (
	"github.com/google/uuid"
	"github.com/labstack/gommon/random"
	"github.com/saman2000hoseini/iot-platform/internal/app/iot-platform/actuator"
	"github.com/saman2000hoseini/iot-platform/internal/app/iot-platform/config"
	"github.com/saman2000hoseini/iot-platform/internal/app/iot-platform/model"
	"github.com/saman2000hoseini/iot-platform/internal/app/iot-platform/router"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
	"os/signal"
	"syscall"
)

const randLen = 36

func main(cfg config.Config) {
	actuatorConfig := cfg.LightBulb
	var id string

	u, err := uuid.NewRandom()
	if err != nil || u.String() == "" {
		id = random.String(randLen, random.Alphanumeric)
	} else {
		id = u.String()
	}

	myActuator := actuator.NewActuator(model.NewNode(id, "", "secret", actuatorConfig.Type))
	e := router.New(config.Server{ReadTimeout: actuatorConfig.ReadTimeout, WriteTimeout: actuatorConfig.WriteTimeout})

	e.POST("", myActuator.SetState)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	go myActuator.Act()

	go func() {
		if err := e.Start(actuatorConfig.Address); err != nil {
			logrus.Fatalf("failed to start iot platform actuator: %s", err.Error())
		}
	}()

	logrus.Infof("iot platform sensor with id %s started!", myActuator.ActuatorInfo.ID)

	s := <-sig

	logrus.Infof("signal %s received", s)
}

// Register registers central-server command for iot-platform binary.
func Register(root *cobra.Command, cfg config.Config) {
	runServer := &cobra.Command{
		Use:   "actuator",
		Short: "actuator node",
		Run: func(cmd *cobra.Command, args []string) {
			main(cfg)
		},
	}

	root.AddCommand(runServer)
}
