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
	check = []string{"ADDR", "PORT", "REDIS_PASSWORD", "REDIS_DB", "MONGO_URL", "MONGO_DB"}
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
		if _, ok := os.LookupEnv(check[i]); !ok {
			log.Fatalf("Required environmental variable '%s' doesn't exist!", check[i])
		}
	}
}

func main() {
	if !util.Dev {
		util.BuildClient()
		err := util.FindKubernetesNode()
		if err != nil {
			log.Errorf("Failed to get the node this pod is on: %v", err)
		} else {
			log.Infof("Currently on node: %s", util.Node)
		}
	}
	util.Database.OpenRedisConnection()
	util.Database.OpenMongoConnection()
	if util.Dev {
		entities.PopulateDevCache()
	}
	util.Router = chi.NewRouter()
	util.Router.Use(util.RealIP)
	util.Router.Use(entities.RequestLogger)
	util.Router.NotFound(entities.NotFound)
	util.Router.Use(entities.TokenValidator)
	routes.InitGeneralRoutes()
	routes.InitBotRoutes()
	routes.InitUserRoutes()
	routes.InitServerRoutes()
	routes.InitTemplateRoutes()
	routes.InitDebugRoutes()
	ip := os.Getenv("ADDR")
	port := os.Getenv("PORT")
	serve := fmt.Sprintf("%s:%s", ip, port)
	log.Infof("Starting to serve at: %s", serve)
	log.Fatal(http.ListenAndServe(serve, util.Router))
}
