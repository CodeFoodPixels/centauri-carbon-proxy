package types

import (
	"fmt"
	"net"
)

type Config struct {
	PrinterIp      string
	PrinterIpBytes []byte
	Hostname       string
	HostnameBytes  []byte
	LocalIp        string
	LocalIpBytes   []byte
	Port           string
}

func NewConfig(printerIp string, hostname string, port string) (Config, error) {
	localIp, err := getLocalIp()
	if err != nil {
		return Config{}, fmt.Errorf("Unable to get local IP: %s", err)
	}

	if hostname == "" {
		hostname = localIp
	}

	if port == "" {
		port = "3030"
	}

	return Config{
		PrinterIp:      printerIp,
		PrinterIpBytes: []byte(printerIp),
		Hostname:       hostname,
		HostnameBytes:  []byte(hostname),
		LocalIp:        localIp,
		LocalIpBytes:   []byte(localIp),
		Port:           port,
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
