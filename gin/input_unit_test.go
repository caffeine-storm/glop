package gin

import (
	"os"
	"testing"

	"github.com/runningwild/glop/glog"
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
		aggregator: &standardAggregator{},
	}
	inputObj.pressKey(key, 1.0, Event{}, &group)
}

func TestBindingGetPrimaryPressAmt(t *testing.T) {
	require := require.New(t)

	glogopts := &glog.Opts{
		Output: os.Stdout,
	}
	inputObj := MakeLogged(glog.New(glogopts))
	require.NotNil(inputObj)

	xKey := AnyKeyX
	leftControl := AnyLeftControl

	// TODO(tmckee): GetKeyByParts is being called here for the side effect that
	// it primes caches/collections in Input. We should make it work without this
	// priming.
	flatKey := inputObj.GetKeyByParts(xKey.Index, DeviceTypeKeyboard, DeviceIndexAny)
	require.False(flatKey.IsDown(), "sanity check ; a new key must not be down already!!!")

	// LeftCtrl+x
	binding := inputObj.MakeBinding(xKey, []KeyId{leftControl}, []bool{true})

	testKey := inputObj.BindDerivedKey("test-binding", binding)

	primaryPressAmount := binding.primaryPressAmt()
	require.Equal(0.0, primaryPressAmount)

	keyPressAmount := testKey.CurPressAmt()
	require.Equal(0.0, keyPressAmount)
}
