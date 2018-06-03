package x

import (
	"math"
	"sync"
)

var pool = &sync.Pool{
	New: func() interface{} {
		return &Helper{
			buf: make([]byte, math.MaxUint16),
		}
	},
}

type Helper struct {
	buf []byte
}

func (h *Helper) Decompress() {

}
