package handler

import (
	"context"
	"github.com/labstack/echo/v4"
	"github.com/opentracing/opentracing-go"
	"github.com/saman2000hoseini/iot-platform/internal/app/iot-platform/model"
	"github.com/saman2000hoseini/iot-platform/internal/app/iot-platform/request"
	"github.com/saman2000hoseini/iot-platform/pkg/tracing"
	"github.com/sirupsen/logrus"
	"net/http"
	"strconv"
)

type SensorThresholdHandler struct {
	ThresholdRepo model.SQLSensorThresholdRepo
}

func NewSensorThresholdHandler(repo model.SQLSensorThresholdRepo) *SensorThresholdHandler {
	return &SensorThresholdHandler{
		ThresholdRepo: repo,
	}
}

func (h *SensorThresholdHandler) Submit(c echo.Context) error {
	tracer, closer := tracing.Init("local-server")
	defer closer.Close()
	opentracing.SetGlobalTracer(tracer)

	span := tracer.StartSpan("submit-threshold")
	defer span.Finish()

	ctx := opentracing.ContextWithSpan(context.Background(), span)

	req := new(request.SensorThreshold)

	reqType, err := strconv.ParseInt(c.FormValue("type"), 10, 32)
	if err != nil {
		return c.NoContent(http.StatusBadRequest)
	}

	reqThreshold, err := strconv.ParseInt(c.FormValue("threshold"), 10, 32)
	if err != nil {
		return c.NoContent(http.StatusBadRequest)
	}

	req.Type = int(reqType)
	req.Threshold = int(reqThreshold)

	if err := req.Validate(); err != nil {
		logrus.Infof("threshold: failed to validate: %s", err.Error())
		return c.NoContent(http.StatusBadRequest)
	}

	if err = h.ThresholdRepo.Save(model.NewSensorThreshold(req.Threshold, req.Type), ctx); err != nil {
		return c.NoContent(http.StatusInternalServerError)
	}

	return c.NoContent(http.StatusOK)
}
