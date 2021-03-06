package main

import (
	"efeed"
	"log"
	"net/http"
	"os"
	"path"
	"runtime"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gorilla/sessions"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/kardianos/osext"
	"github.com/robfig/cron"
	"github.com/rs/cors"
	"github.com/spf13/viper"
)

type efeedConfig struct {
	Port                string
	DatabaseURL         string
	IsDevelopment       bool
	DoAccessKey         string
	DoSecretAccessKey   string
	DoEndpoint          string
	DoBucket            string
	DoSpaceURL          string
	EnableCrawling      bool
	EnableProxy         bool
	ProxyURL            string
	EnableReuploadImage bool
}

// App in main app
type App struct {
	router *Router
	gp     globalPresenter
	logr   appLogger
	config efeedConfig
	store  *sessions.CookieStore
	db     *gorm.DB
	svc    *s3.S3
}

// globalPresenter contains the fields neccessary for presenting in all templates
type globalPresenter struct {
	SiteName    string
	Description string
	SiteURL     string
}

func getDoClient(config efeedConfig) (*s3.S3, error) {
	sess := session.New(&aws.Config{
		Region:      aws.String("ap-southeast-1"),
		Endpoint:    aws.String(config.DoEndpoint),
		Credentials: credentials.NewStaticCredentials(config.DoAccessKey, config.DoSecretAccessKey, ""),
	})

	svc := s3.New(sess)

	return svc, nil
}

// SetupApp for main
func SetupApp(r *Router, logger appLogger, templateDirectoryPath string) *App {
	var config efeedConfig

	if viper.GetBool("isDevelopment") {
		config = efeedConfig{
			DoAccessKey:         viper.GetString("DO_ACCESS_KEY"),
			DoSecretAccessKey:   viper.GetString("DO_SECRET_ACCESS_KEY"),
			DoEndpoint:          viper.GetString("DO_ENDPOINT"),
			DoBucket:            viper.GetString("DO_BUCKET"),
			DoSpaceURL:          viper.GetString("DO_SpaceURL"),
			IsDevelopment:       viper.GetBool("isDevelopment"),
			Port:                viper.GetString("port"),
			DatabaseURL:         viper.GetString("DATABASE_URL"),
			EnableCrawling:      viper.GetBool("ENABLE_CRAWLING"),
			EnableProxy:         viper.GetBool("ENABLE_PROXY"),
			ProxyURL:            viper.GetString("PROXY_URL"),
			EnableReuploadImage: viper.GetBool("ENABLE_REUP_IMAGE"),
		}
	} else {
		enableCrawling, _ := strconv.ParseBool(os.Getenv("ENABLE_CRAWLING"))
		enableProxy, _ := strconv.ParseBool(os.Getenv("ENABLE_PROXY"))
		enableReuploadImage, _ := strconv.ParseBool(os.Getenv("ENABLE_REUP_IMAGE"))
		config = efeedConfig{
			DoAccessKey:         os.Getenv("DO_ACCESS_KEY"),
			DoSecretAccessKey:   os.Getenv("DO_SECRET_ACCESS_KEY"),
			DoEndpoint:          os.Getenv("DO_ENDPOINT"),
			DoBucket:            os.Getenv("DO_BUCKET"),
			DoSpaceURL:          os.Getenv("DO_SpaceURL"),
			Port:                os.Getenv("PORT"),
			DatabaseURL:         os.Getenv("DATABASE_URL"),
			EnableCrawling:      enableCrawling,
			EnableProxy:         enableProxy,
			ProxyURL:            os.Getenv("PROXY_URL"),
			EnableReuploadImage: enableReuploadImage,
		}
	}

	gp := globalPresenter{
		SiteName:    "efeed",
		Description: "Api",
		SiteURL:     "wtf",
	}
	db, err := efeed.OpenDB(config.DatabaseURL)
	if err != nil {
		log.Fatalln("cannot connect to db: ", err)
	}

	svc, err := getDoClient(config)
	if err != nil {
		log.Fatalln("cannot connect to svc: ", err)
	}

	return &App{
		router: r,
		gp:     gp,
		logr:   logger,
		config: config,
		db:     db,
		svc:    svc,
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
	if a.config.IsDevelopment == true {
		handler = cors.Default().Handler(r)
	}

	r.Get("/export", a.Wrap(a.ExportCSVHandler()))
	r.Get("/products/search", a.ProductSearchHandler())

	r.Get("/", a.Index())
	if a.config.EnableCrawling {
		go a.RunCrawlerSoccerProAndSave()
		go a.RunCrawlerFanaticsAndSave()
		go a.RunCrawlerRevzillaAndSave()
	}

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
		err = a.RunCrawlerRevzillaAndSave()
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

	err = c.AddFunc("@every 25m", func() {
		err = a.RunPingHeroku()
		if err != nil {
			log.Println("error running RunPingHeroku ", err)
		}
	})

	c.Start()

	err = http.ListenAndServe(":"+a.config.Port, handler)
	if err != nil {
		log.Println("error on serve server", err)
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
