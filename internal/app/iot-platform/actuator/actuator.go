package actuator

import (
	"github.com/labstack/echo/v4"
	"github.com/saman2000hoseini/iot-platform/internal/app/iot-platform/model"
	"github.com/sirupsen/logrus"
	"net/http"
	"strconv"
	"time"
)

type Actuator struct {
	ActuatorInfo *model.Node
}

func NewActuator(node *model.Node) *Actuator {
	return &Actuator{
		ActuatorInfo: node,
	}
}

func (a *Actuator) Act() {
	for {
		logrus.Infof("State: %v, Time: %s", a.ActuatorInfo.State, time.Now().Local())

		<-time.Tick(30 * time.Second)
	}
}

func (a *Actuator) SetState(c echo.Context) error {
	state, err := strconv.ParseInt(c.FormValue("state"), 10, 32)
	if err != nil {
		logrus.Errorf("failed to update state: %s", err.Error())
		return c.NoContent(http.StatusBadRequest)
	}

	a.ActuatorInfo.State = int(state)
	return c.NoContent(http.StatusOK)
}
