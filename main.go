package main

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"text/template"

	_ "github.com/go-sql-driver/mysql"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gopkg.in/go-playground/validator.v8"
)

type configuration struct {
	SiteName string `json:"site_name" check:"required"`
	DBLink   string `json:"db_link" check:"required"`
	LogFile  string `json:"log_file" check:"required"`
}

var validate = validator.New(&validator.Config{
	TagName: "check",
})

func (c *configuration) useDefault() {
	*c = configuration{
		SiteName: "Blog",
		DBLink:   "mongodb://localhost:27017/",
		LogFile:  "test_log.txt",
	}
}

// Server - структура сервера
type Server struct {
	db           *mongo.Client
	dbname       string
	dbcollection string
}

// Post - структура одного поста
type Post struct {
	ID      string `bson:"id"`
	Title   string `bson:"title"`
	Date    string `bson:"date"`
	Link    string `bson:"link"`
	Comment string `bson:"comment"`
}

var (
	tmplList   = template.Must(template.New("MyTemplate").ParseFiles("./templates/list.html"))
	tmplSingle = template.Must(template.New("MyTemplate").ParseFiles("./templates/single.html"))
	tmplEdit   = template.Must(template.New("MyTemplate").ParseFiles("./templates/edit.html"))
	tmplCreate = template.Must(template.New("MyTemplate").ParseFiles("./templates/create.html"))
)

func main() {
	bytes, err := ioutil.ReadFile("app.json")
	if err != nil {
		log.Fatal(err)
	}

	Config := new(configuration)
	if err = json.Unmarshal(bytes, Config); err != nil {
		log.Fatal(err)
	}

	if err := validate.Struct(Config); err != nil {
		log.Printf("try to load configuration: %v", err)
		Config.useDefault()
		log.Println("use default")
	}

	// Не получается логи e.Use(middleware.Logger()) - перенаправить в лог файл 'test_log.txt'
	f, err := os.OpenFile(Config.LogFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	log.SetOutput(f)

	db, err := mongo.NewClient(options.Client().ApplyURI(Config.DBLink))
	if err != nil {
		log.Fatal(err)
	}

	if err := db.Connect(context.Background()); err != nil {
		log.Fatal(err)
	}
	defer db.Disconnect(context.Background())

	s := Server{
		db:           db,
		dbname:       "blog_posts",
		dbcollection: "posts",
	}

	e := echo.New()
	e.HideBanner = true
	e.Renderer = &TemplateRenderer{
		templates: template.Must(template.ParseGlob("templates/*.html")),
	}

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.GET("/", s.postsList)
	e.GET("/post/:id", s.singlePostGET)
	e.POST("/post/:id", s.singlePostPOST)
	// для обновления, удаления и создания постов используются методы GET и POST, т.к. необходима
	// работоспособность при использовании базового HTML (через браузер при нажатии кнопок).
	e.GET("/edit/:id", s.editPost)
	e.GET("/create/new", s.createPostGET)
	e.POST("/create/new", s.createPostPOST)
	e.GET("/delete/:id", s.deletePostGET)

	port := "8080"
	log.Fatal(e.Start(":" + port))
}
