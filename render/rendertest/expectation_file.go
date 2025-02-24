package rendertest

import "fmt"

type TestDataReference string

func NewTestdataReference(datakey string) TestDataReference {
	return TestDataReference(datakey)
}

func (ref *TestDataReference) Path(args ...interface{}) string {
	// Work around get*FromArgs having to skip arg0
	args = append([]interface{}{nil}, args...)

	filename := *ref
	testnumber := getTestNumberFromArgs(args)
	fileExtension := getFileExtensionFromArgs(args)
	return fmt.Sprintf("testdata/%s/%d.%s", filename, testnumber, fileExtension)
}

func (ref *TestDataReference) PathNumber(n int) string {
	return fmt.Sprintf("testdata/%s/%d.png", *ref, n)
}

func (ref *TestDataReference) PathExtension(ext string) string {
	return fmt.Sprintf("testdata/%s/0.%s", *ref, ext)
}
