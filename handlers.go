package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
)

type response struct {
	Result string `json:"result"`
}

func handleErr(err error) response {
	msg := "success"
	if err != nil {
		msg = "failed"
		log.Printf("err: %v\n", err)
	}
	return response{
		Result: msg,
	}
}

func (s *Server) postsList(c echo.Context) error {
	posts, err := getAllPosts(s.db)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, handleErr(err))
	}

	return c.Render(http.StatusOK, "list", posts)
}

func (s *Server) singlePostGET(c echo.Context) error {
	id := c.Param("id")

	if len(id) == 0 {
		return c.String(http.StatusBadRequest, "empty param id")
	}

	post, err := getPost(s.db, id)
	if err != nil {
		err = fmt.Errorf("getting post by id: %v: %w", id, err)
		return c.JSON(http.StatusInternalServerError, handleErr(err))
	}

	return c.Render(http.StatusOK, "single", post)
}

func (s *Server) singlePostPOST(c echo.Context) error {
	idVal := c.FormValue("id")
	if len(idVal) == 0 {
		return c.JSON(http.StatusInternalServerError, handleErr(errors.New("empty id")))
	}

	id, err := strconv.Atoi(idVal)
	if err != nil {
		err := fmt.Errorf("convert id val from form-value: %v: %w", idVal, err)
		return c.JSON(http.StatusBadRequest, handleErr(err))
	}

	upgPost := Post{
		ID:      string(idVal),
		Title:   c.FormValue("title"),
		Date:    c.FormValue("date"),
		Link:    c.FormValue("link"),
		Comment: c.FormValue("comment"),
	}

	if err = updatePost(s.db, upgPost); err != nil {
		err := fmt.Errorf("edit post by id: %v: %w", id, err)
		return c.JSON(http.StatusInternalServerError, handleErr(err))
	}

	post, err := getPost(s.db, idVal)
	if err != nil {
		err = fmt.Errorf("getting post by id: %v: %w", idVal, err)
		return c.JSON(http.StatusInternalServerError, handleErr(err))
	}

	return c.Render(http.StatusOK, "single", post)
}

func (s *Server) editPost(c echo.Context) error {
	id := c.Param("id")
	if len(id) == 0 {
		return c.String(http.StatusBadRequest, "empty param id")

	}

	post, err := getPost(s.db, id)
	if err != nil {
		err = fmt.Errorf("getting post by id: %v: %w", id, err)
		return c.JSON(http.StatusInternalServerError, handleErr(err))
	}

	return c.Render(http.StatusOK, "edit", post)
}

func (s *Server) createPostGET(c echo.Context) error {
	newPost := Post{}

	return c.Render(http.StatusOK, "create", newPost)
}

func (s *Server) createPostPOST(c echo.Context) error {
	newPost := Post{
		ID:      c.FormValue("id"),
		Title:   c.FormValue("title"),
		Date:    c.FormValue("date"),
		Link:    c.FormValue("link"),
		Comment: c.FormValue("comment"),
	}

	// проверка на наличие поста с таким-же ID в БД. check := Post{}
	check, err := getPost(s.db, newPost.ID)
	if err != nil {
		log.Printf("err: %v\n", err)
	}

	if check.ID == newPost.ID {
		return c.HTML(http.StatusOK, "post with the specified id already exists")
	}

	if err := createPost(s.db, newPost); err != nil {
		err := fmt.Errorf("create post ERROR: %w", err)
		return c.JSON(http.StatusInternalServerError, handleErr(err))
	}

	return c.Redirect(http.StatusFound, "/")
}

func (s *Server) deletePostGET(c echo.Context) error {
	id := c.Param("id")
	if len(id) == 0 {
		return c.String(http.StatusBadRequest, "empty param id")
	}

	if err := deletePost(s.db, id); err != nil {
		err := fmt.Errorf("delete post ERROR: %w", err)
		return c.JSON(http.StatusInternalServerError, handleErr(err))
	}

	return c.Redirect(http.StatusFound, "/")
}
