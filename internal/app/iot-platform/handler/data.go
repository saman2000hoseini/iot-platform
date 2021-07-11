package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/labstack/echo/v4"
	"github.com/saman2000hoseini/iot-platform/internal/app/iot-platform/model"
	"github.com/saman2000hoseini/iot-platform/internal/app/iot-platform/request"
	"github.com/saman2000hoseini/iot-platform/internal/pkg/nodestate"
	"github.com/saman2000hoseini/iot-platform/internal/pkg/nodetype"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

const (
	TOKEN                = "token"
	centralServerAddress = "http://localhost:65432"
)

type SensorDataHandler struct {
	DataRepo      model.SQLSensorDataRepo
	ThresholdRepo model.SQLSensorThresholdRepo
	restyClient   *resty.Client
}

type Authorization struct {
	Token string `json:"token"`
}

func NewSensorDataHandler(repo model.SQLSensorDataRepo, token string, thresholdRepo model.SQLSensorThresholdRepo) *SensorDataHandler {
	return &SensorDataHandler{
		DataRepo:      repo,
		ThresholdRepo: thresholdRepo,
		restyClient: resty.New().SetHostURL(centralServerAddress).
			SetCookie(&http.Cookie{Name: TOKEN, Value: token}).SetTimeout(2 * time.Minute),
	}
}

func (h *SensorDataHandler) Authorize(err error, c echo.Context) error {
	req := new(request.SensorData)

	var bodyBytes []byte
	if c.Request().Body != nil {
		bodyBytes, _ = ioutil.ReadAll(c.Request().Body)
	}

	c.Request().Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))

	if err := json.Unmarshal(bodyBytes, req); err != nil {
		logrus.Infof("node data authorize: failed to bind: %v, %s", c.Request(), err.Error())
		return c.NoContent(http.StatusBadRequest)
	}

	if err := req.Validate(); err != nil {
		logrus.Infof("node data authorize: failed to validate: %s", err.Error())
		return c.NoContent(http.StatusBadRequest)
	}

	url := "/user/node/auth"

	resp, err := h.restyClient.R().SetBody(req).Post(url)
	if err != nil {
		logrus.Errorf("failed sending request to central server: %s", err.Error())
		return c.NoContent(http.StatusInternalServerError)
	}

	if resp.StatusCode() != http.StatusOK {
		logrus.Errorf("authorization failed with status code: %d", resp.StatusCode())
		return c.NoContent(http.StatusUnauthorized)
	}

	authResp := &Authorization{}

	err = json.Unmarshal(resp.Body(), authResp)
	if err != nil {
		logrus.Infof("unauthorized: %s, %s, %s", req.SensorID, req.IP, req.EntryCode)
		return c.NoContent(http.StatusInternalServerError)
	}

	c.SetCookie(&http.Cookie{
		Name:     "token",
		Value:    authResp.Token,
		HttpOnly: true,
	})

	return h.Submit(c)
}

func (h *SensorDataHandler) Submit(c echo.Context) error {
	req := new(request.SensorData)

	if err := c.Bind(req); err != nil {
		logrus.Infof("node data: failed to bind: %s", err.Error())
		return c.NoContent(http.StatusBadRequest)
	}

	if err := req.Validate(); err != nil {
		logrus.Infof("node data: failed to validate: %s", err.Error())
		return c.NoContent(http.StatusBadRequest)
	}

	err := h.DataRepo.Save(model.NewSensorData(req.SensorID, req.Value, req.Type))
	if err != nil {
		return c.NoContent(http.StatusInternalServerError)
	}

	threshold, err := h.ThresholdRepo.Find(req.Type)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.NoContent(http.StatusOK)
		}

		return c.NoContent(http.StatusInternalServerError)
	}

	var state int

	if req.Value <= threshold.Threshold {
		state = nodestate.ON
	} else {
		state = nodestate.OFF
	}

	url := "/user/node/" + strconv.Itoa(nodetype.RelatedActuator(req.Type))

	resp, err := h.restyClient.R().Get(url)
	if err != nil {
		logrus.Errorf("local host: failed to send request: %s", err.Error())
		return c.NoContent(http.StatusInternalServerError)
	}

	if resp.StatusCode() != http.StatusOK {
		logrus.Errorf("local server: request failed with status = %d and body = (%s)",
			resp.StatusCode(), resp.String())
	}

	nodeResp := []model.Node{}
	err = json.Unmarshal(resp.Body(), &nodeResp)
	if err != nil {
		logrus.Errorf("local server: unmarshalling json response failed: %s", err.Error())
	}

	for _, node := range nodeResp {
		if node.State != state {
			restyClient := resty.New().SetHostURL(node.IP).SetTimeout(2 * time.Second)

			restyClient.R().SetFormData(map[string]string{"state": strconv.Itoa(state)}).Post("")

			update := request.NodeUpdate{ID: node.ID, State: state}
			h.restyClient.R().SetBody(update).Post("/user/node/update")
		}
	}

	return c.NoContent(http.StatusOK)
}

func (h *SensorDataHandler) Get(c echo.Context) error {
	t := c.Param("type")
	if len(t) == 0 || t == "" {
		return c.NoContent(http.StatusNotFound)
	}

	nodeType, err := strconv.ParseInt(t, 10, 32)
	if err != nil {
		return c.NoContent(http.StatusBadRequest)
	}

	if nodetype.IsSensor(int(nodeType)) {
		data, err := h.DataRepo.FindLast(int(nodeType))
		if err != nil {
			return c.NoContent(http.StatusNotFound)
		}

		return c.HTML(http.StatusOK, generateHTML(data.Value))
	}

	url := "/user/node/" + t

	resp, err := h.restyClient.R().Get(url)
	if err != nil {
		logrus.Errorf("local host: failed to send request: %s", err.Error())
		return c.NoContent(http.StatusInternalServerError)
	}

	if resp.StatusCode() != http.StatusOK {
		logrus.Errorf("local server: request failed with status = %d and body = (%s)",
			resp.StatusCode(), resp.String())
		return c.NoContent(http.StatusBadRequest)
	}

	nodeResp := []model.Node{}
	err = json.Unmarshal(resp.Body(), &nodeResp)
	if err != nil {
		logrus.Errorf("local server: unmarshalling json response failed: %s", err.Error())
		return c.NoContent(http.StatusInternalServerError)
	}

	return c.HTML(http.StatusOK, generateHTML(nodeResp[0].State))
}

func generateHTML(value interface{}) string {
	return fmt.Sprintf("<!DOCTYPE html><html lang='en'><head><meta charset='UTF-8'><title></title></head>"+
		"<body style='background: #7da7f3'>"+
		"<div style='height:100vh; display: flex; flex-direction: column; justify-content: center;align-items: center;align-content: center'>"+
		"<div style='font-size: 40px'>%v</div></div></body></html>", value)
}
