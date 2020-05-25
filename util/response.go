package util

import (
	"github.com/discordextremelist/api/entities"
	"github.com/pkg/errors"
	"net/http"
)

type APIHealthResponse struct {
	Status  int  `json:"status"`
	RedisOK bool `json:"redis_ok"`
	MongoOK bool `json:"mongo_ok"`
}

type APIStatsResponseServers struct {
	Total int `json:"total"`
}

type APIStatsResponseBots struct {
	Total    int `json:"total"`
	Approved int `json:"approved"`
	Premium  int `json:"premium"`
}

type APIStatsResponseUsers struct {
	Total   int                   `json:"total"`
	Premium int                   `json:"premium"`
	Staff   APIStatsResponseStaff `json:"staff"`
}

type APIStatsResponseStaff struct {
	Total      int `json:"total"`
	Mods       int `json:"mods"`
	Assistants int `json:"assistants"`
	Admins     int `json:"admins"`
}

type APIStatsResponse struct {
	Servers APIStatsResponseServers `json:"servers"`
	Bots    APIStatsResponseBots    `json:"bots"`
	Users   APIStatsResponseUsers   `json:"users"`
}

type APIResponse struct {
	Error   bool             `json:"error"`
	Status  int              `json:"status"`
	Message *string          `json:"message,omitempty"`
	Bot     *entities.Bot    `json:"bot,omitempty"`
	Server  *entities.Server `json:"server,omitempty"`
	User    *entities.User   `json:"user,omitempty"`
}

func buildInternal(error bool, status int, message string, bot *entities.Bot, server *entities.Server, user *entities.User) APIResponse {
	ptr := &message
	if message == "" {
		ptr = nil
	}
	return APIResponse{
		Error:   error,
		Status:  status,
		Message: ptr,
		Bot:     bot,
		Server:  server,
		User:    user,
	}
}

var (
	RatelimitedError = buildInternal(true, 429, "Too Many Requests", nil, nil, nil)
	TempBannedError  = buildInternal(true, 403, "You've been temporarily API banned!", nil, nil, nil)
	PermBannedError  = buildInternal(true, 403, "You've been permanently API banned!", nil, nil, nil)
	NotFoundError    = buildInternal(true, 404, "Not Found", nil, nil, nil)
	NoAuthError      = buildInternal(true, 403, `No "Authorization" header specified, or it was invalid!`, nil, nil, nil)
	LookupError      = errors.New("An error occurred when looking up this resource, try again later!")
	notImplemented   = buildInternal(true, 501, "Not implemented", nil, nil, nil)
	GetServersFailed = buildInternal(true, 500, "An error occurred when getting all servers, try again later!", nil, nil, nil)
	GetBotsFailed    = buildInternal(true, 500, "An error occurred when getting all bots, try again later!", nil, nil, nil)
	GetUsersFailed   = buildInternal(true, 500, "An error occurred when getting all users, try again later!", nil, nil, nil)
)

func WriteJson(status int, writer http.ResponseWriter, v interface{}) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(status)
	Json.NewEncoder(writer).Encode(v)
}

func WriteErrorResponse(w http.ResponseWriter, err error) {
	WriteJson(500, w, buildInternal(true, 500, err.Error(), nil, nil, nil))
}

func WriteBotResponse(w http.ResponseWriter, bot *entities.Bot) {
	WriteJson(200, w, buildInternal(false, 200, "", bot, nil, nil))
}

func WriteUserResponse(w http.ResponseWriter, user *entities.User) {
	WriteJson(200, w, buildInternal(false, 200, "", nil, nil, user))
}

func WriteServerResponse(w http.ResponseWriter, server *entities.Server) {
	WriteJson(200, w, buildInternal(false, 200, "", nil, server, nil))
}

func WriteNotImplementedResponse(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(501)
	Json.NewEncoder(w).Encode(notImplemented)
}
