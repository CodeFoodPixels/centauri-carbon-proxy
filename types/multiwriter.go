package types

import (
	"io"
	"sync"
)

type Multiwriter struct {
	writers   map[io.Writer]any
	writersMu sync.Mutex
}

func NewMultiwriter() *Multiwriter {
	return &Multiwriter{
		writers: make(map[io.Writer]any),
	}
}

func (mw *Multiwriter) Add(writer io.Writer) func() {
	mw.writersMu.Lock()
	mw.writers[writer] = struct{}{}
	mw.writersMu.Unlock()

	return func() {
		mw.writersMu.Lock()
		delete(mw.writers, writer)
		mw.writersMu.Unlock()
	}
}

func (mw *Multiwriter) Write(p []byte) (int, error) {
	mw.writersMu.Lock()
	defer mw.writersMu.Unlock()
	for writer := range mw.writers {
		n, err := writer.Write(p)
		if err != nil {
			return n, err
		}
		if n != len(p) {
			err = io.ErrShortWrite
			return n, err
		}
	}

	return len(p), nil
}
