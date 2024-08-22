package algorithm

import (
  "fmt"
  "reflect"
)

type Chooser func(interface{}) bool

// TODO(tmckee): this could be done better with generics.
// Given a pointer to a slice and a Chooser, rewrite the slice to contain only
// those elements for which choose() returns true. The elements of the
// resulting slice will be in the same order as they were in the input.
func Choose(_a interface{}, chooser interface{}) {
  a := reflect.ValueOf(_a)
  if a.Kind() != reflect.Ptr || a.Elem().Kind() != reflect.Slice {
    panic(fmt.Sprintf("Can only Choose from a pointer to a slice, not a %v", a))
  }

  c := reflect.ValueOf(chooser)
  if c.Kind() != reflect.Func {
    panic(fmt.Sprintf("chooser must be a func, not a %v", c))
  }
  if c.Type().NumIn() != 1 {
    panic("chooser must take exactly 1 input parameter")
  }
  if c.Type().In(0).Kind() != a.Elem().Type().Elem().Kind() {
    panic(fmt.Sprintf("chooser's parameter must be %v, not %v", a.Elem().Type().Elem().Kind(), c.Type().In(0)))
  }
  if c.Type().NumOut() != 1 || c.Type().Out(0).Kind() != reflect.Bool {
    panic("chooser must have exactly 1 output parameter, a bool")
  }

  outputIndex := 0
  slice := a.Elem()
  fmt.Println("got input slice", slice)
  for i := 0; i < slice.Len(); i++ {
    inElem := slice.Index(i)
    filterResult := c.Call([]reflect.Value{inElem})
    include := filterResult[0].Bool()
    if include {
      // Don't need to overwrite if outputIndex is inputIndex
      if outputIndex != i {
        fmt.Println("writing idx", i, "to idx", outputIndex)
        slice.Index(outputIndex).Set(slice.Index(i))
      }
      outputIndex++
    }
  }

  // *a = (*a)[:newLen]
  slice.Set(slice.Slice(0, outputIndex))
}

type Mapper func(a interface{}) interface{}

// TODO(tmckee): this could be done better with generics.
func Map(_in interface{}, _out interface{}, mapper interface{}) {
  in := reflect.ValueOf(_in)
  if in.Kind() != reflect.Slice {
    panic(fmt.Sprintf("Can only Map from a slice, not a %v", in))
  }

  out := reflect.ValueOf(_out)
  if out.Kind() != reflect.Ptr || out.Elem().Kind() != reflect.Slice {
    panic(fmt.Sprintf("Can only Map to a pointer to a slice, not a %v", out))
  }

  m := reflect.ValueOf(mapper)
  if m.Kind() != reflect.Func {
    panic(fmt.Sprintf("mapper must be a func, not a %v", m))
  }
  if m.Type().NumIn() != 1 {
    panic("chooser must take exactly 1 input parameter")
  }
  if m.Type().In(0).Kind() != in.Type().Elem().Kind() {
    panic(fmt.Sprintf("mapper's parameter must be %v, not %v", in.Type().Elem().Kind(), m.Type().In(0)))
  }
  if m.Type().NumOut() != 1 {
    panic("chooser must have exactly 1 output parameter")
  }
  if m.Type().Out(0).Kind() != out.Elem().Type().Elem().Kind() {
    panic(fmt.Sprintf("mapper's output parameter must be %v, not %v", out.Elem().Type().Elem().Kind(), m.Type().Out(0)))
  }

  if out.Elem().Len() < in.Len() {
    slice := reflect.MakeSlice(out.Elem().Type(), in.Len(), in.Len())
    out.Elem().Set(slice)
  }

  out.Elem().SetLen(in.Len())
  for i := 0; i < out.Elem().Len(); i++ {
    v := m.Call([]reflect.Value{in.Index(i)})
    out.Elem().Index(i).Set(v[0])
  }
}
