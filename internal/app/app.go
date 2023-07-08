package app

import (
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/ilyabukanov123/api-mail/internal/config"
	"github.com/ilyabukanov123/api-mail/internal/http-server/handlers"
	"github.com/ilyabukanov123/api-mail/internal/lib/wpsev"
)

func Start(configPath string) {
	app := config.NewConfig(configPath)
	handler := handlers.New(*app)

	middleware := func(w http.ResponseWriter, r *http.Request) {

	}

	hs := &http.Server{}

	myServer := wpsev.NewServer(hs, wpsev.HTTP3)

	myServer.AddRouter(http.MethodPost, "/*username", middleware, handler.NewUsernameEmail)
	myServer.AddRouter(http.MethodGet, "/get", middleware, handler.GetUsername)
	myServer.AddRouter(http.MethodGet, "/get/*link", middleware, handler.GetArchiveUsername)

	go func() {
		err := myServer.Start(app.Config.Addr, app.Config.Port)
		if err != nil {
			panic(err)
		}
	}()

	osSignalsCh := make(chan os.Signal, 1)
	signal.Notify(osSignalsCh, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)

	<-osSignalsCh

	err := myServer.Stop()
	if err != nil {
		panic(err)
	}
}
