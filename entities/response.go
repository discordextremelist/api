package entities

import (
	"errors"
	"github.com/discordextremelist/api/util"
	"net/http"
)

type APIHealthResponse struct {
	Error   bool `json:"error"`
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
	Status    int                     `json:"status"`
	Error     bool                    `json:"error"`
	Servers   APIStatsResponseServers `json:"servers"`
	Bots      APIStatsResponseBots    `json:"bots"`
	Users     APIStatsResponseUsers   `json:"users"`
	Templates int                     `json:"templates"`
}

type APIResponse struct {
	Error    bool            `json:"error"`
	Status   int             `json:"status"`
	Message  *string         `json:"message,omitempty"`
	Bot      *Bot            `json:"bot,omitempty"`
	Server   *Server         `json:"server,omitempty"`
	User     *User           `json:"user,omitempty"`
	Template *ServerTemplate `json:"template,omitempty"`
}

type APIResponseBots struct {
	Error  bool  `json:"error"`
	Status int   `json:"status"`
	Bots   []Bot `json:"bots"`
}

func buildInternal(error bool, status int, message string, bot *Bot, server *Server, user *User, template *ServerTemplate) APIResponse {
	ptr := &message
	if message == "" {
		ptr = nil
	}
	return APIResponse{
		Error:    error,
		Status:   status,
		Message:  ptr,
		Bot:      bot,
		Server:   server,
		User:     user,
		Template: template,
	}
}

var (
	RatelimitedError   = buildInternal(true, 429, "Too Many Requests", nil, nil, nil, nil)
	TempBannedError    = buildInternal(true, 403, "You've been temporarily API banned!", nil, nil, nil, nil)
	PermBannedError    = buildInternal(true, 403, "You've been permanently API banned!", nil, nil, nil, nil)
	NotFoundError      = buildInternal(true, 404, "Not Found", nil, nil, nil, nil)
	NoAuthError        = buildInternal(true, 403, `No "Authorization" header specified, or it was invalid!`, nil, nil, nil, nil)
	LookupError        = errors.New("an error occurred when looking up this resource")
	ReadFailed         = errors.New("failed to read request body")
	notImplemented     = buildInternal(true, 501, "Not implemented", nil, nil, nil, nil)
	GetServersFailed   = buildInternal(true, 500, "An error occurred when getting all servers, try again later!", nil, nil, nil, nil)
	GetBotsFailed      = buildInternal(true, 500, "An error occurred when getting all bots, try again later!", nil, nil, nil, nil)
	GetUsersFailed     = buildInternal(true, 500, "An error occurred when getting all users, try again later!", nil, nil, nil, nil)
	GetTemplatesFailed = buildInternal(true, 500, "An error occurred when getting all templates, try again later!", nil, nil, nil, nil)
	BadContentType     = buildInternal(true, 415, "Unsupported Content Type, or non was provided!", nil, nil, nil, nil)
)

func WriteJson(status int, writer http.ResponseWriter, v interface{}) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(status)
	util.Json.NewEncoder(writer).Encode(v)
}

func WriteErrorResponse(w http.ResponseWriter, err error) {
	WriteJson(500, w, buildInternal(true, 500, err.Error(), nil, nil, nil, nil))
}

func WriteBotResponse(w http.ResponseWriter, bot *Bot) {
	WriteJson(200, w, buildInternal(false, 200, "", bot, nil, nil, nil))
}

func WriteUserResponse(w http.ResponseWriter, user *User) {
	WriteJson(200, w, buildInternal(false, 200, "", nil, nil, user, nil))
}

func WriteServerResponse(w http.ResponseWriter, server *Server) {
	WriteJson(200, w, buildInternal(false, 200, "", nil, server, nil, nil))
}

func WriteTemplateResponse(w http.ResponseWriter, template *ServerTemplate) {
	WriteJson(200, w, buildInternal(false, 200, "", nil, nil, nil, template))
}

func WriteNotImplementedResponse(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(501)
	util.Json.NewEncoder(w).Encode(notImplemented)
}

// DELAPI_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx-000000000000000000
func TokenValidator(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health" || util.CheckIP(r.RemoteAddr) {
			next.ServeHTTP(w, r)
			return
		}
		auth := r.Header.Get(util.Authorization)
		if auth != "" && !util.Dev {
			matches := util.TokenPattern.FindStringSubmatch(auth)
			if len(matches) < 2 {
				BadAuth(w, r)
				return
			}
			err, bot := LookupBot(matches[1], false)
			if err != nil {
				BadAuth(w, r)
			} else {
				if bot.Token != auth {
					BadAuth(w, r)
				} else {
					next.ServeHTTP(w, r)
				}
			}
		} else if auth == "" && !util.Dev {
			BadAuth(w, r)
		} else {
			next.ServeHTTP(w, r)
		}
	})
}

func NotFound(w http.ResponseWriter, _ *http.Request) {
	WriteJson(http.StatusNotFound, w, NotFoundError)
}

func BadAuth(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(403)
	util.Json.NewEncoder(w).Encode(NoAuthError)
}
