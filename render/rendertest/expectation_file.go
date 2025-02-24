package rendertest

type TestDataReference string

func NewTestdataReference(datakey string) TestDataReference {
	return TestDataReference(datakey)
}
