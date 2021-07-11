package handler

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo/v4"
	"github.com/saman2000hoseini/iot-platform/internal/app/iot-platform/config"
	"github.com/saman2000hoseini/iot-platform/internal/app/iot-platform/model"
	"github.com/saman2000hoseini/iot-platform/internal/app/iot-platform/request"
	"github.com/sirupsen/logrus"
	"net/http"
	"strconv"
	"time"
)

type NodeHandler struct {
	NodeRepo model.NodeRepo
	Cfg      config.Config
}

func NewNodeHandler(NodeRepo model.NodeRepo) *NodeHandler {
	return &NodeHandler{
		NodeRepo: NodeRepo,
	}
}

func (h *NodeHandler) Register(c echo.Context) error {
	req := new(request.NodeRequest)

	req.ID = c.FormValue("id")
	req.IP = c.FormValue("ip")
	req.EntryCode = c.FormValue("entrycode")
	reqType, err := strconv.ParseInt(c.FormValue("type"), 10, 32)
	if err != nil {
		logrus.Infof("register: failed to parse req type: %s", err.Error())
		return c.NoContent(http.StatusBadRequest)
	}

	req.Type = int(reqType)

	if err := req.Validate(); err != nil {
		logrus.Infof("register: failed to validate: %s", err.Error())
		return c.NoContent(http.StatusBadRequest)
	}

	Node := model.NewNode(req.ID, req.IP, req.EntryCode, req.Type)

	if err := h.NodeRepo.Save(Node); err != nil {
		logrus.Infof("register: failed to save: %s", err.Error())
		return c.NoContent(http.StatusBadRequest)
	}

	return c.NoContent(http.StatusOK)
}

func (h *NodeHandler) GetNodes(c echo.Context) error {
	t := c.Param("type")
	if len(t) == 0 || t == "" {
		logrus.Info("bad request with nil type")
		return c.NoContent(http.StatusBadRequest)
	}

	nodeType, err := strconv.ParseInt(t, 10, 32)
	if err != nil {
		logrus.Info("bad request with bad node type")
		return c.NoContent(http.StatusBadRequest)
	}

	data, err := h.NodeRepo.FindByType(int(nodeType))
	if err != nil {
		logrus.Errorf("error finding node: %s", err.Error())
		return c.NoContent(http.StatusNotFound)
	}

	return c.JSON(http.StatusOK, data)
}

func (h *NodeHandler) UpdateNodes(c echo.Context) error {
	req := new(request.NodeUpdate)

	if err := c.Bind(req); err != nil {
		logrus.Infof("node update: failed to bind: %s", err.Error())
		return c.NoContent(http.StatusBadRequest)
	}

	if err := req.Validate(); err != nil {
		logrus.Infof("node update: failed to validate: %s", err.Error())
		return c.NoContent(http.StatusBadRequest)
	}

	err := h.NodeRepo.Update(model.Node{ID: req.ID, State: req.State})
	if err != nil {
		logrus.Errorf("error updating node: %s", err.Error())
		return c.NoContent(http.StatusInternalServerError)
	}

	return c.NoContent(http.StatusOK)
}

func (h *NodeHandler) Authorize(c echo.Context) error {
	req := new(request.SensorData)

	if err := c.Bind(req); err != nil {
		logrus.Infof("node authorize: failed to bind: %s", err.Error())
		return c.NoContent(http.StatusBadRequest)
	}

	if err := req.Validate(); err != nil {
		logrus.Infof("node authorize: failed to validate: %s", err.Error())
		return c.NoContent(http.StatusBadRequest)
	}

	if !h.NodeRepo.IsValid(req.SensorID, req.IP, req.EntryCode) {
		logrus.Info("node authorize: node does not exist")
		return c.NoContent(http.StatusUnauthorized)
	}

	token, err := h.generateJWT(model.NewNode(req.SensorID, req.IP, req.EntryCode, req.Type))
	if err != nil {
		logrus.Infof("node authorize: failed generating jwt: %s", err.Error())
		return c.NoContent(http.StatusInternalServerError)
	}

	return c.JSON(http.StatusOK, map[string]string{
		TOKEN: token,
	})
}

func (h *NodeHandler) generateJWT(node *model.Node) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["id"] = node.ID
	claims["ip"] = node.IP
	claims["exp"] = time.Now().Add(h.Cfg.JWT.Expiration).Unix()

	return token.SignedString([]byte(h.Cfg.JWT.Secret))
}
