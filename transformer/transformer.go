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

type Transformer struct {
	Find    []byte
	Replace []byte
	buf     []byte
}

func (t *Transformer) Transform(src []byte, flush bool) []byte {
	if len(t.Find) == 0 {
		return src
	}

	output := []byte{}
	t.buf = append(t.buf, src...)

	for {
		i := bytes.Index(t.buf, t.Find)
		if i == -1 {
			break
		}
		output = append(output, t.buf[:i]...)
		output = append(output, t.Replace...)
		t.buf = t.buf[i+len(t.Find):]
	}

	tailChars := len(t.buf) - (len(t.Find) - 1)
	if tailChars > 0 {
		output = append(output, t.buf[:tailChars]...)
		t.buf = t.buf[tailChars:]
	}

	if flush {
		output = append(output, t.buf...)
		t.buf = []byte{}
	}
	return output
}

type Chain struct {
	Transformers []Transformer
}

func (c *Chain) Process(src []byte, flush bool) []byte {
	output := src
	for i, transformer := range c.Transformers {
		output = transformer.Transform(output, flush)
		c.Transformers[i] = transformer
	}

	return output
}

func TransformWebsocketFrame(config types.Config, data []byte) []byte {
	transformers := []Transformer{}
	if config.UseHttps {
		transformers = append(transformers, Transformer{Find: httpBytes, Replace: httpsBytes}, Transformer{Find: wsBytes, Replace: wssBytes})
	}

	transformers = append(
		transformers,
		Transformer{Find: websocketPortBytes, Replace: []byte{}},
		Transformer{Find: videoPortBytes, Replace: []byte{}},
		Transformer{Find: config.PrinterUrlBytes, Replace: config.AppUrlBytes},
		Transformer{Find: config.PrinterIpBytes, Replace: config.AppHostBytes},
	)

	chain := Chain{Transformers: transformers}

	return chain.Process(data, true)
}

func TransformReaderWriter(config types.Config, dst io.Writer, src io.Reader, buffer io.Writer) error {
	transformers := []Transformer{}
	if config.UseHttps {
		transformers = append(transformers, Transformer{Find: httpBytes, Replace: httpsBytes}, Transformer{Find: wsBytes, Replace: wssBytes})
	}

	transformers = append(
		transformers,
		Transformer{Find: websocketPortBytes, Replace: []byte{}},
		Transformer{Find: videoPortBytes, Replace: []byte{}},
		Transformer{Find: config.PrinterWsUrlBytes, Replace: config.AppWsUrlBytes},
		Transformer{Find: config.PrinterUrlBytes, Replace: config.AppUrlBytes},
		Transformer{Find: []byte("this.hostName=window.location.hostname"), Replace: []byte("this.hostName=`${window.location.hostname}${window.location.port?`:${window.location.port}`:''}`")},
		Transformer{Find: []byte("${this.webSocketService.hostName}:80"), Replace: []byte("${this.webSocketService.hostName}")},
		Transformer{Find: []byte("</body>"), Replace: []byte(fmt.Sprintf("<script>if(window.location.origin !== '%s'){window.location.href=window.location.href.replace(window.location.origin, '%s')}</script></body>", config.AppUrl, config.AppUrl))},
	)

	if len(config.CustomCssBytes) > 0 {
		transformers = append(transformers, Transformer{Find: []byte("</body>"), Replace: bytes.Join([][]byte{[]byte("<style>"), config.CustomCssBytes, []byte("</style></body>")}, []byte{})})
	}

	chain := Chain{Transformers: transformers}

	readBufSize := 32 * 1024
	readBuf := make([]byte, readBufSize)
	isEOF := false
	for {
		n, err := src.Read(readBuf)
		if err != nil && err != io.EOF {
			return err
		}

		isEOF = err == io.EOF

		if isEOF && n == 0 {
			return nil
		}

		data := chain.Process(readBuf[:n], isEOF)

		_, err = dst.Write(data)
		if err != nil {
			return err
		}
		_, err = buffer.Write(data)
		if err != nil {
			return err
		}
	}
}
