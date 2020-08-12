package main

import (
	"flag"
	"html/template"
	"net/http"
	"os"

	"github.com/CyCoreSystems/cycore-web/db"
	"github.com/inconshreveable/log15"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
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

	log := log15.New("app", "cycore-web")
	if os.Getenv("KUBERNETES_SERVICE_HOST") != "" {
		h, err := log15.NetHandler("tcp", "oklog.log", log15.JsonFormat())
		if err != nil {
			log.Error("failed to construct network logger", "error", err)
		} else {
			log.SetHandler(h)
		}
	}

	err := db.Connect()
	if err != nil {
		log.Crit("failed to open database", "error", err)
		os.Exit(1)
	}
	defer db.Get().Close() // nolint

	e := echo.New()
	// e.Use(middleware.CSRF())
	e.Use(middleware.Gzip())
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	// e.Use(middleware.Secure())

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

	e.Static("/css/*", "css")
	e.Static("/js/*", "js")
	e.Static("/img/*", "img")
	e.File("/scm.asc", "public/scm.asc")

	e.GET("/", home)

	e.POST("/contact/request", contactRequest)

	if err = e.Start(addr); err != nil {
		log.Crit(err.Error())
		os.Exit(1)
	}
	os.Exit(0)
}

func home(c echo.Context) error {
	return c.Render(http.StatusOK, "index.html", nil)
}
