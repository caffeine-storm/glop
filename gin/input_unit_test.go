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

func TestKeyDependency(t *testing.T) {
	t.Run("attempting to create a cycle panics", func(t *testing.T) {
		require := require.New(t)

		// TODO(tmckee): refactor other tests to use VoidLogger
		inputObj := MakeLogged(glog.VoidLogger())
		require.NotNil(inputObj)

		keyboard1 := DeviceId{
			Index: 1,
			Type:  DeviceTypeKeyboard,
		}
		kid := KeyId{
			Index:  0,
			Device: keyboard1,
		}

		keyIdA := kid
		keyIdA.Index = 'a'

		keyIdB := kid
		keyIdB.Index = 'b'

		keyIdC := kid
		keyIdC.Index = 'c'

		keyA := inputObj.GetKeyById(keyIdA)
		require.Panics(func() {
			inputObj.registerDependence(keyA, keyA.Id())
		}, "1-cycles are not allowed")

		keyB := inputObj.GetKeyById(keyIdB)
		inputObj.registerDependence(keyA, keyB.Id())

		require.Panics(func() {
			inputObj.registerDependence(keyB, keyA.Id())
		}, "2-cycles are not allowed")

		keyC := inputObj.GetKeyById(keyIdC)
		inputObj.registerDependence(keyB, keyC.Id())

		require.Panics(func() {
			inputObj.registerDependence(keyC, keyA.Id())
		}, "no cycles are allowed")
	})
}
