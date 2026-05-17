package websocket

import (
	"centauri-carbon-proxy/logging"
	"centauri-carbon-proxy/types"
	"context"
)

func New(ctx context.Context, logger logging.Logger, config types.Config) types.RouteMap {
	fromPrinterChan := make(chan []byte)
	toPrinterChan := make(chan []byte)

	printerConnector := NewPrinterConnector(ctx, logger, config, toPrinterChan, fromPrinterChan)
	webUiConnector := NewWebUiConnector(ctx, logger, config, fromPrinterChan, toPrinterChan)
	go printerConnector.Start()
	webUiConnector.Start()

	return webUiConnector.Routes
}
