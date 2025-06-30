package sprite_test

import (
	"fmt"
	"testing"

	"github.com/runningwild/glop/cache"
	"github.com/runningwild/glop/render"
	"github.com/runningwild/glop/render/rendertest/testbuilder"
	"github.com/runningwild/glop/sprite"
)

func givenASpriteManager(queue render.RenderQueueInterface) *sprite.Manager {
	return sprite.MakeManager(queue, func(path string) cache.ByteBank {
		return cache.MakeLockingByteBank(cache.MakeRamByteBank())
	})
}

func TestManagerLoadSprite(t *testing.T) {
	testbuilder.Run(func(queue render.RenderQueueInterface) {
		manager := givenASpriteManager(queue)
		sheet, err := manager.LoadSprite("test_sprite")
		if err != nil {
			panic(fmt.Errorf("couldn't LoadSprite(test_sprite): %w", err))
		}

		if sheet == nil {
			t.Fatalf("got a nil sheet back")
		}

		// Make sure to purge the queue to run any queued jobs that could be
		// broken.
		queue.Purge()
	})
}
