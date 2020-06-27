package util

import (
	"github.com/discordextremelist/api/database"
	"github.com/go-chi/chi"
	"github.com/json-iterator/go"
)

var (
	IPWhitelist []string
	Database    = database.NewManager()
	Json        = jsoniter.ConfigFastest
	Router      chi.Router
	Dev         bool
)

func CheckIP(remoteAddr string) bool {
	for _, ip := range IPWhitelist {
		if remoteAddr == ip {
			return true
		}
	}
	return false
}
