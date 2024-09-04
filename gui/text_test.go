package gui

import (
  "os"
  "testing"

  "github.com/runningwild/glop/render"
  "github.com/stretchr/testify/require"
)

func TestDictionaryMaxHeight(t *testing.T) {
  t.Run("default-height-is-zero", func(t *testing.T) {
    require := require.New(t)

    d := Dictionary{}

    require.Equal(0.0, d.MaxHeight())
  })
  t.Run("zero-height-at-non-zero-offset", func(t *testing.T) {
    require := require.New(t)
    
    d := Dictionary{
      data: dictData{
        Miny: 42.0,
        Maxy: 42.0,
      },
    }

    require.Equal(0.0, d.MaxHeight())
  })
  t.Run("height-clamped-non-negative", func(t *testing.T) {
    require := require.New(t)
    
    d := Dictionary{
      data: dictData{
        Miny: 42.0,
        Maxy: 0.0,
      },
    }

    require.Equal(0.0, d.MaxHeight())
  })
  t.Run("height-is-delta-min-max", func(t *testing.T) {
    require := require.New(t)
    
    d := Dictionary{
      data: dictData{
        Miny: 0.0,
        Maxy: 42.0,
      },
    }

    require.InDelta(42.0, d.MaxHeight(), 0.001)
  })
}

func TestRenderString(t *testing.T) {
  // TODO(tmckee): probably need to stop exporting Dictionary from gui and call
  // LoadDictionary to get an instance instead; it'll register shaders and
  // such.
  t.Run("rendering-should-not-panic", func(t *testing.T) {
    require := require.New(t)

    render.Init()

    dictReader, err := os.Open("../testdata/fonts/dict_10.gob")
    require.Nil(err)

    d, err := LoadDictionary(dictReader)
    require.Nil(err)

    d.RenderString("lol", 0, 0, 0, 12.0, Left)
  })
}
