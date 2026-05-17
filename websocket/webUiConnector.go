package websocket

import (
	"centauri-carbon-proxy/logging"
	"centauri-carbon-proxy/types"
	"context"
	"net/http"
	"sync"

	"github.com/coder/websocket"
)

type WebUiConnector struct {
	ctx       context.Context
	logger    logging.Logger
	config    types.Config
	clients   map[chan []byte]any
	clientsMu sync.Mutex
	readChan  chan []byte
	writeChan chan []byte
	Routes    types.RouteMap
}

func NewWebUiConnector(ctx context.Context, logger logging.Logger, config types.Config, readChan chan []byte, writeChan chan []byte) *WebUiConnector {
	server := WebUiConnector{
		ctx:       ctx,
		logger:    logger,
		config:    config,
		readChan:  readChan,
		writeChan: writeChan,
		clients:   make(map[chan []byte]any),
	}

	server.Routes = types.RouteMap{"GET /websocket": server.SubscribeHandler}

	return &server
}

func (wc *WebUiConnector) Start() {
	go wc.writeLoop()
	go wc.startHttpServer()
}

func (wc *WebUiConnector) startHttpServer() {
	serveMux := http.ServeMux{}

	for route, handler := range wc.Routes {
		serveMux.HandleFunc(route, handler)
	}

	httpServer := &http.Server{
		Addr:           ":3030",
		Handler:        &serveMux,
		ReadTimeout:    0,
		WriteTimeout:   0,
		MaxHeaderBytes: 1 << 20,
	}
	wc.logger.Fatal(httpServer.ListenAndServe())
}

func (wc *WebUiConnector) SubscribeHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithCancel(context.Background())
	msgChan := make(chan []byte)
	wc.clientsMu.Lock()
	wc.clients[msgChan] = struct{}{}
	wc.clientsMu.Unlock()
	defer func() {
		wc.clientsMu.Lock()
		delete(wc.clients, msgChan)
		wc.clientsMu.Unlock()
	}()

	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		InsecureSkipVerify: true,
	})
	if err != nil {
		wc.logger.Errorf("Error connecting client websocket: %s", err.Error())
	}

	go wc.readLoop(ctx, cancel, conn)

	for {
		select {
		case data := <-msgChan:
			err = conn.Write(ctx, websocket.MessageText, data)
			if err != nil {
				wc.logger.Error(err.Error())
			}
		case <-ctx.Done():
			return
		}
	}

}

func (wc *WebUiConnector) readLoop(ctx context.Context, cancel context.CancelFunc, conn *websocket.Conn) {
	for {
		wc.logger.Debug("Reading from webUi")
		_, data, err := conn.Read(ctx)
		if websocket.CloseStatus(err) == websocket.StatusNormalClosure || websocket.CloseStatus(err) == websocket.StatusGoingAway {
			cancel()
			break
		}
		if err != nil {
			wc.logger.Errorf("Error reading from the WebUI: %s", err.Error())
			cancel()
			conn.CloseNow()
			break
		}

		wc.writeChan <- data
	}
}

func (wc *WebUiConnector) writeLoop() {
	for {
		data, open := <-wc.readChan
		if !open {
			return
		}

		for client := range wc.clients {
			client <- data
		}
	}
}
