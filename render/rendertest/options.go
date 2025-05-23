package rendertest

import (
	"fmt"
	"image/color"
	"reflect"
)

type TestNumber uint8
type FileExtension string
type Threshold uint8
type BackgroundColour color.Color
type MakeRejectFiles bool
type RawDumpFilePath string

var defaultTestNumber = TestNumber(0)
var defaultFileExtension = FileExtension("png")
var defaultThreshold = Threshold(3)
var defaultFilePath = RawDumpFilePath("/dev/null")

// DefaultBackground is an opaque black
var DefaultBackground BackgroundColour = color.RGBA{
	R: 0,
	G: 0,
	B: 0,
	A: 255,
}

var defaultMakeRejectFiles = MakeRejectFiles(true)

// For the given slice of trailing arguments to a 'convey.So' call, look for a
// value with the same type as 'defaultValue'. If found, assign it to the
// pointer wrapped in 'output', otherwise, assign 'defaultValue' to the pointer
// wrapped in 'output'. Return true iff the value written to 'output' was found
// in 'args'.
func getFromArgs(args []interface{}, defaultValue interface{}, output interface{}) bool {
	defaultReflectValue := reflect.ValueOf(defaultValue)
	targetType := defaultReflectValue.Type()
	outPtr := reflect.ValueOf(output).Elem()

	// We start at the second element because the first element always has to be
	// the testdata 'key'.
	for i := 1; i < len(args); i++ {
		val := reflect.ValueOf(args[i])
		if val.Type() == targetType {
			outPtr.Set(val)
			return true
		}
	}

	outPtr.Set(defaultReflectValue)
	return false
}

// TODO(tmckee:clean): this should just take a single arg so that callsites
// will have 'getTestDataKey(args[0])' when they go to clobber 'args[0]'.
// Otherwise, it looks like we're overwriting the first arg with... who knows
// what.
func getTestDataKeyFromArgs(args []interface{}) TestDataReference {
	// The only valid spot to look for a test data reference is at the head of
	// the slice.
	if len(args) < 1 {
		panic(fmt.Errorf("need a non-empty slice of options for getting the test data key"))
	}

	// It might be a TestDataKey already, otherwise it has to be a string.
	switch v := args[0].(type) {
	case string:
		return NewTestdataReference(v)
	case TestDataReference:
		v.MustValidate()
		return v
	}

	panic(fmt.Errorf("expected type string or TestDataReference, got %T", args[0]))
}

func getTestNumberFromArgs(args []interface{}) TestNumber {
	var result TestNumber
	getFromArgs(args, defaultTestNumber, &result)
	return result
}

func getThresholdFromArgs(args []interface{}) Threshold {
	var result Threshold
	getFromArgs(args, defaultThreshold, &result)
	return result
}

func getBackgroundFromArgs(args []interface{}) (BackgroundColour, bool) {
	var result BackgroundColour
	found := getFromArgs(args, DefaultBackground, &result)
	return result, found
}

func getFileExtensionFromArgs(args []interface{}) FileExtension {
	var result FileExtension
	getFromArgs(args, defaultFileExtension, &result)
	return result
}

func getMakeRejectFilesFromArgs(args []interface{}) MakeRejectFiles {
	var result MakeRejectFiles
	getFromArgs(args, defaultMakeRejectFiles, &result)
	return result
}

func getDumpRawImageFromArgs(args []interface{}) (RawDumpFilePath, bool) {
	var result RawDumpFilePath
	found := getFromArgs(args, defaultFilePath, &result)
	return result, found
}
