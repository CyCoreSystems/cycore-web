package main

//go:generate esc -o static.go -prefix assets -ignore \.map$ assets

import (
	"flag"
	"html/template"
	"net/http"

	"github.com/CyCoreSystems/cycore-web/db"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"go.uber.org/zap"
)

var addr string
var debug bool

// Error indicates an error from processing
type Error struct {
	Message string `json:"message"`
}

// NewError converts a standard error to an error response
func NewError(err error) *Error {
	return &Error{
		Message: err.Error(),
	}
}

func init() {
	flag.StringVar(&addr, "addr", ":9000", "listen address")
	flag.BoolVar(&debug, "debug", false, "run with debug logging")
}

func main() {

	flag.Parse()

	var err error

	var logger *zap.Logger
	if debug {
		logger, err = zap.NewDevelopment()
	} else {
		logger, err = zap.NewProduction()
	}
	if err != nil {
		panic("failed to create logger: " + err.Error())
	}
	defer logger.Sync() // nolint

	log := logger.Sugar()

	err = db.Connect()
	if err != nil {
		log.Panicf("failed to open database: %v", err)
	}
	defer db.Get().Close() // nolint

	e := echo.New()
	//e.Use(middleware.CSRF())
	e.Use(middleware.Gzip())
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	//e.Use(middleware.Secure())

	// Create custom context
	e.Use(func(h echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			cc := &Context{
				Context: c,
				DB:      db.Get(),
				Log:     log,
			}
			return h(cc)
		}
	})

	e.Renderer = &Template{
		templates: template.Must(template.ParseGlob("views/*.html")),
	}

	assets := http.FileServer(FS(debug))

	e.GET("/css/*", echo.WrapHandler(assets))
	e.GET("/img/*", echo.WrapHandler(assets))
	e.GET("/js/*", echo.WrapHandler(assets))
	e.GET("/scm.asc", echo.WrapHandler(assets))

	e.GET("/", home)

	e.POST("/contact/request", contactRequest)

	log.Fatal(e.Start(addr))
}

func home(c echo.Context) error {
	return c.Render(http.StatusOK, "index.html", nil)
}
