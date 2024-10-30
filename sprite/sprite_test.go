package sprite_test

import (
	"fmt"
	"testing"

	"github.com/runningwild/glop/cache"
	"github.com/runningwild/glop/cache/cachetest"
	"github.com/runningwild/glop/render/rendertest"
	"github.com/runningwild/glop/sprite"

	. "github.com/smartystreets/goconvey/convey"
)

func TestSpriteSpecs(t *testing.T) {
	Convey("LoadSpriteSpec", t, LoadSpriteSpec)
	Convey("CommandNSpec", t, CommandNSpec)
	Convey("SyncSpec", t, SyncSpec)
	Convey("ManagerSpec", t, ManagerSpec)
}

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

func LoadSpriteSpec() {
	Convey("Sample sprite loads correctly", func() {
		sList, err := loadSprites("test_sprite")
		So(err, ShouldEqual, nil)
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
		So(s.Facing(), ShouldEqual, 1)
	})
}

func CommandNSpec() {
	Convey("Sample sprite loads correctly", func() {
		sList, err := loadSprites("test_sprite")
		So(err, ShouldEqual, nil)
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
		So(s.Facing(), ShouldEqual, 1)
		for i := 0; i < 3000; i++ {
			s.Think(50)
		}
	})
}

func SyncSpec() {
	Convey("Sample sprite loads correctly", func() {
		sList, err := loadSprites("test_sprite", "test_sprite")
		So(err, ShouldEqual, nil)
		s1, s2 := sList[0], sList[1]
		sprite.CommandSync([]*sprite.Sprite{s1, s2}, [][]string{{"melee"}, {"defend", "damaged"}}, "hit")
		hit := false
		for i := 0; i < 20; i++ {
			s1.Think(50)
			s2.Think(50)
			if s1.Anim() == "melee_01" && s2.Anim() == "damaged_01" {
				hit = true
			}
		}
		So(hit, ShouldEqual, true)
	})
}

func ManagerSpec() {
	Convey("A manager must take a cacheing dependency", func() {
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
