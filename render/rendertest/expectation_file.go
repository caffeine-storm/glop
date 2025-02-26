package rendertest

import (
	"fmt"
	"path"
	"strings"
)

type TestDataReference string

func NewTestdataReference(datakey string) TestDataReference {
	result := TestDataReference(datakey)
	result.MustValidate()
	return result
}

func (ref *TestDataReference) Validate() bool {
	return !strings.HasPrefix(string(*ref), "testdata")
}

func (ref *TestDataReference) MustValidate() {
	if !ref.Validate() {
		panic(fmt.Errorf("invalid TestDataReference: %q; there must be no leading 'testdata' prefix", string(*ref)))
	}
}

func (ref *TestDataReference) Path(args ...interface{}) string {
	ref.MustValidate()

	// Work around get*FromArgs having to skip arg0
	args = append([]interface{}{nil}, args...)

	filename := *ref
	testnumber := getTestNumberFromArgs(args)
	fileExtension := getFileExtensionFromArgs(args)
	return fmt.Sprintf("testdata/%s/%d.%s", filename, testnumber, fileExtension)
}

func (ref *TestDataReference) PathNumber(n int) string {
	ref.MustValidate()
	return fmt.Sprintf("testdata/%s/%d.png", *ref, n)
}

func (ref *TestDataReference) PathExtension(ext string) string {
	ref.MustValidate()
	return fmt.Sprintf("testdata/%s/0.%s", *ref, ext)
}

func ExpectationFile(testDataKey TestDataReference, fileExt FileExtension, testnumber TestNumber) string {
	return testDataKey.Path(fileExt, testnumber)
}

// Return the given file but with a '.rej' component to signify a 'rejection'.
func MakeRejectName(exp, suffix string) string {
	dir, expectedFileName := path.Split(exp)
	rejectFileNameBase, ok := strings.CutSuffix(expectedFileName, suffix)
	if !ok {
		panic(fmt.Errorf("need a %s file, got %s", suffix, exp))
	}
	return path.Join(dir, rejectFileNameBase+".rej"+suffix)
}
