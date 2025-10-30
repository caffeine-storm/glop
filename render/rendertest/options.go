package rendertest

import (
	"fmt"
	"image/color"
	"reflect"
)

type (
	TestNumber        uint8
	FileExtension     string
	Threshold         uint8
	BackgroundColour  color.Color
	MakeRejectFiles   bool
	DebugDumpFilePath string
)

var (
	defaultTestNumber        = TestNumber(0)
	defaultFileExtension     = FileExtension("png")
	defaultThreshold         = Threshold(3)
	defaultDebugDumpFilePath = DebugDumpFilePath("/dev/null")
)

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

func getTestDataKey(arg any) TestDataReference {
	// It might be a TestDataKey already, otherwise it has to be a string.
	switch v := arg.(type) {
	case string:
		return NewTestdataReference(v)
	case TestDataReference:
		v.MustValidate()
		return v
	}

	panic(fmt.Errorf("expected type string or TestDataReference, got %T", arg))
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

func getDebugDumpFilePathFromArgs(args []interface{}) (DebugDumpFilePath, bool) {
	var result DebugDumpFilePath
	found := getFromArgs(args, defaultDebugDumpFilePath, &result)
	return result, found
}
