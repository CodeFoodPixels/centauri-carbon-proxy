package types

import (
	"centauri-carbon-proxy/logging"
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
)

type Config struct {
	PrinterIp         string
	PrinterIpBytes    []byte
	PrinterUrlBytes   []byte
	PrinterWsUrlBytes []byte
	AppUrl            string
	AppHostBytes      []byte
	AppUrlBytes       []byte
	AppWsUrlBytes     []byte
	CustomCssBytes    []byte
	Port              int
	UseHttps          bool
}

func NewConfig(logger logging.Logger) (Config, error) {
	useHttps := false
	localIp, err := getLocalIp()
	if err != nil {
		return Config{}, fmt.Errorf("Unable to get local IP: %s", err)
	}

	printerIp := os.Getenv("PROXY_PRINTER_IP")
	if printerIp == "" {
		logger.Fatal("PROXY_PRINTER_IP not set")
	}

	parsedIp := net.ParseIP(printerIp)
	if parsedIp == nil {
		logger.Fatal("invalid value for PROXY_PRINTER_IP")
	}

	port, err := strconv.Atoi(os.Getenv("PROXY_PORT"))
	if err != nil {
		port = 3030
		logger.Infof("PROXY_PORT not set or set to invalid value, defaulting to %d", port)
	}

	var appUrl string
	var appHost string
	var appWsUrl string
	parsedUrl, err := url.Parse(os.Getenv("PROXY_APP_URL"))
	if parsedUrl.Host == "" || parsedUrl.Scheme == "" || err != nil {
		appUrl = fmt.Sprintf("http://%s:%d", localIp, port)
		appWsUrl = fmt.Sprintf("ws://%s:%d", localIp, port)
		appHost = fmt.Sprintf("%s:%d", localIp, port)
		logger.Infof("PROXY_APP_URL invalid or not set, defaulting to %s", appUrl)
	} else {
		appUrl = fmt.Sprintf("%s://%s", parsedUrl.Scheme, parsedUrl.Host)
		appHost = parsedUrl.Host
		if parsedUrl.Scheme == "https" {
			useHttps = true
			appWsUrl = fmt.Sprintf("wss://%s", parsedUrl.Host)
		} else {
			appWsUrl = fmt.Sprintf("ws://%s", parsedUrl.Host)
		}
	}

	customCssPath := os.Getenv("PROXY_CUSTOM_CSS_PATH")
	var customCssBytes []byte
	if customCssPath != "" {
		customCssBytes, err = os.ReadFile(customCssPath)
		if err != nil {
			logger.Errorf("Error loading custom CSS file: %s", err)
		}
	}

	return Config{
		PrinterIp:         printerIp,
		PrinterIpBytes:    []byte(printerIp),
		PrinterUrlBytes:   []byte(fmt.Sprintf("http://%s", printerIp)),
		PrinterWsUrlBytes: []byte(fmt.Sprintf("ws://%s", printerIp)),
		AppUrl:            appUrl,
		AppHostBytes:      []byte(appHost),
		AppUrlBytes:       []byte(appUrl),
		AppWsUrlBytes:     []byte(appWsUrl),
		CustomCssBytes:    customCssBytes,
		Port:              port,
		UseHttps:          useHttps,
	}, nil
}

func getLocalIp() (string, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "", err
	}
	defer conn.Close()

	localAddress := conn.LocalAddr().(*net.UDPAddr)

	return localAddress.IP.String(), nil
}
