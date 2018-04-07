package handler

import (
	"github.com/jinzhu/gorm"
	"golang.org/x/oauth2"
)

type (
	AppContext struct {
		DB               *gorm.DB
		OauthConf        *oauth2.Config
		OauthStateString string
		JwtSigningKey    string
	}
)

type Handler struct {
	*AppContext
}
