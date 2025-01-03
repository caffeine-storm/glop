package ai_test

import (
	"github.com/runningwild/glop/ai"
	"github.com/runningwild/polish"
	yed "github.com/runningwild/yedparse"

	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestAiSpecs(t *testing.T) {
	Convey("Specs for the AI package", t, func() {
		Convey("XGML Load", XgmlLoadSpec)
		Convey("Term", TermSpec)
		Convey("Chunk", ChunkSpec)
	})
}

func XgmlLoadSpec() {
	Convey("Load a simple .xgml file", func() {
		g, err := yed.ParseFromFile("state.xgml")
		So(err, ShouldEqual, nil)
		aig := ai.NewGraph()
		aig.Graph = &g.Graph
		aig.Context = polish.MakeContext()
		polish.AddIntMathContext(aig.Context)

		dist := 0
		dist_func := func() int {
			return dist
		}

		var nearest int = 7
		nearest_func := func() int {
			return nearest
		}

		attacks := 0
		attack_func := func() int {
			attacks++
			return 0
		}

		aig.Context.AddFunc("dist", dist_func)
		aig.Context.AddFunc("nearest", nearest_func)
		aig.Context.AddFunc("move", func() int { nearest--; return 0 })
		aig.Context.AddFunc("wait", func() int { return 0 })
		aig.Context.AddFunc("attack", attack_func)
		aig.Eval(2, func() bool { return true })

		So(attacks, ShouldEqual, 0)
		So(nearest, ShouldEqual, 4)
	})
}

func TermSpec() {
	g, err := yed.ParseFromFile("state.xgml")
	So(err, ShouldEqual, nil)
	aig := ai.NewGraph()
	aig.Graph = &g.Graph
	aig.Context = polish.MakeContext()
	polish.AddIntMathContext(aig.Context)
	polish.AddIntMathContext(aig.Context)

	Convey("Calling AiGraph.Term() will terminate evaluation early.", func() {
		var nearest int = 7
		nearest_func := func() int {
			return nearest
		}

		dist := 0
		term := true
		dist_func := func() int {
			if nearest == 6 && term {
				aig.Term() <- nil
			}
			return dist
		}

		attacks := 0
		attack_func := func() int {
			attacks++
			return 0
		}

		aig.Context.AddFunc("dist", dist_func)
		aig.Context.AddFunc("nearest", nearest_func)
		aig.Context.AddFunc("move", func() int { nearest--; return 0 })
		aig.Context.AddFunc("wait", func() int { return 0 })
		aig.Context.AddFunc("attack", attack_func)
		aig.Eval(2, func() bool { return true })

		So(attacks, ShouldEqual, 0)
		So(nearest, ShouldEqual, 6)

		term = false
		aig.Eval(2, func() bool { return true })
		So(nearest, ShouldEqual, 4)
	})
}

func ChunkSpec() {
	g, err := yed.ParseFromFile("state.xgml")
	So(err, ShouldEqual, nil)
	aig := ai.NewGraph()
	aig.Graph = &g.Graph
	aig.Context = polish.MakeContext()
	polish.AddIntMathContext(aig.Context)
	polish.AddIntMathContext(aig.Context)
	Convey("cont() returning false will terminate evaluation early.", func() {
		var nearest int = 7
		nearest_func := func() int {
			return nearest
		}

		dist := 0
		term := true
		dist_func := func() int {
			if nearest == 6 && term {
				aig.Term() <- nil
			}
			return dist
		}

		attacks := 0
		attack_func := func() int {
			attacks++
			return 0
		}

		aig.Context.AddFunc("dist", dist_func)
		aig.Context.AddFunc("nearest", nearest_func)
		aig.Context.AddFunc("move", func() int { nearest--; return 0 })
		aig.Context.AddFunc("wait", func() int { return 0 })
		aig.Context.AddFunc("attack", attack_func)
		_, err := aig.Eval(4, func() bool { return false })
		// Only have time for 1 move before we terminate early
		So(err, ShouldEqual, nil)
		So(nearest, ShouldEqual, 6)
	})
}
