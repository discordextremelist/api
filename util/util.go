package util

import (
	"github.com/discordextremelist/api/database"
	"github.com/go-chi/chi"
	"github.com/json-iterator/go"
)

var (
	Database = database.NewManager()
	Json     = jsoniter.ConfigFastest
	Router   chi.Router
	Dev      bool
)
