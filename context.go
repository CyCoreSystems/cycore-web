package main

import (
	"github.com/inconshreveable/log15"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo"
)

// Context is the custom context for this web server
type Context struct {
	echo.Context

	// DB is the database connection
	DB *sqlx.DB

	// Log is the core logger
	Log log15.Logger
}
