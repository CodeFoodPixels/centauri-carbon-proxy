package transformer

import (
	"bytes"
	"centauri-carbon-proxy/types"
	"io"
)

var (
	httpBytes          = []byte("http://")
	doubleSlashBytes   = []byte("//")
	websocketPortBytes = []byte(":3030")
	videoPortBytes     = []byte(":3031")
)

type Replacement struct {
	Find    []byte
	Replace []byte
}

func TransformWebsocketFrame(config types.Config, data []byte) []byte {
	replacements := []Replacement{
		{httpBytes, doubleSlashBytes},
		{config.PrinterIpBytes, config.HostnameBytes},
		{websocketPortBytes, []byte{}},
		{videoPortBytes, []byte{}},
	}

	for _, replacement := range replacements {
		data = bytes.ReplaceAll(data, replacement.Find, replacement.Replace)
	}

	return data
}

func TransformReaderWriter(config types.Config, dst io.Writer, src io.Reader, buffer io.Writer) error {
	replacements := []Replacement{
		{httpBytes, doubleSlashBytes},
		{config.PrinterIpBytes, config.HostnameBytes},
		{websocketPortBytes, []byte{}},
		{videoPortBytes, []byte{}},
		{[]byte("${this.webSocketService.hostName}:80"), []byte("${this.webSocketService.hostName}")},
	}

	buf := []byte{}
	readBufSize := 32 * 1024
	readBuf := make([]byte, readBufSize)
	isEOF := false
	for {
		n, err := src.Read(readBuf)
		if err != nil && err != io.EOF {
			return err
		}

		isEOF = err == io.EOF

		buf = append(buf, readBuf[:n]...)

		for _, replacement := range replacements {
			matchCount := bytes.Count(buf, replacement.Find)
			if matchCount > 0 {
				buf = bytes.ReplaceAll(buf, replacement.Find, replacement.Replace)
			}
		}

		splitPos := len(buf) - n

		_, err = dst.Write(buf[:splitPos])
		if err != nil {
			return err
		}
		_, err = buffer.Write(buf[:splitPos])
		if err != nil {
			return err
		}
		buf = buf[splitPos:]

		if isEOF && n == 0 {
			return nil
		}
	}
}
