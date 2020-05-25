package routes

import (
	"github.com/discordextremelist/api/ratelimit"
	"github.com/discordextremelist/api/util"
	"github.com/go-chi/chi"
	"net/http"
	"time"
)

func Stats(w http.ResponseWriter, _ *http.Request) {
	result := util.APIStatsResponse{}
	err, servers := util.GetAllServers(false)
	if err != nil {
		util.WriteJson(500, w, util.GetServersFailed)
		return
	}
	result.Servers = util.APIStatsResponseServers{Total: len(servers)}
	err, bots := util.GetAllBots(false)
	if err != nil {
		util.WriteJson(500, w, util.GetBotsFailed)
		return
	}
	botRes := util.APIStatsResponseBots{Total: len(bots), Approved: 0, Premium: 0}
	for _, bot := range bots {
		if bot.Status.Approved {
			botRes.Approved++
		}
		if bot.Status.Premium {
			botRes.Premium++
		}
	}
	result.Bots = botRes
	err, users := util.GetAllUsers(false)
	if err != nil {
		util.WriteJson(500, w, util.GetUsersFailed)
		return
	}
	userRes := util.APIStatsResponseUsers{Total: len(users), Premium: 0, Staff: util.APIStatsResponseStaff{Total: 0, Mods: 0, Assistants: 0, Admins: 0}}
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
	util.WriteJson(200, w, result)
}

func Health(w http.ResponseWriter, _ *http.Request) {
	result := util.APIHealthResponse{
		Status:  200,
		RedisOK: true,
		MongoOK: true,
	}
	if !util.Database.IsRedisOpen() {
		result.Status = http.StatusServiceUnavailable
		result.RedisOK = false
	}
	if !util.Database.IsMongoOpen() {
		if result.Status != http.StatusServiceUnavailable {
			result.Status = http.StatusServiceUnavailable
		}
		result.MongoOK = false
	}
	util.WriteJson(result.Status, w, result)
}

func InitGeneralRoutes() {
	ratelimiter := ratelimit.NewRatelimiter(ratelimit.RatelimiterOptions{
		Limit:         3,
		Reset:         1000,
		RedisPrefix:   "rl_general",
		TempBanLength: 48 * time.Hour,
		TempBanAfter:  3,
		PermBanAfter:  2,
	})
	util.Router.Route("/", func(r chi.Router) {
		r.Use(ratelimiter.Ratelimit)
		r.Get("/health", Health)
		r.Get("/stats", Stats)
	})
}
