package main

import (
	"centauri-carbon-proxy/file"
	"centauri-carbon-proxy/logging"
	"centauri-carbon-proxy/types"
	"centauri-carbon-proxy/video"
	"centauri-carbon-proxy/websocket"
	"context"
	"fmt"
	"net/http"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	logger := logging.NewLogger()

	config, err := types.NewConfigFromEnv(logger)
	if err != nil {
		logger.Fatal(err)
	}

	serveMux := http.ServeMux{}

	websocketRoutes := websocket.New(ctx, logger, config)
	addRoutes(&serveMux, websocketRoutes)

	videoRoutes := video.New(ctx, logger, config)
	addRoutes(&serveMux, videoRoutes)

	fileRoutes := file.New(ctx, logger, config)
	addRoutes(&serveMux, fileRoutes)

	httpServer := &http.Server{
		Addr:           fmt.Sprintf(":%d", config.Port),
		Handler:        &serveMux,
		ReadTimeout:    0,
		WriteTimeout:   0,
		MaxHeaderBytes: 1 << 20,
	}
	logger.Infof("Starting web server on port %d", config.Port)
	logger.Fatal(httpServer.ListenAndServe())
}

func addRoutes(mux *http.ServeMux, routes types.RouteMap) {
	for route, handler := range routes {
		mux.HandleFunc(route, handler)
	}
}
