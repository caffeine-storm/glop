package cachetest

import (
	"fmt"

	"github.com/runningwild/glop/cache"
)

type SpyedCache struct {
	Ops []string
	Impl cache.ByteBank
}

var _ cache.ByteBank = (*SpyedCache)(nil)

func (spy *SpyedCache) Read(key string) ([]byte, bool, error) {
	spy.Ops = append(spy.Ops, fmt.Sprintf("read %q", key))
	return spy.Impl.Read(key)
}

func (spy *SpyedCache) Write(key string, data []byte) error {
	spy.Ops = append(spy.Ops, fmt.Sprintf("write %q", key))
	return spy.Impl.Write(key, data)
}

