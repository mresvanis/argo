package util

import (
	"math/rand"
	"time"
	"unsafe"
)

const symbolBytes = "-_0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const (
	symbolIdxBits = 6                    // 6 bits to represent a symbol index
	symbolIdxMask = 1<<symbolIdxBits - 1 // All 1-bits, as many as symbolIdxBits
	symbolIdxMax  = 63 / symbolIdxBits   // # of symbol indices fitting in 63 bits
)

var (
	src = rand.NewSource(time.Now().UnixNano())
)

func GenerateRandomString(n int) string {
	b := make([]byte, n)
	for i, cache, remain := n-1, src.Int63(), symbolIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), symbolIdxMax
		}
		if idx := int(cache & symbolIdxMask); idx < len(symbolBytes) {
			b[i] = symbolBytes[idx]
			i--
		}
		cache >>= symbolIdxBits
		remain--
	}

	return *(*string)(unsafe.Pointer(&b))
}
