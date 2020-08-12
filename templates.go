package main

import (
	"html/template"
	"io"

	"github.com/labstack/echo/v4"
)

// Template implements echo.Renderer
type Template struct {
	templates *template.Template
}

// Render executes the selected template with the given data
func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}
