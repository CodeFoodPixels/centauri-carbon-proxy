package file

import (
	"bytes"
	"centauri-carbon-proxy/logging"
	"centauri-carbon-proxy/transformer"
	"centauri-carbon-proxy/types"
	"context"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strings"
)

type File struct {
	ctx    context.Context
	logger logging.Logger
	config types.Config
	cache  map[string]CachedResponse
}

type CachedResponse struct {
	Header http.Header
	Body   []byte
}

func New(ctx context.Context, logger logging.Logger, config types.Config) types.RouteMap {
	file := File{
		ctx:    ctx,
		logger: logger,
		config: config,
		cache:  make(map[string]CachedResponse),
	}

	return types.RouteMap{
		"GET /{FILE...}":          file.Get,
		"POST /uploadFile/upload": file.Upload,
	}
}

func (f *File) Get(w http.ResponseWriter, r *http.Request) {
	filePath := r.PathValue("FILE")

	cachedResponse, ok := f.cache[filePath]
	if ok {
		for key, val := range cachedResponse.Header {
			w.Header()[key] = val
		}
		w.Write(cachedResponse.Body)
		return
	}

	res, _ := http.Get(fmt.Sprintf("http://%s/%s", f.config.PrinterIp, filePath))

	contentType := strings.Split(res.Header.Get("content-type"), ";")[0]

	for key, val := range res.Header {
		if slices.Contains(ALLOWED_RESPONSE_HEADERS[r.Method], strings.ToLower(key)) {
			w.Header()[key] = val
		}
	}

	if slices.Contains(CACHEABLE_MIME_TYPES, contentType) && w.Header().Get("cache-control") == "" {
		w.Header()["cache-control"] = []string{"public, max-age=31536000"}
	}

	w.Header()["Transfer-Encoding"] = nil
	w.Header()["Date"] = nil
	w.Header()["Content-Length"] = nil

	w.WriteHeader(res.StatusCode)

	if slices.Contains(TRANSFORMABLE_MIME_TYPES, contentType) {
		cacheBuffer := bytes.Buffer{}
		err := transformer.TransformReaderWriter(f.config, w, res.Body, &cacheBuffer)
		if err != nil {
			f.logger.Errorf("Error transforming file for path %s: %s", filePath, err.Error())
		} else {
			f.cache[filePath] = CachedResponse{
				Header: w.Header().Clone(),
				Body:   cacheBuffer.Bytes(),
			}
		}
	} else {
		io.Copy(w, res.Body)
	}

	flusher, ok := w.(http.Flusher)
	if ok {
		flusher.Flush()
	}
}

func (f *File) Upload(w http.ResponseWriter, r *http.Request) {
	f.logger.Debug(r.Header)
	f.logger.Debug(r)
	body, err := io.ReadAll(r.Body)
	if err != nil {
		f.logger.Errorf("Error uploading file: %s", err.Error())
		return
	}
	f.logger.Debug(body)
	res, err := http.Post(fmt.Sprintf("http://%s/uploadFile/upload", f.config.PrinterIp), r.Header.Get("content-type"), bytes.NewReader(body))
	if err != nil {
		f.logger.Errorf("Error uploading file: %s", err.Error())
		return
	}

	for key, val := range res.Header {
		if slices.Contains(ALLOWED_RESPONSE_HEADERS[r.Method], strings.ToLower(key)) {
			w.Header()[key] = val
		}
	}

	w.WriteHeader(res.StatusCode)
	_, err = io.Copy(w, res.Body)
	if err != nil {
		f.logger.Errorf("Error writing response for upload: %s", err.Error())
		return
	}
}
