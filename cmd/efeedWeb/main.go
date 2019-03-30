package main

import (
	"efeed"
	"log"
	"net/http"
	"os"
	"path"
	"runtime"

	"github.com/gorilla/sessions"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/kardianos/osext"
	"github.com/robfig/cron"
	"github.com/rs/cors"
	"github.com/spf13/viper"
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
	db     *gorm.DB
}

// globalPresenter contains the fields neccessary for presenting in all templates
type globalPresenter struct {
	SiteName    string
	Description string
	SiteURL     string
}

// SetupApp for main
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
	db, err := efeed.OpenDB(config.URI)
	if err != nil {
		log.Fatalln("cannot connect to db: ", err)
	}

	return &App{
		router: r,
		gp:     gp,
		logr:   logger,
		config: config,
		db:     db,
	}
}

func main() {
	pwd, err := osext.ExecutableFolder()
	if err != nil {
		log.Fatalln("cannot retrieve present working directory: ", 0600, nil)
	}

	err = LoadConfiguration(pwd)
	if err != nil && os.Getenv("PORT") == "" {
		log.Panicln("panicking, Fatal error config file:", err)
	}

	logr := newLogger()
	r := NewRouter()

	a := SetupApp(r, logr, "")
	defer efeed.CloseDB()

	// Add CORS support (Cross Origin Resource Sharing)
	corsSetting := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://efeed.me", "http://www.efeed.me", "http://localhost:3000", "https://efeed-client.herokuapp.com"},
		AllowCredentials: true,
	})
	handler := corsSetting.Handler(r)
	if a.config.IsDevelopment == "true" {
		handler = cors.Default().Handler(r)
	}

	go a.RunCrawlerSoccerProAndSave()
	go a.RunCrawlerFanaticsAndSave()

	c := cron.New()
	err = c.AddFunc("@every 12h", func() {
		err = a.RunCrawlerFanaticsAndSave()
		if err != nil {
			log.Println("error running RunCrawlerOpenDotaTeamAndSave ", err)
		}
	})
	if err != nil {
		log.Println("error on cron job ", err)
	}
	err = c.AddFunc("@every 12h", func() {
		err = a.RunCrawlerSoccerProAndSave()
		if err != nil {
			log.Println("error running RunCrawlerOpenDotaTeamAndSave ", err)
		}
	})
	if err != nil {
		log.Println("error on cron job ", err)
	}

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
