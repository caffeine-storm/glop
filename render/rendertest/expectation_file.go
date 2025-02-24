package rendertest

import "fmt"

type TestDataReference string

func NewTestdataReference(datakey string) TestDataReference {
	return TestDataReference(datakey)
}

func (ref *TestDataReference) Path() string {
	return fmt.Sprintf("testdata/%s/0.png", *ref)
}

func (ref *TestDataReference) PathNumber(n int) string {
	return fmt.Sprintf("testdata/%s/%d.png", *ref, n)
}
