package main

import (
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo"
	"go.uber.org/zap"
)

// Context is the custom context for this web server
type Context struct {
	echo.Context

	// DB is the database connection
	DB *sqlx.DB

	// Log is the core logger
	Log *zap.SugaredLogger
}
