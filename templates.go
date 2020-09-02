package main

import (
	"io"
	"text/template"

	"github.com/labstack/echo/v4"
)

// TemplateRenderer - генерирование шаблона
type TemplateRenderer struct {
	templates *template.Template
}

// Render - рендерим конкретный темплейт
func (t *TemplateRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}
