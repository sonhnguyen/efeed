package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"runtime"

	"github.com/gorilla/sessions"
	"github.com/kardianos/osext"
	"github.com/rs/cors"
	"github.com/spf13/viper"
	cron "gopkg.in/robfig/cron.v2"
)

type efeedConfig struct {
	Port          string
	URI           string
	Dbname        string
	IsDevelopment string
}

// App in main app
type App struct {
	router *Router
	gp     globalPresenter
	logr   appLogger
	config efeedConfig
	store  *sessions.CookieStore
}

// globalPresenter contains the fields neccessary for presenting in all templates
type globalPresenter struct {
	SiteName    string
	Description string
	SiteURL     string
}

// TODO localPresenter if we have using template
func SetupApp(r *Router, logger appLogger, templateDirectoryPath string) *App {
	var config efeedConfig
	if viper.GetBool("isDevelopment") {
		config = efeedConfig{
			IsDevelopment: viper.GetString("isDevelopment"),
			Port:          viper.GetString("port"),
			URI:           viper.GetString("uri"),
			Dbname:        viper.GetString("dbname"),
		}
	} else {
		config = efeedConfig{
			IsDevelopment: os.Getenv("isDevelopment"),
			Port:          os.Getenv("PORT"),
			URI:           os.Getenv("uri"),
			Dbname:        os.Getenv("dbname"),
		}
	}

	if viper.GetBool("isLocal") {
		config.URI = viper.GetString("uriLocal")
	}

	gp := globalPresenter{
		SiteName:    "efeed",
		Description: "Api",
		SiteURL:     "wtf",
	}

	return &App{
		router: r,
		gp:     gp,
		logr:   logger,
		config: config,
	}
}

func main() {
	pwd, err := osext.ExecutableFolder()
	if err != nil {
		log.Fatalf("cannot retrieve present working directory: %i", 0600, nil)
	}
	fmt.Println("hello")

	err = LoadConfiguration(pwd)
	if err != nil && os.Getenv("PORT") == "" {
		log.Panicln("panicking, Fatal error config file: %s", err)
	}

	logr := newLogger()
	r := NewRouter()

	a := SetupApp(r, logr, "")

	// Add CORS support (Cross Origin Resource Sharing)
	corsSetting := cors.New(cors.Options{
		AllowedOrigins:   []string{"https://f10k.herokuapp.com", "http://efeed.me", "http://www.efeed.me", "http://localhost:3000", "https://efeed-client.herokuapp.com"},
		AllowCredentials: true,
	})
	handler := corsSetting.Handler(r)
	if a.config.IsDevelopment == "true" {
		handler = cors.Default().Handler(r)
	}
	fmt.Println("hello")

	c := cron.New()
	_, err = c.AddFunc("@every 1s", func() {
		err = a.RunCrawlerFanaticsAndSave()
		if err != nil {
			log.Println("error running RunCrawlerOpenDotaTeamAndSave %s", err)
		}
	})
	if err != nil {
		log.Println("error on cron job %s", err)
	}
	fmt.Println()
	c.Start()
	err = http.ListenAndServe(":"+a.config.Port, handler)
	if err != nil {
		log.Println("error on serve server %s", err)
	}
}

func LoadConfiguration(pwd string) error {
	viper.SetConfigName("efeed-config")
	viper.AddConfigPath(pwd)
	devPath := pwd[:len(pwd)-3] + "src/efeed/cmd/efeedweb/"
	_, file, _, _ := runtime.Caller(1)
	configPath := path.Dir(file)
	viper.AddConfigPath(devPath)
	viper.AddConfigPath(configPath)
	return viper.ReadInConfig() // Find and read the config file
}
