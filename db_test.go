package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"reflect"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// описываем структуру передаваемых данных для теста
type testStruct struct {
	Method dbMethod
	ID     string
	Post   Post

	ExpectedPosts []Post
	ExpectedPost  Post
}

type dbMethod string

const (
	readAll dbMethod = "readAll"
	readOne dbMethod = "readOne"
	create  dbMethod = "create"
	update  dbMethod = "update"
	delete  dbMethod = "delete"
)

func createTests() []testStruct {
	initialPosts := createPosts()
	return []testStruct{
		// i = 0, show all
		{
			Method:        readAll,
			ExpectedPosts: initialPosts, // the first our state
		},
		// i = 1, show one
		{
			Method: readOne,
			ID:     "1",
			ExpectedPost: Post{
				ID:      "1",
				Title:   "How to Delete an Instagram Account",
				Date:    "2020-07-22 21:49:00",
				Link:    "https://www.wikihow.com/Delete-an-Instagram-Account",
				Comment: "his wikiHow teaches you how to permanently delete your Instagram account.",
			},
		},
		// i = 2, add one
		{
			Method: create,
			Post: Post{
				ID:      "27",
				Title:   "Test title",
				Date:    "test date",
				Link:    "test link",
				Comment: "test comment",
			},
		},
		// i = 3, show all (check previous case)
		{
			Method: readAll,
			ExpectedPosts: append(initialPosts, Post{
				ID:      "27",
				Title:   "Test title",
				Date:    "test date",
				Link:    "test link",
				Comment: "test comment",
			}),
		},
		// i = 4, edit one
		{
			Method: update,
			ID:     "27",
			Post: Post{
				ID:      "27",
				Title:   "Test update title",
				Date:    "test update date",
				Link:    "test update link",
				Comment: "test update comment",
			},
		},
		// i = 5, show all (check previous case)
		{
			Method: readAll,
			ID:     "27",
			ExpectedPosts: append(initialPosts, Post{
				ID:      "27",
				Title:   "Test update title",
				Date:    "test update date",
				Link:    "test update link",
				Comment: "test update comment",
			}),
		},
		// i = 6, delete one
		{
			Method: delete,
			ID:     "27",
		},
		// i = 7, show all (check previous case)
		{
			Method:        readAll,
			ExpectedPosts: initialPosts,
		},
	}
}

func TestDatabase(t *testing.T) {
	// инициализируем запуск сервера монго
	st, err := initDb()
	if err != nil {
		t.Error(err)
		return
	}

	// анонимная функция вызывается после завершения основной программы
	// метод truncate - удаляет данные в тестовой ДБ
	// метод Discinnect - прерывает соединение с ДБ
	defer func() {
		_ = st.truncate()
		_ = st.db.Disconnect(context.Background())
	}()

	// создаем тесты
	tests := createTests()

	// перебираем тесты
	for i, test := range tests {
		var (
			posts = make([]Post, 0, 1)
			post  = Post{}
			err   error
		)

		switch test.Method {
		case readAll:
			posts, err = getAllPosts(st.db)
		case readOne:
			post, err = getPost(st.db, test.ID)
		case create:
			err = createPost(st.db, test.Post)
		case update:
			err = updatePost(st.db, test.Post)
		case delete:
			err = deletePost(st.db, test.ID)
		default:
			t.Error("unknown method")
			continue
		}

		if err != nil {
			t.Error(err)
		}

		if test.Method == readAll {
			if !reflect.DeepEqual(posts, test.ExpectedPosts) {
				t.Errorf("[%d] Expected: %v; Result: %v", i, test.ExpectedPosts, posts)
				break
			}
		} else if test.Method == readOne {
			if !reflect.DeepEqual(post, test.ExpectedPost) {
				t.Errorf("[%d] Expected: %v; Result: %v", i, test.ExpectedPost, post)
				break
			}
		}
	}

}

func initDb() (Server, error) {
	bytes, err := ioutil.ReadFile("app.json")
	if err != nil {
		log.Fatal(err)
	}

	Config := new(configuration)
	if err = json.Unmarshal(bytes, Config); err != nil {
		log.Fatal(err)
	}

	db, err := mongo.NewClient(options.Client().ApplyURI(Config.DBLink))
	if err != nil {
		return Server{}, err
	}

	if err = db.Connect(context.Background()); err != nil {
		return Server{}, err
	}
	fmt.Println("mongo connected")

	st := Server{
		db:           db,
		dbname:       "blog_posts_test",
		dbcollection: "posts_test",
	}

	// запускаем функцию создания 2х постов в БД
	if err := st.insertDefault(); err != nil {
		return Server{}, err
	}

	return st, nil
}

func (st Server) insertDefault() error {
	for _, post := range createPosts() {
		if err := createPost(st.db, post); err != nil {
			return err
		}
	}

	return nil
}

// создаем слайс начальных 2х постов
func createPosts() []Post {
	return []Post{
		{
			ID:      "0",
			Title:   "How to Keep Your Resume to One Page",
			Date:    "2020-07-17 21:49:00",
			Link:    "https://www.wikihow.com/Keep-Your-Resume-to-One-Page",
			Comment: "While a longer resume may be merited if you are applying for an executive-level.",
		},
		{
			ID:      "1",
			Title:   "How to Delete an Instagram Account",
			Date:    "2020-07-22 21:49:00",
			Link:    "https://www.wikihow.com/Delete-an-Instagram-Account",
			Comment: "his wikiHow teaches you how to permanently delete your Instagram account.",
		},
	}
}

// truncate() - функция для удаления всей информации в БД
func (st Server) truncate() error {
	c := st.db.Database(st.dbname).Collection(st.dbcollection)
	_, err := c.DeleteMany(context.Background(), bson.D{})
	// лог для отладки очистк данных в тестовой БД
	fmt.Printf("Database %v, collection %v: truncated \n", st.dbname, st.dbcollection)

	return err
}
