package sensor

import (
	"github.com/google/uuid"
	"github.com/labstack/gommon/random"
	"github.com/saman2000hoseini/iot-platform/internal/app/iot-platform/config"
	"github.com/saman2000hoseini/iot-platform/internal/app/iot-platform/model"
	"github.com/saman2000hoseini/iot-platform/internal/app/iot-platform/sensor"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
	"os/signal"
	"syscall"
)

const randLen = 36

func main(cfg config.Config) {
	sensorConfig := cfg.Light
	var id string

	u, err := uuid.NewRandom()
	if err != nil || u.String() == "" {
		id = random.String(randLen, random.Alphanumeric)
	} else {
		id = u.String()
	}

	mySensor := sensor.NewSensor(model.NewNode(id, "", "secret", sensorConfig.Type), sensorConfig)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	go mySensor.Sense()

	logrus.Infof("iot platform sensor with id %s started!", mySensor.SensorInfo.ID)

	s := <-sig

	logrus.Infof("signal %s received", s)
}

// Register registers central-server command for iot-platform binary.
func Register(root *cobra.Command, cfg config.Config) {
	runServer := &cobra.Command{
		Use:   "sensor",
		Short: "sensor node",
		Run: func(cmd *cobra.Command, args []string) {
			main(cfg)
		},
	}

	root.AddCommand(runServer)
}
