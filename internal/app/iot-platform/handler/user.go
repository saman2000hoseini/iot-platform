package handler

import (
	"github.com/saman2000hoseini/iot-platform/internal/app/iot-platform/config"
	"github.com/saman2000hoseini/iot-platform/internal/app/iot-platform/model"
	"github.com/saman2000hoseini/iot-platform/internal/app/iot-platform/request"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

type UserHandler struct {
	Cfg      config.Config
	UserRepo model.UserRepo
}

func NewUserHandler(cfg config.Config, userRepo model.UserRepo) *UserHandler {
	return &UserHandler{
		Cfg:      cfg,
		UserRepo: userRepo,
	}
}

func (h *UserHandler) Register(c echo.Context) error {
	req := new(request.UserRequest)

	req.Username = c.FormValue("username")
	req.Password = c.FormValue("password")

	if err := req.Validate(); err != nil {
		logrus.Infof("register: failed to validate: %s", err.Error())
		return c.NoContent(http.StatusBadRequest)
	}

	user := model.NewUser(req.Username, req.Password, req.Role)

	if err := h.UserRepo.Save(user); err != nil {
		logrus.Infof("register: failed to save: %s", err.Error())
		return c.NoContent(http.StatusBadRequest)
	}

	token, err := h.generateJWT(*user)
	if err != nil {
		logrus.Infof("register: failed to generate jwt: %s", err.Error())
		return c.NoContent(http.StatusInternalServerError)
	}

	c.SetCookie(&http.Cookie{
		Path:     "",
		Name:     "token",
		Value:    token,
		HttpOnly: true,
	})

	return c.File("./web/static/node.html")
}

func (h *UserHandler) Login(c echo.Context) error {
	req := new(request.UserRequest)

	req.Username = c.FormValue("username")
	req.Password = c.FormValue("password")

	if err := req.Validate(); err != nil {
		logrus.Infof("login: failed to validate: %s", err.Error())
		return c.NoContent(http.StatusBadRequest)
	}

	user, err := h.UserRepo.Find(req.Username)
	if err != nil {
		logrus.Infof("login: failed to find: %s", err.Error())
		return c.NoContent(http.StatusForbidden)
	}

	if !user.CheckPassword(req.Password) {
		logrus.Info("login: incorrect password")
		return c.NoContent(http.StatusForbidden)
	}

	token, err := h.generateJWT(user)
	if err != nil {
		logrus.Infof("login: failed to generate jwt: %s", err.Error())
		return c.NoContent(http.StatusInternalServerError)
	}

	c.SetCookie(&http.Cookie{
		Path:     "",
		Name:     "token",
		Value:    token,
		HttpOnly: true,
	})

	return c.File("./web/static/node.html")
}

func (h *UserHandler) generateJWT(user model.User) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["username"] = user.Username
	claims["role"] = user.Role
	claims["exp"] = time.Now().Add(h.Cfg.JWT.Expiration).Unix()

	return token.SignedString([]byte(h.Cfg.JWT.Secret))
}
