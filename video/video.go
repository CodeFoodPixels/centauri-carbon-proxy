package video

import (
	"centauri-carbon-proxy/logging"
	"centauri-carbon-proxy/types"
	"context"
	"fmt"
	"io"
	"net/http"
)

type Video struct {
	ctx             context.Context
	logger          logging.Logger
	config          types.Config
	clientWriter    *types.Multiwriter
	responseHeaders http.Header
	Routes          types.RouteMap
}

func New(ctx context.Context, logger logging.Logger, config types.Config) types.RouteMap {
	video := Video{
		ctx:          ctx,
		logger:       logger,
		config:       config,
		clientWriter: types.NewMultiwriter(),
	}

	video.Start()

	return types.RouteMap{"GET /video": video.SubscribeHandler}
}

func (v *Video) Start() {
	v.logger.Debug("Starting video server")
	go v.readLoop()
}

func (v *Video) SubscribeHandler(w http.ResponseWriter, r *http.Request) {
	for key := range v.responseHeaders {
		for _, val := range v.responseHeaders[key] {
			w.Header().Add(key, val)
		}
	}
	w.Header()["Transfer-Encoding"] = nil

	remove := v.clientWriter.Add(w)
	<-r.Context().Done()
	remove()
}

func (v *Video) readLoop() {
	for {
		res, err := http.Get(fmt.Sprintf("http://%s:3031/video", v.config.PrinterIp))
		if err != nil {
			v.logger.Errorf("Error reading video stream from printer: %s", err.Error())
		}

		v.responseHeaders = res.Header.Clone()

		_, err = io.Copy(v.clientWriter, res.Body)
		if err != nil {
			v.logger.Errorf("Error writing video stream to clients: %s", err.Error())
		}
		res.Body.Close()
	}

}
