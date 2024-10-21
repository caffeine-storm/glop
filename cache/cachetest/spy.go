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
	spy.Ops = append(spy.Ops, spy.ReadOp(key))
	return spy.Impl.Read(key)
}

func (spy *SpyedCache) Write(key string, data []byte) error {
	spy.Ops = append(spy.Ops, spy.WriteOp(key))
	return spy.Impl.Write(key, data)
}

func (*SpyedCache) ReadOp(key string) string {
	return fmt.Sprintf("read %q", key)
}

func (*SpyedCache) WriteOp(key string) string {
	return fmt.Sprintf("write %q", key)
}
