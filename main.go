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
	"os"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	logger := logging.NewLogger()
	printerIp := os.Getenv("PRINTER_IP")
	if printerIp == "" {
		logger.Fatal("PRINTER_IP not set")
	}

	config, err := types.NewConfig(printerIp, os.Getenv("PROXY_HOSTNAME"), os.Getenv("PROXY_PORT"))
	if err != nil {
		logger.Fatal(err)
	}
	logger.Debugf("Local IP: %s", config.LocalIp)

	serveMux := http.ServeMux{}

	websocketRoutes := websocket.New(ctx, logger, config)
	addRoutes(&serveMux, websocketRoutes)

	videoRoutes := video.New(ctx, logger, config)
	addRoutes(&serveMux, videoRoutes)

	fileRoutes := file.New(ctx, logger, config)
	addRoutes(&serveMux, fileRoutes)

	httpServer := &http.Server{
		Addr:           fmt.Sprintf(":%s", config.Port),
		Handler:        &serveMux,
		ReadTimeout:    0,
		WriteTimeout:   0,
		MaxHeaderBytes: 1 << 20,
	}
	logger.Infof("Starting web server on port %s", config.Port)
	logger.Fatal(httpServer.ListenAndServe())
}

func addRoutes(mux *http.ServeMux, routes types.RouteMap) {
	for route, handler := range routes {
		mux.HandleFunc(route, handler)
	}
}

func getEnv(key string, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
