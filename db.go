package main

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	dbName       = "blog_posts_test"
	dbCollection = "posts_test"

	// dbName       = "blog_posts"
	// dbCollection = "posts"
)

// getAllLists — получение всех списков с задачами
func getAllPosts(db *mongo.Client) ([]Post, error) {
	c := db.Database(dbName).Collection(dbCollection)

	cur, err := c.Find(context.Background(), bson.D{})
	if err != nil {
		return nil, errors.Wrap(err, "Found an error in getAllPosts() func, specifically in cur, err := c.Find(context.Background(), bson.D{})")
	}

	res := make([]Post, 0, 1)
	if err := cur.All(context.Background(), &res); err != nil {
		return nil, errors.Wrap(err, "All")
	}

	return res, nil
}

// getList — получение поста из ДБ по id
func getPost(db *mongo.Client, id string) (Post, error) {
	c := db.Database(dbName).Collection(dbCollection)

	filter := bson.D{{Key: "id", Value: id}}

	res := c.FindOne(context.Background(), filter)

	post := new(Post)
	if err := res.Decode(post); err != nil {
		return Post{}, errors.Wrap(err, "decode")
	}

	return *post, nil
}

// updatePost - обновление существующего поста в БД
func updatePost(db *mongo.Client, upgPost Post) error {
	c := db.Database(dbName).Collection(dbCollection)

	if len(upgPost.Date) == 0 {
		upgPost.Date = time.Now().Format("2006-01-02 15:04:05")
	}

	if len(upgPost.Title) == 0 {
		return errors.New("EMPTY title")
	}

	filter := bson.D{{Key: "id", Value: upgPost.ID}}
	update := bson.D{}

	if len(upgPost.ID) != 0 {
		update = append(update, bson.E{Key: "id", Value: upgPost.ID})
	}

	if len(upgPost.Title) != 0 {
		update = append(update, bson.E{Key: "title", Value: upgPost.Title})
	}

	if len(upgPost.Date) != 0 {
		update = append(update, bson.E{Key: "date", Value: upgPost.Date})
	}

	if len(upgPost.Link) != 0 {
		update = append(update, bson.E{Key: "link", Value: upgPost.Link})
	}

	if len(upgPost.Comment) != 0 {
		update = append(update, bson.E{Key: "comment", Value: upgPost.Comment})
	}

	update = bson.D{{Key: "$set", Value: update}}
	_, err := c.UpdateOne(context.Background(), filter, update)

	return err
}

// createPost - обновление существующего поста в БД
func createPost(db *mongo.Client, post Post) error {
	c := db.Database(dbName).Collection(dbCollection)

	if len(post.Date) == 0 {
		post.Date = time.Now().Format("2006-01-02 15:04:05")
	}

	if len(post.Title) == 0 {
		return errors.New("EMPTY title")
	}

	res, err := c.InsertOne(context.Background(), post)
	fmt.Println(res)

	return err
}

// deletePost - удаление существующего поста в БД
func deletePost(db *mongo.Client, id string) error {
	c := db.Database(dbName).Collection(dbCollection)
	filter := bson.D{{Key: "id", Value: id}}

	_, err := c.DeleteOne(context.Background(), filter)

	return err
}
