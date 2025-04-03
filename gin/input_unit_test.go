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

	t.Run("attempting to create a cycle panics", func(t *testing.T) {
		require := require.New(t)

		// TODO(tmckee): refactor other tests to use VoidLogger
		inputObj := MakeLogged(glog.VoidLogger())
		require.NotNil(inputObj)

		keyA := inputObj.GetKeyById(keyIdA)
		require.True(inputObj.willTrigger(keyA.Id(), keyA.Id()))
		require.Panics(func() {
			inputObj.addCauseEffect(keyA.Id(), keyA)
		}, "1-cycles are not allowed")

		keyB := inputObj.GetKeyById(keyIdB)
		require.False(inputObj.willTrigger(keyA.Id(), keyB.Id()))
		require.False(inputObj.willTrigger(keyB.Id(), keyA.Id()))

		inputObj.addCauseEffect(keyA.Id(), keyB)
		require.True(inputObj.willTrigger(keyA.Id(), keyB.Id()))

		require.Panics(func() {
			inputObj.addCauseEffect(keyB.Id(), keyA)
		}, "2-cycles are not allowed")

		keyC := inputObj.GetKeyById(keyIdC)
		inputObj.addCauseEffect(keyB.Id(), keyC)
		require.True(inputObj.willTrigger(keyB.Id(), keyC.Id()))
		require.True(inputObj.willTrigger(keyA.Id(), keyC.Id()))

		require.Panics(func() {
			inputObj.addCauseEffect(keyC.Id(), keyA)
		}, "no cycles are allowed")
	})

	t.Run("can remove cause-effect links", func(t *testing.T) {
		require := require.New(t)

		// TODO(tmckee): refactor other tests to use VoidLogger
		inputObj := MakeLogged(glog.VoidLogger())
		require.NotNil(inputObj)

		keyA := inputObj.GetKeyById(keyIdA)
		keyB := inputObj.GetKeyById(keyIdB)
		keyC := inputObj.GetKeyById(keyIdC)

		// Setup A -causes-> B -causes-> C
		inputObj.addCauseEffect(keyA.Id(), keyB)
		inputObj.addCauseEffect(keyB.Id(), keyC)

		require.True(inputObj.willTrigger(keyA.Id(), keyB.Id()))
		require.True(inputObj.willTrigger(keyA.Id(), keyC.Id()))

		inputObj.removeCauseEffect(keyB.Id(), keyC)

		require.True(inputObj.willTrigger(keyA.Id(), keyB.Id()))
		require.False(inputObj.willTrigger(keyA.Id(), keyC.Id()))
	})
}
