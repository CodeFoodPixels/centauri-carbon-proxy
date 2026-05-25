package transformer

import (
	"bytes"
	"centauri-carbon-proxy/types"
	"fmt"
	"io"
)

var (
	httpBytes          = []byte("http://")
	httpsBytes         = []byte("https://")
	wsBytes            = []byte("ws://")
	wssBytes           = []byte("wss://")
	websocketPortBytes = []byte(":3030")
	videoPortBytes     = []byte(":3031")
)

type Replacement struct {
	Find    []byte
	Replace []byte
}

func TransformWebsocketFrame(config types.Config, data []byte) []byte {
	replacements := []Replacement{}
	if config.UseHttps {
		replacements = append(replacements, Replacement{httpBytes, httpsBytes}, Replacement{wsBytes, wssBytes})
	}
	replacements = append(
		replacements,
		Replacement{config.PrinterUrlBytes, config.AppUrlBytes},
		Replacement{config.PrinterIpBytes, config.AppHostBytes},
		Replacement{websocketPortBytes, []byte{}},
		Replacement{videoPortBytes, []byte{}},
	)

	for _, replacement := range replacements {
		data = bytes.ReplaceAll(data, replacement.Find, replacement.Replace)
	}

	return data
}

func TransformReaderWriter(config types.Config, dst io.Writer, src io.Reader, buffer io.Writer) error {
	replacements := []Replacement{}
	if config.UseHttps {
		replacements = append(replacements, Replacement{httpBytes, httpsBytes}, Replacement{wsBytes, wssBytes})
	}

	replacements = append(
		replacements,
		Replacement{config.PrinterUrlBytes, config.AppUrlBytes},
		Replacement{config.PrinterWsUrlBytes, config.AppWsUrlBytes},
		Replacement{websocketPortBytes, []byte{}},
		Replacement{videoPortBytes, []byte{}},
		Replacement{[]byte("this.hostName=window.location.hostname"), []byte("this.hostName=`${window.location.hostname}${window.location.port?`:${window.location.port}`:''}`")},
		Replacement{[]byte("${this.webSocketService.hostName}:80"), []byte("${this.webSocketService.hostName}")},
		Replacement{[]byte("</body>"), []byte(fmt.Sprintf("<script>if(window.location.origin !== '%s'){window.location.href=window.location.href.replace(window.location.origin, '%s')}</script></body>", config.AppUrl, config.AppUrl))},
	)

	if len(config.CustomCssBytes) > 0 {
		replacements = append(replacements, Replacement{[]byte("</body>"), bytes.Join([][]byte{[]byte("<style>"), config.CustomCssBytes, []byte("</style></body>")}, []byte{})})
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
