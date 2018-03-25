package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/fsnotify/fsnotify"
	fb "github.com/huandu/facebook"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/facebook"
)

//Task used to persist
type Task struct {
	ID int64 `gorm:"primary_key" json:"id"`
	// gorm.Model
	Title       string     `gorm:"size:255" json:"title"`
	Description string     `gorm:"size:255" json:"description"`
	Priority    int        `gorm:"type:int"`
	CreatedAt   time.Time  `gorm:"default:now()"`
	UpdatedAt   time.Time  `gorm:"default:now()"`
	CompletedAt *time.Time `gorm:"default:null"`
	IsDeleted   bool       `gorm:"default:false"`
	IsCompleted bool       `gorm:"default:false"`
	DeletedAt   *time.Time `gorm:"default:null"`
}

type (
	user struct {
		ID       int    `json:"id"`
		Name     string `json:"name"`
		Password string `json:"password"`
		Username string `json:"username"`
		Email    string `json:"email"`
	}
)

var db *gorm.DB

var oauthConf = &oauth2.Config{
	ClientID:     "1420755331473536",
	ClientSecret: "somesecret",
	Scopes:       []string{"email"},
	Endpoint:     facebook.Endpoint,
	RedirectURL:  "https://thissite.com/auth/facebook",
}

var oauthStateString = "thisshouldberandom"
var jwtSigningKey = "secret"

func main() {
	initApp()

	logger, _ := zap.NewProduction()
	defer logger.Sync() // flushes buffer, if any

	e := echo.New()
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowMethods: []string{echo.GET, echo.POST, echo.PATCH, echo.DELETE},
	}))

	logFile, err := os.OpenFile("log.txt", os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	if err != nil {
		panic(err)
	}
	mw := io.MultiWriter(os.Stdout, logFile)

	// Middleware

	lg := middleware.LoggerWithConfig(middleware.LoggerConfig{
		Output: mw,
	})
	e.Use(lg)

	e.Logger.SetOutput(mw)
	// e.Use(middleware.Logger())
	// e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
	// 	Output: mw,
	// }))
	e.Use(ZapLogger(logger))
	e.Use(middleware.Recover())
	// Login route
	e.POST("/login", login)
	e.POST("/login/facebook", facebookIndex)
	e.GET("/auth/facebook", authFacebook)

	g := e.Group("/tasks")

	//waiting for merge of https://github.com/labstack/echo/pull/1041 as it will allow to return 403
	// g.Use(middleware.JWTWithConfig(middleware.JWTConfig{
	// 	SigningKey: []byte(jwtSigningKey),
	// }))
	g.GET("/:id", getTask)
	g.PATCH("/:id", updateTask)
	g.POST("", createTask)
	g.DELETE("/:id", deleteTask)
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

	defer db.Close()

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
	db.AutoMigrate(&Task{})
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

func login(c echo.Context) error {
	u := new(user)
	if err := c.Bind(u); err != nil {
		return echo.ErrUnauthorized
	}

	if u.Username != "jon" && u.Password != "shhh!" {
		return echo.ErrUnauthorized
	}

	t, err := createJTWToken("jon")
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, map[string]string{
		"token": t,
	})

}

func createJTWToken(name string) (string, error) {
	// Create token
	token := jwt.New(jwt.SigningMethodHS256)

	// Set claims
	claims := token.Claims.(jwt.MapClaims)
	claims["name"] = name
	claims["admin"] = true
	claims["exp"] = time.Now().Add(time.Hour * 72).Unix()

	// Generate encoded token and send it as response.
	t, err := token.SignedString([]byte(jwtSigningKey))
	if err != nil {
		return "", err
	}

	return t, nil
}

func facebookIndex(c echo.Context) error {
	url := oauthConf.AuthCodeURL(oauthStateString, oauth2.AccessTypeOffline)
	return c.JSON(http.StatusOK, url)
}

func authFacebook(c echo.Context) error {
	state := c.FormValue("state")
	if state != oauthStateString {
		return c.JSON(http.StatusBadRequest, "invalid oauth state")
	}

	code := c.FormValue("code")
	tok, err := oauthConf.Exchange(oauth2.NoContext, code)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error)
	}

	res, err := fb.Get("/me", fb.Params{
		"fields":       "first_name",
		"access_token": tok.AccessToken,
	})

	if err != nil {
		return err
	}

	var firstName string
	res.DecodeField("first_name", &firstName)
	//for demo only, server should not return access token
	//return c.JSON(http.StatusCreated, tok.AccessToken)
	t, err := createJTWToken(firstName)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, map[string]string{
		"token": t,
	})
}

func accessible(c echo.Context) error {
	return c.String(http.StatusOK, "Accessible")
}

func restricted(c echo.Context) error {
	user := c.Get("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	name := claims["name"].(string)
	return c.String(http.StatusOK, "Welcome "+name+"!")
}

func createTask(c echo.Context) error {
	task := &Task{}
	if err := c.Bind(&task); err != nil {
		return err
	}

	db.Create(&task)

	return c.JSON(http.StatusCreated, task)
}

func deleteTask(c echo.Context) error {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	task := &Task{ID: id}
	db.Model(&task).Update("IsDeleted", true).Delete(&task)

	return c.JSON(http.StatusNoContent, nil)
}

func getTask(c echo.Context) error {
	c.Logger().Info("test YOOOOOOOOOOO")
	c.Logger().Print("AAAA")

	id := c.Param("id")
	task := &Task{}
	if err := db.Where("is_deleted = ?", 0).First(&task, id).Error; err != nil {
		// c.Logger(err)
		return echo.ErrNotFound
	}

	return c.JSON(http.StatusOK, task)
}

func updateTask(c echo.Context) error {
	task := new(Task)
	if err := c.Bind(&task); err != nil {
		return err
	}

	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if task.ID != id {
		return c.JSON(http.StatusBadRequest, nil)
	}

	db.Save(&task)

	return c.JSON(http.StatusCreated, task)
}
