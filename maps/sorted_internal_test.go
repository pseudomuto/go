package maps

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/require"
)

// orderedEntries is documented as re-iterable and never serving a stale
// snapshot. No exported function exercises that, because every caller either
// re-invokes it per iteration or consumes it immediately, so it is asserted
// directly here.
func TestOrderedEntriesRefreshesEachIteration(t *testing.T) {
	t.Parallel()

	m := map[string]int{"a": 1}
	entries := orderedEntries(m)

	require.Equal(t, []Entry[string, int]{{Key: "a", Value: 1}}, slices.Collect(entries))

	m["b"] = 2

	require.Equal(t,
		[]Entry[string, int]{{Key: "a", Value: 1}, {Key: "b", Value: 2}},
		slices.Collect(entries),
		"orderedEntries served entries from a stale call-time snapshot")
}
