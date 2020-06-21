package routes

import (
	"github.com/discordextremelist/api/entities"
	"github.com/discordextremelist/api/ratelimit"
	"github.com/discordextremelist/api/util"
	"github.com/go-chi/chi"
	"net/http"
	"time"
)

func Stats(w http.ResponseWriter, _ *http.Request) {
	result := entities.APIStatsResponse{}
	err, servers := entities.GetAllServers(false)
	if err != nil {
		entities.WriteJson(500, w, entities.GetServersFailed)
		return
	}
	result.Servers = entities.APIStatsResponseServers{Total: len(servers)}
	err, bots := entities.GetAllBots(false)
	if err != nil {
		entities.WriteJson(500, w, entities.GetBotsFailed)
		return
	}
	botRes := entities.APIStatsResponseBots{Total: len(bots), Approved: 0, Premium: 0}
	for _, bot := range bots {
		if bot.Status.Approved {
			botRes.Approved++
		}
		if bot.Status.Premium {
			botRes.Premium++
		}
	}
	result.Bots = botRes
	err, users := entities.GetAllUsers(false)
	if err != nil {
		entities.WriteJson(500, w, entities.GetUsersFailed)
		return
	}
	userRes := entities.APIStatsResponseUsers{Total: len(users), Premium: 0, Staff: entities.APIStatsResponseStaff{Total: 0, Mods: 0, Assistants: 0, Admins: 0}}
	for _, user := range users {
		if user.Rank.Premium {
			userRes.Premium++
		}
		if user.Rank.Admin {
			userRes.Staff.Total++
			userRes.Staff.Admins++
			continue
		}
		if user.Rank.Assistant {
			userRes.Staff.Total++
			userRes.Staff.Assistants++
			continue
		}
		if user.Rank.Mod {
			userRes.Staff.Total++
			userRes.Staff.Mods++
			continue
		}
	}
	result.Users = userRes
	err, templates := entities.GetAllTemplates()
	if err != nil {
		entities.WriteJson(500, w, entities.GetTemplatesFailed)
		return
	}
	result.Templates = len(templates)
	result.Status = 200
	result.Error = false
	entities.WriteJson(200, w, result)
}

func Health(w http.ResponseWriter, _ *http.Request) {
	result := entities.APIHealthResponse{
		Status:  200,
		Error:   false,
		RedisOK: true,
		MongoOK: true,
	}
	if !util.Database.IsRedisOpen() {
		result.Status = http.StatusServiceUnavailable
		result.Error = true
		result.RedisOK = false
	}
	if !util.Database.IsMongoOpen() {
		if result.Status != http.StatusServiceUnavailable {
			result.Status = http.StatusServiceUnavailable
			result.Error = true
		}
		result.MongoOK = false
	}
	entities.WriteJson(result.Status, w, result)
}

func InitGeneralRoutes() {
	ratelimiter := ratelimit.NewRatelimiter(ratelimit.RatelimiterOptions{
		Limit:         3,
		Reset:         5000,
		RedisPrefix:   "rl_general",
		TempBanLength: 1 * time.Hour,
		TempBanAfter:  3,
		PermBanAfter:  2,
	})
	util.Router.Route("/", func(r chi.Router) {
		r.Use(ratelimiter.Ratelimit)
		r.Get("/health", Health)
		r.Get("/stats", Stats)
	})
}
