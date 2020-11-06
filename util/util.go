package util

import (
	"github.com/discordextremelist/api/database"
	"github.com/go-chi/chi"
)

var (
	Database = database.NewManager()
	Router   chi.Router
	Dev      bool
)
