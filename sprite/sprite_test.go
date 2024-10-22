package sprite_test

import (
	"fmt"

	"github.com/orfjackal/gospec/src/gospec"
	. "github.com/orfjackal/gospec/src/gospec"
	"github.com/runningwild/glop/cache"
	"github.com/runningwild/glop/cache/cachetest"
	"github.com/runningwild/glop/render/rendertest"
	"github.com/runningwild/glop/sprite"
)

func loadSprites(sprite_paths ...string) ([]*sprite.Sprite, error) {
	discardQueue := rendertest.MakeDiscardingRenderQueue()
	spriteMan := sprite.MakeManager(discardQueue, func(path string) cache.ByteBank {
		return cache.MakeRamByteBank()
	})

	result := []*sprite.Sprite{}

	for _, spritePath := range sprite_paths {
		s, err := spriteMan.LoadSprite(spritePath)
		if err != nil {
			return nil, fmt.Errorf("couldn't load sprite at path %q: %v", spritePath, err)
		}
		result = append(result, s)
	}

	return result, nil
}

func LoadSpriteSpec(c gospec.Context) {
	c.Specify("Sample sprite loads correctly", func() {
		sList, err := loadSprites("test_sprite")
		c.Expect(err, Equals, nil)
		s := sList[0]
		for i := 0; i < 2000; i++ {
			s.Think(50)
		}
		s.Command("defend")
		s.Command("undamaged")
		s.Command("defend")
		s.Command("undamaged")
		for i := 0; i < 3000; i++ {
			s.Think(50)
		}
		s.Command("turn_right")
		s.Command("turn_right")
		s.Command("turn_right")
		s.Command("turn_right")
		s.Command("turn_right")
		s.Command("turn_right")
		s.Command("turn_left")
		s.Command("turn_left")
		s.Command("turn_right")
		s.Command("turn_right")
		s.Command("turn_right")
		s.Command("turn_left")
		s.Command("turn_left")
		// s.Think(5000)
		for i := 0; i < 300; i++ {
			s.Think(50)
		}
		c.Expect(s.Facing(), Equals, 1)
	})
}

func CommandNSpec(c gospec.Context) {
	c.Specify("Sample sprite loads correctly", func() {
		sList, err := loadSprites("test_sprite")
		c.Expect(err, Equals, nil)
		s := sList[0]
		for i := 0; i < 2000; i++ {
			s.Think(50)
		}
		s.CommandN([]string{
			"turn_right",
			"turn_right",
			"turn_right",
			"turn_right",
			"turn_right",
			"turn_right",
			"turn_left",
			"turn_left",
			"turn_right",
			"turn_right",
			"turn_right",
			"turn_left",
			"turn_left"})
		s.Think(5000)
		c.Expect(s.Facing(), Equals, 1)
		for i := 0; i < 3000; i++ {
			s.Think(50)
		}
	})
}

func SyncSpec(c gospec.Context) {
	c.Specify("Sample sprite loads correctly", func() {
		sList, err := loadSprites("test_sprite", "test_sprite")
		c.Expect(err, Equals, nil)
		s1, s2 := sList[0], sList[1]
		sprite.CommandSync([]*sprite.Sprite{s1, s2}, [][]string{[]string{"melee"}, []string{"defend", "damaged"}}, "hit")
		hit := false
		for i := 0; i < 20; i++ {
			s1.Think(50)
			s2.Think(50)
			if s1.Anim() == "melee_01" && s2.Anim() == "damaged_01" {
				hit = true
			}
		}
		c.Expect(hit, Equals, true)
	})
}

func ManagerSpec(c gospec.Context) {
	c.Specify("A manager must take a cacheing dependency", func() {
		discardQueue := rendertest.MakeDiscardingRenderQueue()
		spyCache := &cachetest.SpyedCache{
			Impl: cache.MakeRamByteBank(),
		}
		spyCacheFactory := func(s string) cache.ByteBank {
			return spyCache
		}
		_ = sprite.MakeManager(discardQueue, spyCacheFactory)
	})
}
