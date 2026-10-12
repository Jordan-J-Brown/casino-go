package slicesx

import (
	"testing"

	"github.com/alecthomas/assert/v2"
)

func TestDefragLeftPtr(t *testing.T) {
	t.Run("moves non-nil values left and nils to the right", func(t *testing.T) {
		a, b, c := 1, 2, 3
		slice := []*int{nil, &a, nil, &b, &c, nil}

		indexLast := DefragLeftPtr(slice)

		assert.Equal(t, []*int{&a, &b, &c, nil, nil, nil}, slice)
		assert.Equal(t, 2, indexLast)
	})

	t.Run("keeps existing order for non-nil values", func(t *testing.T) {
		a, b, c := 10, 20, 30
		slice := []*int{&a, nil, &b, nil, &c, nil, nil}

		indexLast := DefragLeftPtr(slice)

		assert.Equal(t, []*int{&a, &b, &c, nil, nil, nil, nil}, slice)
		assert.Equal(t, 2, indexLast)
	})

	t.Run("all nils", func(t *testing.T) {
		slice := []*int{nil, nil}
		lastIndex := DefragLeftPtr(slice)
		assert.Equal(t, []*int{nil, nil}, slice)
		assert.Equal(t, 0, lastIndex)
	})

	t.Run("all valid", func(t *testing.T) {
		a, b := new(1), new(2)
		slice := []*int{a, b}
		lastIndex := DefragLeftPtr(slice)
		assert.Equal(t, []*int{a, b}, slice)
		assert.Equal(t, 1, lastIndex)
	})
}
