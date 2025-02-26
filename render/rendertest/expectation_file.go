package rendertest

import (
	"fmt"
	"path"
	"strings"
)

type TestDataReference string

func NewTestdataReference(datakey string) TestDataReference {
	if strings.HasPrefix(datakey, "testdata/") {
		panic(fmt.Errorf("can't make a TestDataReference to a path that already starts with 'testdata/'"))
	}
	return TestDataReference(datakey)
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
