package websocket

import (
	"centauri-carbon-proxy/logging"
	"centauri-carbon-proxy/transformer"
	"centauri-carbon-proxy/types"
	"context"
	"fmt"
	"time"

	"github.com/coder/websocket"
)

type PrinterConnector struct {
	ctx        context.Context
	logger     logging.Logger
	config     types.Config
	connection *websocket.Conn
	cancel     context.CancelFunc
	readChan   chan []byte
	writeChan  chan []byte
}

func NewPrinterConnector(ctx context.Context, logger logging.Logger, config types.Config, readChan chan []byte, writeChan chan []byte) PrinterConnector {
	return PrinterConnector{
		ctx:       ctx,
		logger:    logger,
		config:    config,
		readChan:  readChan,
		writeChan: writeChan,
	}
}

func (pc *PrinterConnector) Start() {
	for {
		pc.runLoop()
		time.Sleep(time.Second * 30)
	}
}

func (pc *PrinterConnector) runLoop() {
	ctx, cancel := context.WithCancel(pc.ctx)
	pc.cancel = cancel

	printerAddr := fmt.Sprintf("ws://%s:3030/websocket", pc.config.PrinterIp)

	pc.logger.Debugf("Connecting to printer on %s", printerAddr)

	connection, _, err := websocket.Dial(pc.ctx, printerAddr, nil)
	if err != nil {
		pc.logger.Errorf("Error connecting to printer: %s", err.Error())
		return
	}
	connection.SetReadLimit(-1)

	pc.connection = connection
	go pc.readLoop(ctx)
	go pc.writeLoop(ctx)

	<-ctx.Done()
	pc.Close()
}

func (pc *PrinterConnector) readLoop(ctx context.Context) {
	for {
		pc.logger.Debug("Reading from printer")
		_, data, err := pc.connection.Read(ctx)
		if err != nil {
			pc.logger.Errorf("Error reading from the printer: %s", err.Error())
			pc.cancel()
			return
		}

		pc.writeChan <- transformer.TransformWebsocketFrame(pc.config, data)
	}
}

func (pc *PrinterConnector) writeLoop(ctx context.Context) {
	tickerInterval := time.Second * 30
	t := time.NewTicker(tickerInterval)
	for {
		select {
		case data := <-pc.readChan:
			err := pc.connection.Write(ctx, websocket.MessageText, data)
			if err != nil {
				pc.logger.Errorf("Error writing to the printer: %s", err.Error())
				t.Stop()
				pc.cancel()
				return
			}
			t.Reset(tickerInterval)
			pc.logger.Debug("Wrote to printer")
		case <-t.C:
			pc.logger.Debug("Sending ping to printer")
			err := pc.connection.Write(ctx, websocket.MessageText, []byte("ping"))
			if err != nil {
				pc.logger.Errorf("Error sending ping to the printer: %s", err.Error())
				t.Stop()
				pc.cancel()
				return
			}
		case <-ctx.Done():
			t.Stop()
			pc.cancel()
			return
		}
	}
}

func (pc *PrinterConnector) Close() {
	pc.connection.CloseNow()
	pc.cancel()
}
