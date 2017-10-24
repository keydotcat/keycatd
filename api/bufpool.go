package api

import (
	"bytes"
	"sync"
)

type bufferPool struct {
	p sync.Pool
}

func (bp bufferPool) Get() *bytes.Buffer {
	b := bp.p.Get().(*bytes.Buffer)
	return b
}

func (bp bufferPool) Put(b *bytes.Buffer) {
	b.Reset()
	bp.p.Put(b)
}

var bufPool = bufferPool{
	sync.Pool{
		New: func() interface{} {
			return bytes.NewBuffer(nil)
		},
	},
}
