package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/chickenzord/go-huawei-client/pkg/hn8010ts"
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog/log"
)

type Config struct {
	BindHost string `envconfig:"bind_host" default:"0.0.0.0"`
	BindPort int    `envconfig:"bind_port" default:"8080"`
	Router   struct {
		URL      string `envconfig:"url"`
		Username string `envconfig:"username"`
		Password string `envconfig:"password"`
	} `envconfig:"router"`
}

func (c *Config) Bind() string {
	return fmt.Sprintf("%s:%d", c.BindHost, c.BindPort)
}

func main() {
	_ = godotenv.Overload(".env")

	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		panic(err)
	}

	r := prometheus.NewRegistry()
	r.MustRegister(NewRouterCollector(&hn8010ts.Config{
		URL:      cfg.Router.URL,
		Username: cfg.Router.Username,
		Password: cfg.Router.Password,
	}))

	e := echo.New()
	e.HideBanner = true
	e.Pre(middleware.RemoveTrailingSlash())
	e.Use(middleware.Recover())
	e.Use(middleware.Logger())
	e.GET("/ping", func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})
	e.GET("/metrics", echo.WrapHandler(promhttp.HandlerFor(r, promhttp.HandlerOpts{
		MaxRequestsInFlight: 2,
		EnableOpenMetrics:   true,
	})))

	log.Info().
		Str("bind", cfg.Bind()).
		Str("url", cfg.Router.URL).
		Str("user", cfg.Router.Username).
		Msg("starting metrics exporter server")

	if err := e.Start(cfg.Bind()); err != nil {
		fmt.Println()
		fmt.Println(err)
		os.Exit(1)
	}
}
