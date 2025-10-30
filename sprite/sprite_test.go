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

func TestSprites(t *testing.T) {
	Convey("Sprites", t, func() {
		Convey("LoadSpriteSpec", LoadSpriteSpec)
		Convey("CommandNSpec", CommandNSpec)
		Convey("SyncSpec", SyncSpec)
		Convey("ManagerSpec", ManagerSpec)
	})
}

func loadSprites(sprite_paths ...string) ([]*sprite.Sprite, error) {
	discardQueue := rendertest.MakeStubbedRenderQueue()
	spriteMan := givenASpriteManager(discardQueue)

	result := []*sprite.Sprite{}

	for _, spritePath := range sprite_paths {
		s, err := spriteMan.LoadSprite(spritePath)
		if err != nil {
			return nil, fmt.Errorf("couldn't load sprite at path %q: %w", spritePath, err)
		}
		result = append(result, s)
	}

	return result, nil
}

func LoadSpriteSpec() {
	Convey("Sample sprite", func() {
		Convey("loads correctly", func() {
			sList, err := loadSprites("test_sprite")
			So(err, ShouldEqual, nil)

			Convey("can animate with \"Command\"", func() {
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
		})
	})
}

func CommandNSpec() {
	Convey("CommandN can run a sequence of animations", func() {
		sList, err := loadSprites("test_sprite")
		So(err, ShouldEqual, nil)
		s := sList[0]

		// Simulate 100 seconds of idle time. Nothing should have changed.
		initialFacing := s.Facing()
		for i := 0; i < 2000; i++ {
			s.Think(50)
			So(s.Facing(), ShouldEqual, initialFacing)
		}

		// Tell the sprite to do a little spinning dance.
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
			"turn_left",
		})

		// Let the sprite finish its dance by simulating 5 seconds of animation.
		s.Think(5000)

		// In the end, each transition is equivalent to changing the facing by +1
		// modulo 2. There's an odd number of transitions so we should end up at 1
		// mod 2.
		So(s.Facing(), ShouldEqual, 1)

		// For the next 150s, verify the sprite isn't animating.
		for i := 0; i < 3000; i++ {
			s.Think(50)
			So(s.Facing(), ShouldEqual, 1)
		}
	})
}

func SyncSpec() {
	Convey("sprites will animate in lock-step with a CommandSync", func() {
		sList, err := loadSprites("test_sprite", "test_sprite")
		So(err, ShouldEqual, nil)
		s1, s2 := sList[0], sList[1]
		sprite.CommandSync([]*sprite.Sprite{s1, s2}, [][]string{{"melee"}, {"defend", "damaged"}}, "hit")
		hit := false
		for i := 0; i < 20; i++ {
			s1.Think(50)
			s2.Think(50)
			// Since we check both animation states, this test will only pass if the
			// sprites are animating at the same rate.
			if s1.Anim() == "melee_01" && s2.Anim() == "damaged_01" {
				hit = true
			}
		}
		So(hit, ShouldEqual, true)
	})
}

func ManagerSpec() {
	Convey("A manager must take a cacheing dependency", func() {
		discardQueue := rendertest.MakeStubbedRenderQueue()
		spyCache := &cachetest.SpyedCache{
			Impl: cache.MakeRamByteBank(),
		}
		spyCacheFactory := func(s string) cache.ByteBank {
			return spyCache
		}
		_ = sprite.MakeManager(discardQueue, spyCacheFactory)
	})
}
