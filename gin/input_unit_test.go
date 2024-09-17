package gin

import (
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPressKeyWorksForControlKey(t *testing.T) {
	require := require.New(t)

	inputObj := Make()
	require.NotNil(inputObj)

	group := EventGroup{Timestamp: 10}
	keyId := AnyLeftControl
	keyId.Device.Index = 1
	key := &keyState{
		id:         keyId,
		name:       "test-keystate(any-left-control)",
		cursor:     nil,
		aggregator: &standardAggregator{},
	}
	inputObj.pressKey(key, 1.0, Event{}, &group)
}

func TestBindingGetPrimaryPressAmt(t *testing.T) {
	require := require.New(t)

	inputObj := MakeLogged(log.New(os.Stdout, "herp: ", 0))
	require.NotNil(inputObj)

	xKey := AnyKeyX
	leftControl := AnyLeftControl

	// TODO(tmckee): GetKeyFlat is being called here for the side effect that it
	// primes caches/collections in Input. We should make it work without this
	// priming.
	flatKey := inputObj.GetKeyFlat(xKey.Index, DeviceTypeKeyboard, DeviceIndexAny)
	require.False(flatKey.IsDown(), "sanity check ; a new key must not be down already!!!")

	// LeftCtrl+x
	binding := inputObj.MakeBinding(xKey, []KeyId{leftControl}, []bool{true})

	testKey := inputObj.BindDerivedKey("test-binding", binding)

	primaryPressAmount := binding.primaryPressAmt()
	require.Equal(0.0, primaryPressAmount)

	keyPressAmount := testKey.CurPressAmt()
	require.Equal(0.0, keyPressAmount)
}
