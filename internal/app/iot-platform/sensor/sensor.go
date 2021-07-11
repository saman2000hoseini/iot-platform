package sensor

import (
	"github.com/go-resty/resty/v2"
	"github.com/saman2000hoseini/iot-platform/internal/app/iot-platform/config"
	"github.com/saman2000hoseini/iot-platform/internal/app/iot-platform/model"
	"github.com/saman2000hoseini/iot-platform/internal/app/iot-platform/request"
	"github.com/sirupsen/logrus"
	"math/rand"
	"net/http"
	"time"
)

type Sensor struct {
	SensorInfo  *model.Node
	restyClient *resty.Client
}

func NewSensor(node *model.Node, sensor config.Sensor) *Sensor {
	return &Sensor{
		SensorInfo:  node,
		restyClient: resty.New().SetHostURL(sensor.LocalServerAddress).SetTimeout(sensor.ReadTimeout),
	}
}

func (s *Sensor) Sense() {
	for {
		value := rand.Intn(51)

		if time.Now().Local().Hour() >= 6 || time.Now().Local().Hour() <= 18 {
			value += 50
		}

		req := request.SensorData{
			SensorID:  s.SensorInfo.ID,
			IP:        s.SensorInfo.IP,
			EntryCode: s.SensorInfo.EntryCode,
			Type:      s.SensorInfo.Type,
			Value:     value,
		}

		url := "/nodes/data"

		resp, err := s.restyClient.R().SetBody(req).Post(url)
		if err != nil {
			logrus.Errorf("local host: failed to send request: %s", err.Error())
		}

		if resp.StatusCode() != http.StatusOK {
			logrus.Errorf("local server: request failed with status = %d and body = (%s)",
				resp.StatusCode(), resp.String())
		}

		logrus.Infof("Value: %d, Time: %s", value, time.Now().Local())

		<-time.Tick(30 * time.Second)
	}
}
