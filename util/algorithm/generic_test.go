package algorithm_test

import (
	"fmt"
	"testing"

	"github.com/caffeine-storm/glop/util/algorithm"
	. "github.com/smartystreets/goconvey/convey"
)

func TestAlgorithmGenericSpecs(t *testing.T) {
	Convey("ChooserSpec", t, ChooserSpec)
	Convey("MapperSpec", t, MapperSpec)
}

func ChooserSpec() {
	Convey("Choose on []int", func() {
		data := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
		a := make([]int, len(data))

		copy(a, data)
		algorithm.Choose(&a, func(v int) bool { return (v % 2) == 0 })
		So(a, ShouldResemble, []int{0, 2, 4, 6, 8})

		a = make([]int, len(data))
		copy(a, data)
		algorithm.Choose(&a, func(v int) bool { return (v % 2) == 1 })
		So(a, ShouldResemble, []int{1, 3, 5, 7, 9})

		a = make([]int, len(data))
		copy(a, data)
		algorithm.Choose(&a, func(v int) bool { return true })
		So(a, ShouldResemble, a)

		a = make([]int, len(data))
		copy(a, data)
		algorithm.Choose(&a, func(v int) bool { return false })
		So(a, ShouldResemble, []int{})

		a = make([]int, len(data))
		copy(a, data)
		algorithm.Choose(&a, func(v int) bool { return false })
		So(a, ShouldResemble, []int{})
	})

	Convey("Choose on []string", func() {
		data := []string{"foo", "bar", "wing", "ding", "monkey", "machine"}

		a := make([]string, len(data))
		copy(a, data)
		algorithm.Choose(&a, func(v string) bool { return v > "foo" })
		So(a, ShouldResemble, []string{"wing", "monkey", "machine"})

		a = make([]string, len(data))
		copy(a, data)
		algorithm.Choose(&a, func(v string) bool { return v < "foo" })
		So(a, ShouldResemble, []string{"bar", "ding"})
	})
}

func MapperSpec() {
	Convey("Map from []int to []float64", func() {
		a := []int{0, 1, 2, 3, 4}
		var b []float64
		algorithm.Map(a, &b, func(v int) float64 { return float64(v) })
		So(b, ShouldResemble, []float64{0, 1, 2, 3, 4})
	})
	Convey("Map from []int to []string", func() {
		a := []int{0, 1, 2, 3, 4}
		var b []string
		algorithm.Map(a, &b, func(v int) string { return fmt.Sprintf("%d", v) })
		So(b, ShouldResemble, []string{"0", "1", "2", "3", "4"})
	})
}
