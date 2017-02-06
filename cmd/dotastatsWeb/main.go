package main

import (
	"dotastats"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"runtime"

	"github.com/gorilla/context"
	"github.com/justinas/alice"
	"github.com/kardianos/osext"
	"github.com/rs/cors"
	"github.com/spf13/viper"
	"gopkg.in/robfig/cron.v2"
)

type dotastatsConfig struct {
	Port          string
	URI           string
	Dbname        string
	Collection    string
	IsDevelopment string
}

// App in main app
type App struct {
	router  *Router
	gp      globalPresenter
	logr    appLogger
	mongodb dotastats.Mongodb
	config  dotastatsConfig
}

// globalPresenter contains the fields neccessary for presenting in all templates
type globalPresenter struct {
	SiteName    string
	Description string
	SiteURL     string
}

// TODO localPresenter if we have using template
func SetupApp(r *Router, logger appLogger, templateDirectoryPath string) *App {
	var config dotastatsConfig
	if viper.GetBool("isDevelopment") {
		config = dotastatsConfig{
			IsDevelopment: viper.GetString("isDevelopment"),
			Port:          viper.GetString("port"),
			URI:           viper.GetString("uri"),
			Dbname:        viper.GetString("dbname"),
			Collection:    viper.GetString("collection"),
		}
	} else {
		config = dotastatsConfig{
			IsDevelopment: os.Getenv("isDevelopment"),
			Port:          os.Getenv("PORT"),
			URI:           os.Getenv("uri"),
			Dbname:        os.Getenv("dbname"),
			Collection:    os.Getenv("collection"),
		}
	}

	mongo := dotastats.Mongodb{URI: config.URI, Dbname: config.Dbname, Collection: config.Collection}

	gp := globalPresenter{
		SiteName:    "dotastats",
		Description: "Api",
		SiteURL:     "wtf",
	}

	return &App{
		router:  r,
		gp:      gp,
		logr:    logger,
		config:  config,
		mongodb: mongo,
	}
}

func main() {
	pwd, err := osext.ExecutableFolder()
	if err != nil {
		log.Fatalf("cannot retrieve present working directory: %i", 0600, nil)
	}

	err = LoadConfiguration(pwd)
	if err != nil && os.Getenv("PORT") == "" {
		fmt.Println("panicking")
		panic(fmt.Println("Fatal error config file: %s \n", err))
	}

	r := NewRouter()
	logr := newLogger()
	a := SetupApp(r, logr, "")
	// Add CORS support (Cross Origin Resource Sharing)
	common := alice.New(context.ClearHandler, a.loggingHandler, a.recoverHandler)
	r.Get("/f10k/:name", common.Then(a.Wrap(a.GetF10kResultHandler())))
	r.Get("/team/:name", common.Then(a.Wrap(a.GetTeamMatchesHandler())))
	r.Get("/team/:name/f10k", common.Then(a.Wrap(a.GetTeamF10kMatchesHandler())))
	r.Get("/crawl", common.Then(a.Wrap(a.GetCustomCrawlHandler())))
	handler := cors.Default().Handler(r)
	c := cron.New()
	c.AddFunc("@every 5m", func() {
		err = a.RunCrawlerAndSave()
		if err != nil {
			fmt.Println("error running crawler %s", err)
		}
	})
	c.Start()
	err = http.ListenAndServe(":"+a.config.Port, handler)
	if err != nil {
		fmt.Println("error on serve server %s", err)
	}
}

func LoadConfiguration(pwd string) error {
	viper.SetConfigName("dotastats-config")
	viper.AddConfigPath(pwd)
	devPath := pwd[:len(pwd)-3] + "src/dotastats/cmd/dotastatsweb/"
	_, file, _, _ := runtime.Caller(1)
	configPath := path.Dir(file)
	viper.AddConfigPath(devPath)
	viper.AddConfigPath(configPath)
	return viper.ReadInConfig() // Find and read the config file
}
