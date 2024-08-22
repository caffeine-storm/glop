package algorithm_test

import (
  "fmt"

  "github.com/orfjackal/gospec/src/gospec"
  . "github.com/orfjackal/gospec/src/gospec"
  "github.com/runningwild/glop/util/algorithm"
)

func ChooserSpec(c gospec.Context) {
  c.Specify("Choose on []int", func() {
    data := []int{0,1,2,3,4,5,6,7,8,9}
    a := make([]int, len(data))

    copy(a, data)
    algorithm.Choose(&a, func(v int) bool { return (v % 2) == 0 })
    c.Expect(a, ContainsInOrder, []int{0, 2, 4, 6, 8})

    a = make([]int, len(data))
    copy(a, data)
    algorithm.Choose(&a, func(v int) bool { return (v % 2) == 1 })
    c.Expect(a, ContainsInOrder, []int{1, 3, 5, 7, 9})

    a = make([]int, len(data))
    copy(a, data)
    algorithm.Choose(&a, func(v int) bool { return true })
    c.Expect(a, ContainsInOrder, a)

    a = make([]int, len(data))
    copy(a, data)
    algorithm.Choose(&a, func(v int) bool { return false })
    c.Expect(a, ContainsInOrder, []int{})

    a = make([]int, len(data))
    copy(a, data)
    algorithm.Choose(&a, func(v int) bool { return false })
    c.Expect(a, ContainsInOrder, []int{})
  })

  c.Specify("Choose on []string", func() {
    data := []string{"foo", "bar", "wing", "ding", "monkey", "machine"}

    a := make([]string, len(data))
    copy(a, data)
    algorithm.Choose(&a, func(v string) bool { return v > "foo" })
    c.Expect(a, ContainsInOrder, []string{"wing", "monkey", "machine"})

    a = make([]string, len(data))
    copy(a, data)
    algorithm.Choose(&a, func(v string) bool { return v < "foo" })
    c.Expect(a, ContainsInOrder, []string{"bar", "ding"})
  })
}


func MapperSpec(c gospec.Context) {
  c.Specify("Map from []int to []float64", func() {
    a := []int{0,1,2,3,4}
    var b []float64
    algorithm.Map(a, &b, func(v int) float64 { return float64(v) })
    c.Expect(b, ContainsInOrder, []float64{0,1,2,3,4})
  })
  c.Specify("Map from []int to []string", func() {
    a := []int{0,1,2,3,4}
    var b []string
    algorithm.Map(a, &b, func(v int) string { return fmt.Sprintf("%d", v) })
    c.Expect(b, ContainsInOrder, []string{"0", "1", "2", "3", "4"})
  })
}
