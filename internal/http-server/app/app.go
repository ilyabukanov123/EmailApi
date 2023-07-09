package app

import (
	"github.com/ilyabukanov123/api-mail/internal/config"
	"github.com/ilyabukanov123/api-mail/internal/http-server/handlers"
	"github.com/ilyabukanov123/api-mail/internal/lib/wpsev"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// Launching the email application
func Start(configPath string) {
	// Creating an application instance and handlers
	app := config.NewConfig(configPath)
	handler := handlers.New(*app)

	middleware := func(w http.ResponseWriter, r *http.Request) {

	}

	//Creating a server instance
	hs := &http.Server{}
	myServer := wpsev.NewServer(hs, wpsev.HTTP3)

	// Registration of Roots and Handlers for Processing
	myServer.AddRouter(http.MethodPost, "/*username", middleware, handler.NewUsernameEmail)
	myServer.AddRouter(http.MethodGet, "/get", middleware, handler.GetUsername)
	myServer.AddRouter(http.MethodGet, "/get/*link", middleware, handler.GetArchiveUsername)

	// Calling the method to clear the mappa elements with an expired ttl
	handler.StartCleanup(app.Config.CleaningTime * time.Second)

	// Running the server
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
