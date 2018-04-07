package main

import (
	"fmt"
	"strconv"

	logrusmiddleware "github.com/bakatz/echo-logrusmiddleware"
	"github.com/fsnotify/fsnotify"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/labstack/gommon/log"
	"github.com/lavasov/gorest/handler"
	"github.com/lavasov/gorest/model"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/facebook"
)

var (
	db        *gorm.DB
	oauthConf = &oauth2.Config{
		ClientID:     "1420755331473536",
		ClientSecret: "somesecret",
		Scopes:       []string{"email"},
		Endpoint:     facebook.Endpoint,
		RedirectURL:  "https://thissite.com/auth/facebook",
	}

	oauthStateString = "thisshouldberandom"
	jwtSigningKey    = "secret"
)

//TODO http://www.alexedwards.net/blog/organising-database-access
//https://gist.github.com/elithrar/5aef354a54ba71a32e23
//https://www.netlify.com/blog/2016/10/20/building-a-restful-api-in-go/
//https://github.com/mingrammer/go-todo-rest-api-example/blob/master/app/handler/projects.go
//https://stevenwhite.com/building-a-rest-service-with-golang-2/
//https://www.netlify.com/blog/2016/10/20/building-a-restful-api-in-go/
func main() {
	initApp()
	defer db.Close()

	appContext := &handler.AppContext{
		DB:               db,
		OauthConf:        oauthConf,
		OauthStateString: oauthStateString,
		JwtSigningKey:    jwtSigningKey,
	}

	handler := &handler.Handler{appContext}

	e := echo.New()
	setupLogger(e)

	// Middleware
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowMethods: []string{echo.GET, echo.POST, echo.PATCH, echo.DELETE},
	}))
	e.Use(middleware.Recover())

	// Handlers
	e.POST("/login", handler.Login)
	e.POST("/login/facebook", handler.FacebookIndex)
	e.GET("/auth/facebook", handler.AuthFacebook)

	g := e.Group("/tasks")
	//waiting for merge of https://github.com/labstack/echo/pull/1041 as it will allow to return 403
	g.Use(middleware.JWTWithConfig(middleware.JWTConfig{
		SigningKey: []byte(jwtSigningKey),
	}))
	g.GET("/:id", handler.GetTask)
	g.PATCH("/:id", handler.UpdateTask)
	g.POST("", handler.CreateTask)
	g.DELETE("/:id", handler.DeleteTask)
	e.OPTIONS("/tasks", nil)

	port := fmt.Sprintf(":%s", strconv.Itoa(viper.GetInt("port")))
	e.Logger.Fatal(e.Start(port))
}

func initApp() {
	readConfig()
	// viper.Debug()

	var err error
	err = connectToDB(
		viper.GetString("db.host"),
		viper.GetString("db.name"),
		viper.GetString("db.user"),
		viper.GetString("db.password"))
	if err != nil {
		panic(err)
	}

	viper.OnConfigChange(func(e fsnotify.Event) {
		fmt.Println("Config file changed:", e.Name)
		err = db.Close()
		if err != nil {
			panic(err)
		}

		err = connectToDB(
			viper.GetString("db.host"),
			viper.GetString("db.name"),
			viper.GetString("db.user"),
			viper.GetString("db.password"))
		if err != nil {
			panic(err)
		}
	})

	db.LogMode(true)

	//TODO use https://github.com/go-gormigrate/gormigrate
	db.AutoMigrate(&model.Task{})
}

func readConfig() {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s", err))
	}

	viper.WatchConfig()
	pflag.Int("port", 1234, "port to litsen")
	pflag.String("db.name", "", "database name")
	pflag.String("db.host", "localhost", "database host")
	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)
}

func connectToDB(host string, dbname string, user string, pass string) error {
	dsn := fmt.Sprintf("host=%s dbname=%s user=%s sslmode=disable password=%s", host, dbname, user, pass)
	var err error
	db, err = gorm.Open("postgres", dsn)

	return err
}

func setupLogger(e *echo.Echo) {
	logrusLogger := logrus.StandardLogger()
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetLevel(logrus.InfoLevel)
	e.Logger = logrusmiddleware.Logger{logrusLogger}
	e.Use(logrusmiddleware.Hook())
	e.Logger.SetLevel(log.INFO)
}
