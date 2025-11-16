package algorithm_test

import (
	"testing"

	"github.com/caffeine-storm/glop/util/algorithm"
	. "github.com/smartystreets/goconvey/convey"
)

func TestAlgorithmGraphSpecs(t *testing.T) {
	Convey("DijkstraSpec", t, DijkstraSpec)
	Convey("ReachableSpec", t, ReachableSpec)
	Convey("ReachableDestinationsSpec", t, ReachableDestinationsSpec)
	Convey("TopoSpec", t, TopoSpec)
}

type board [][]int

func (b board) NumVertex() int {
	return len(b) * len(b[0])
}

func (b board) Adjacent(n int) ([]int, []float64) {
	x := n % len(b[0])
	y := n / len(b[0])
	var adj []int
	var weight []float64
	if x > 0 && b[y][x-1] > 0 {
		adj = append(adj, n-1)
		weight = append(weight, float64(b[y][x-1]))
	}
	if y > 0 && b[y-1][x] > 0 {
		adj = append(adj, n-len(b[0]))
		weight = append(weight, float64(b[y-1][x]))
	}
	if x < len(b[0])-1 && b[y][x+1] > 0 {
		adj = append(adj, n+1)
		weight = append(weight, float64(b[y][x+1]))
	}
	if y < len(b)-1 && b[y+1][x] > 0 {
		adj = append(adj, n+len(b[0]))
		weight = append(weight, float64(b[y+1][x]))
	}
	return adj, weight
}

func DijkstraSpec() {
	b := [][]int{
		{1, 2, 9, 4, 3, 2, 1}, // 0 - 6
		{9, 2, 9, 4, 3, 1, 1}, // 7 - 13
		{2, 1, 5, 5, 5, 2, 1}, // 14 - 20
		{1, 1, 1, 1, 1, 1, 1}, // 21 - 27
	}
	Convey("Check Dijkstra's gives the right path and weight", func() {
		weight, path := algorithm.Dijkstra(board(b), []int{0}, []int{11})
		So(weight, ShouldEqual, 16.0)
		So(path, ShouldResemble, []int{0, 1, 8, 15, 22, 23, 24, 25, 26, 19, 12, 11})
	})
	Convey("Check multiple sources", func() {
		weight, path := algorithm.Dijkstra(board(b), []int{0, 1, 7, 2}, []int{11})
		So(weight, ShouldEqual, 10.0)
		So(path, ShouldResemble, []int{2, 3, 4, 11})
	})
	Convey("Check multiple destinations", func() {
		weight, path := algorithm.Dijkstra(board(b), []int{0}, []int{6, 11, 21})
		So(weight, ShouldEqual, 7.0)
		So(path, ShouldResemble, []int{0, 1, 8, 15, 22, 21})
	})
}

func ReachableSpec() {
	b := [][]int{
		{1, 2, 9, 4, 3, 2, 1}, // 0 - 6
		{9, 2, 9, 4, 3, 1, 1}, // 7 - 13
		{2, 1, 5, 5, 5, 2, 1}, // 14 - 20
		{1, 1, 1, 1, 1, 1, 1}, // 21 - 27
	}
	Convey("Check reachability", func() {
		reach := algorithm.ReachableWithinLimit(board(b), []int{3}, 5)
		So(reach, ShouldResemble, []int{3, 4, 5, 10})
		reach = algorithm.ReachableWithinLimit(board(b), []int{3}, 10)
		So(reach, ShouldResemble, []int{2, 3, 4, 5, 6, 10, 11, 12, 13, 17, 19, 20, 24, 25, 26, 27})
	})
	Convey("Check reachability with multiple sources", func() {
		reach := algorithm.ReachableWithinLimit(board(b), []int{0, 6}, 3)
		So(reach, ShouldResemble, []int{0, 1, 5, 6, 12, 13, 20, 27})
		reach = algorithm.ReachableWithinLimit(board(b), []int{21, 27}, 2)
		So(reach, ShouldResemble, []int{13, 14, 15, 20, 21, 22, 23, 25, 26, 27})
	})
	Convey("Check bounds with multiple sources", func() {
		reach := algorithm.ReachableWithinBounds(board(b), []int{0, 6}, 2, 4)
		So(reach, ShouldResemble, []int{1, 5, 8, 12, 19, 20, 26, 27})
	})
}

func ReachableDestinationsSpec() {
	b := [][]int{
		{1, 2, 9, 4, 0, 2, 1}, // 0 - 6
		{0, 0, 0, 0, 0, 1, 1}, // 7 - 13
		{2, 1, 5, 5, 0, 2, 1}, // 14 - 20
		{1, 1, 1, 9, 0, 1, 1}, // 21 - 27
	}
	Convey("Check reachability", func() {
		reachable := algorithm.ReachableDestinations(board(b), []int{14}, []int{0, 2, 5, 13, 17, 22})
		So(reachable, ShouldResemble, []int{17, 22})
		reachable = algorithm.ReachableDestinations(board(b), []int{1, 26}, []int{0, 2, 5, 13, 17, 22})
		So(reachable, ShouldResemble, []int{0, 2, 5, 13})
	})
}

type adag [][]int

func (a adag) NumVertex() int {
	return len(a)
}

func (a adag) Successors(n int) []int {
	return a[n]
}

func (a adag) allSuccessorsHelper(n int, m map[int]bool) {
	for _, s := range a[n] {
		m[s] = true
		a.allSuccessorsHelper(s, m)
	}
}

func (a adag) AllSuccessors(n int) map[int]bool {
	if len(a[n]) == 0 {
		return nil
	}
	m := make(map[int]bool)
	a.allSuccessorsHelper(n, m)
	return m
}

func checkOrder(a adag, order []int) {
	So(len(a), ShouldEqual, len(order))
	Convey("Ordering contains all vertices exactly once", func() {
		all := make(map[int]bool)
		for _, v := range order {
			all[v] = true
		}
		So(len(all), ShouldEqual, len(order))
		for i := 0; i < len(a); i++ {
			So(all[i], ShouldEqual, true)
		}
	})
	Convey("Successors of a vertex always occur later in the ordering", func() {
		for i := 0; i < len(order); i++ {
			all := a.AllSuccessors(order[i])
			for j := range order {
				if i == j {
					continue
				}
				succ, ok := all[order[j]]
				if j < i {
					So(!ok, ShouldEqual, true)
				} else {
					So(!ok || succ, ShouldEqual, true)
				}
			}
		}
	})
}

func TopoSpec() {
	Convey("Check toposort on linked list", func() {
		a := adag{
			[]int{1},
			[]int{2},
			[]int{3},
			[]int{4},
			[]int{5},
			[]int{6},
			[]int{},
		}
		order := algorithm.TopoSort(a)
		checkOrder(a, order)
	})

	Convey("multi-edges don't mess up toposort", func() {
		a := adag{
			[]int{1, 1, 1},
			[]int{},
		}
		order := algorithm.TopoSort(a)
		checkOrder(a, order)
	})

	Convey("Check toposort on a more complicated digraph", func() {
		a := adag{
			[]int{8, 7, 4}, // 0
			[]int{5},
			[]int{0},
			[]int{9},
			[]int{14},
			[]int{15}, // 5
			[]int{1},
			[]int{},
			[]int{},
			[]int{13},
			[]int{3}, // 10
			[]int{12},
			[]int{18},
			[]int{16},
			[]int{},
			[]int{14}, // 15
			[]int{},
			[]int{},
			[]int{},
			[]int{},
			[]int{},
		}
		order := algorithm.TopoSort(a)
		checkOrder(a, order)
	})

	Convey("A cyclic digraph returns nil", func() {
		a := adag{
			[]int{8, 7, 4}, // 0
			[]int{5},
			[]int{0},
			[]int{9},
			[]int{14},
			[]int{15}, // 5
			[]int{1},
			[]int{},
			[]int{20},
			[]int{13},
			[]int{3}, // 10
			[]int{12},
			[]int{18},
			[]int{16},
			[]int{2},
			[]int{14}, // 15
			[]int{6},
			[]int{},
			[]int{},
			[]int{},
			[]int{},
		}
		order := algorithm.TopoSort(a)
		So(len(order), ShouldEqual, 0)
	})
}
