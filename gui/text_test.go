package gui

import (
	"testing"

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
