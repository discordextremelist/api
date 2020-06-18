package main

import (
	"fmt"
	"github.com/discordextremelist/api/entities"
	"github.com/discordextremelist/api/routes"
	"github.com/discordextremelist/api/util"
	"github.com/go-chi/chi"
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
)

var (
	check = []string{"ADDR", "PORT", "REDIS_IP", "REDIS_PORT", "REDIS_PASSWORD", "REDIS_DB", "MONGO_URL", "MONGO_DB"}
)

func init() {
	for _, v := range os.Args {
		if v == "--dev" {
			util.Dev = true
		}
	}
	log.SetLevel(log.DebugLevel)
	log.SetFormatter(&log.TextFormatter{ForceColors: true, FullTimestamp: true})
	_ = godotenv.Load()
	for i := 0; i < len(check); i++ {
		_, ok := os.LookupEnv(check[i])
		if !ok {
			log.Fatalf("Environmental variable %s doesn't exist!", check[i])
		}
	}
}

func main() {
	util.Database.OpenRedisConnection()
	util.Database.OpenMongoConnection()
	util.Router = chi.NewRouter()
	util.Router.Use(entities.TokenValidator)
	util.Router.Use(util.RealIP)
	util.Router.Use(util.RequestLogger)
	util.Router.NotFound(entities.NotFound)
	routes.InitGeneralRoutes()
	routes.InitBotRoutes()
	routes.InitUserRoutes()
	routes.InitServerRoutes()
	routes.InitTemplateRoutes()
	ip := os.Getenv("ADDR")
	port := os.Getenv("PORT")
	serve := fmt.Sprintf("%s:%s", ip, port)
	log.Infof("Starting to serve at: %s", serve)
	log.Fatal(http.ListenAndServe(fmt.Sprintf("%s:%s", ip, port), util.Router))
}
