package handler

import (
	"net/http"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	fb "github.com/huandu/facebook"
	"github.com/labstack/echo"
	"github.com/lavasov/gorest/model"
	"golang.org/x/oauth2"
)

func createJTWToken(name string, jwtSigningKey string) (string, error) {
	// Create token
	token := jwt.New(jwt.SigningMethodHS256)

	// Set claims
	claims := token.Claims.(jwt.MapClaims)
	claims["name"] = name
	claims["admin"] = true
	claims["exp"] = time.Now().Add(time.Hour * 72).Unix()

	// Generate encoded token and send it as response.
	t, err := token.SignedString([]byte(jwtSigningKey))
	if err != nil {
		return "", err
	}

	return t, nil
}

func (h *Handler) Login(c echo.Context) error {
	u := new(model.User)
	if err := c.Bind(u); err != nil {
		return echo.ErrUnauthorized
	}

	if u.Username != "jon" && u.Password != "shhh!" {
		return echo.ErrUnauthorized
	}

	t, err := createJTWToken("jon", h.JwtSigningKey)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, map[string]string{
		"token": t,
	})
}

func (h *Handler) AuthFacebook(c echo.Context) error {
	state := c.FormValue("state")
	if state != h.OauthStateString {
		return c.JSON(http.StatusBadRequest, "invalid oauth state")
	}

	code := c.FormValue("code")
	tok, err := h.OauthConf.Exchange(oauth2.NoContext, code)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error)
	}

	res, err := fb.Get("/me", fb.Params{
		"fields":       "first_name",
		"access_token": tok.AccessToken,
	})

	if err != nil {
		return err
	}

	var firstName string
	res.DecodeField("first_name", &firstName)
	//for demo only, server should not return access token
	//return c.JSON(http.StatusCreated, tok.AccessToken)
	t, err := createJTWToken(firstName, h.JwtSigningKey)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, map[string]string{
		"token": t,
	})
}

func (h *Handler) FacebookIndex(c echo.Context) error {
	url := h.OauthConf.AuthCodeURL(h.OauthStateString, oauth2.AccessTypeOffline)
	return c.JSON(http.StatusOK, url)
}

// func restricted(c echo.Context) error {
// 	user := c.Get("user").(*jwt.Token)
// 	claims := user.Claims.(jwt.MapClaims)
// 	name := claims["name"].(string)
// 	return c.String(http.StatusOK, "Welcome "+name+"!")
// }
