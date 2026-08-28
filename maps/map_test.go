package maps_test

import (
	"cmp"
	"errors"
	"fmt"
	"iter"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/pseudomuto/go/maps"
)

// pt is comparable but not cmp.Ordered. Using it as a map key means the tests
// that reference it stop compiling if MapKeys or SortedKeysFunc ever demand
// cmp.Ordered again.
type pt struct{ X, Y int }

func TestMap(t *testing.T) {
	t.Parallel()

	kv := func(e maps.Entry[string, int]) string { return e.Key + "=" + strconv.Itoa(e.Value) }

	tests := []struct {
		name string
		in   map[string]int
		// Map yields in randomized order, so results are compared sorted.
		want []string
	}{
		{name: "nil map", in: nil, want: nil},
		{name: "empty map", in: map[string]int{}, want: nil},
		{name: "single entry", in: map[string]int{"a": 1}, want: []string{"a=1"}},
		{
			name: "several entries",
			in:   map[string]int{"c": 3, "a": 1, "b": 2},
			want: []string{"a=1", "b=2", "c=3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tt.want, slices.Sorted(maps.Map(tt.in, kv)))
		})
	}
}

func TestMapKeepsKeyAndValueDistinct(t *testing.T) {
	t.Parallel()

	// Key and value share a type, so a transposed fn would compile.
	m := map[string]string{"k": "v"}

	got := slices.Collect(maps.Map(m, func(e maps.Entry[string, string]) string {
		return "key=" + e.Key + ",value=" + e.Value
	}))

	require.Equal(t, []string{"key=k,value=v"}, got)
}

func TestMapWithStructKey(t *testing.T) {
	t.Parallel()

	m := map[pt]string{{X: 1, Y: 2}: "a", {X: 3, Y: 4}: "b"}

	got := slices.Sorted(maps.Map(m, func(e maps.Entry[pt, string]) int { return e.Key.X }))

	require.Equal(t, []int{1, 3}, got)
}

func TestMapIsLazy(t *testing.T) {
	t.Parallel()

	calls := 0
	mapped := maps.Map(map[string]int{"a": 1, "b": 2, "c": 3}, func(e maps.Entry[string, int]) int {
		calls++
		return e.Value
	})

	require.Zero(t, calls, "fn ran before the sequence was consumed")
	require.Len(t, slices.Collect(mapped), 3)
	require.Equal(t, 3, calls, "fn should run exactly once per entry")
}

func TestMapStopsEarly(t *testing.T) {
	t.Parallel()

	m := map[string]int{"a": 1, "b": 2, "c": 3, "d": 4}
	calls := 0

	// Which entries arrive is randomized, so only the count is asserted.
	var got []string
	for v := range maps.Map(m, func(e maps.Entry[string, int]) string {
		calls++
		return e.Key
	}) {
		got = append(got, v)
		if len(got) == 2 {
			break
		}
	}

	require.Len(t, got, 2)
	require.Equal(t, 2, calls, "fn ran for entries the consumer never saw")
}

var errBoom = errors.New("boom")

type mapped struct {
	val string
	err error
}

// collectMapped drains a MapErr sequence and sorts it, since map iteration order
// is randomized. Failures all carry the empty value, so they sort together.
func collectMapped(s iter.Seq2[string, error]) []mapped {
	var out []mapped
	for v, err := range s {
		out = append(out, mapped{val: v, err: err})
	}

	slices.SortFunc(out, func(a, b mapped) int { return cmp.Compare(a.val, b.val) })

	return out
}

// evenEntry fails on odd values so error positions are easy to pin.
func evenEntry(e maps.Entry[string, int]) (string, error) {
	if e.Value%2 != 0 {
		return "", errBoom
	}

	return e.Key + "=" + strconv.Itoa(e.Value), nil
}

func TestMapErr(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   map[string]int
		want []mapped
	}{
		{name: "nil map", in: nil, want: nil},
		{name: "empty map", in: map[string]int{}, want: nil},
		{
			name: "every entry succeeds",
			in:   map[string]int{"a": 2, "b": 4},
			want: []mapped{{val: "a=2"}, {val: "b=4"}},
		},
		{
			name: "one entry fails",
			in:   map[string]int{"a": 2, "b": 3},
			want: []mapped{{err: errBoom}, {val: "a=2"}},
		},
		{
			name: "every entry fails",
			in:   map[string]int{"a": 1, "b": 3},
			want: []mapped{{err: errBoom}, {err: errBoom}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tt.want, collectMapped(maps.MapErr(tt.in, evenEntry)))
		})
	}
}

func TestMapErrDoesNotStopOnError(t *testing.T) {
	t.Parallel()

	// Four entries, two of them failing. A consumer that skips errors must still
	// see both successes, so MapErr cannot bail out on the first failure.
	m := map[string]int{"a": 1, "b": 2, "c": 3, "d": 4}

	var ok []string
	failures := 0

	for v, err := range maps.MapErr(m, evenEntry) {
		if err != nil {
			failures++
			continue
		}

		ok = append(ok, v)
	}

	slices.Sort(ok)

	require.Equal(t, []string{"b=2", "d=4"}, ok)
	require.Equal(t, 2, failures)
}

func TestMapErrStopsEarly(t *testing.T) {
	t.Parallel()

	m := map[string]int{"a": 1, "b": 2, "c": 3, "d": 4}
	calls := 0

	// Which entry arrives first is randomized, so only the call count is
	// asserted: breaking must stop fn from running again.
	for range maps.MapErr(m, func(e maps.Entry[string, int]) (string, error) {
		calls++
		return evenEntry(e)
	}) {
		break
	}

	require.Equal(t, 1, calls, "fn kept running after the consumer stopped")
}

func TestMapErrIsLazy(t *testing.T) {
	t.Parallel()

	calls := 0
	mapped := maps.MapErr(map[string]int{"a": 2, "b": 4, "c": 6}, func(e maps.Entry[string, int]) (string, error) {
		calls++
		return evenEntry(e)
	})

	require.Zero(t, calls, "fn ran before the sequence was consumed")
	require.Len(t, collectMapped(mapped), 3)
	require.Equal(t, 3, calls, "fn should run exactly once per entry")
}

func TestMapErrKeepsKeyAndValueDistinct(t *testing.T) {
	t.Parallel()

	// Key and value share a type, so a transposed fn would compile.
	m := map[string]string{"k": "v"}

	got := collectMapped(maps.MapErr(m, func(e maps.Entry[string, string]) (string, error) {
		return "key=" + e.Key + ",value=" + e.Value, nil
	}))

	require.Equal(t, []mapped{{val: "key=k,value=v"}}, got)
}

func TestMapKeys(t *testing.T) {
	t.Parallel()

	m := map[string]int{"ada": 1, "grace": 2, "alan": 3}

	// MapKeys yields in randomized map order, so compare as a set.
	got := slices.Sorted(maps.MapKeys(m, strings.ToUpper))

	require.Equal(t, []string{"ADA", "ALAN", "GRACE"}, got)
}

func TestMapKeysWithStructKey(t *testing.T) {
	t.Parallel()

	m := map[pt]string{{X: 1, Y: 2}: "a", {X: 3, Y: 4}: "b"}

	got := slices.Sorted(maps.MapKeys(m, func(p pt) int { return p.X }))

	require.Equal(t, []int{1, 3}, got)
}

func TestMapValues(t *testing.T) {
	t.Parallel()

	m := map[string]string{"k1": "ada", "k2": "grace", "k3": "alan"}

	// MapValues yields in randomized map order, so compare as a set.
	got := slices.Sorted(maps.MapValues(m, strings.ToUpper))

	require.Equal(t, []string{"ADA", "ALAN", "GRACE"}, got)
}

func TestMapValuesWithStructKey(t *testing.T) {
	t.Parallel()

	m := map[pt]int{{X: 1, Y: 2}: 10, {X: 3, Y: 4}: 20}

	got := slices.Sorted(maps.MapValues(m, func(v int) int { return v * 2 }))

	require.Equal(t, []int{20, 40}, got)
}

func ExampleMap() {
	m := map[string]int{"ada": 1, "grace": 2}

	// Map yields in randomized order, so sort for a stable result.
	fmt.Println(slices.Sorted(maps.Map(m, func(e maps.Entry[string, int]) string {
		return fmt.Sprintf("%s=%d", e.Key, e.Value)
	})))
	// Output: [ada=1 grace=2]
}

func ExampleMapErr() {
	m := map[string]int{"a": 2, "b": 3, "c": 4}

	// fn fails on odd values. MapErr keeps going, so the consumer decides. Order
	// is randomized, so tally rather than list.
	ok, failed := 0, 0

	for _, err := range maps.MapErr(m, func(e maps.Entry[string, int]) (string, error) {
		if e.Value%2 != 0 {
			return "", fmt.Errorf("odd value for %q", e.Key)
		}

		return e.Key, nil
	}) {
		if err != nil {
			failed++
			continue
		}

		ok++
	}

	fmt.Println("ok:", ok, "failed:", failed)
	// Output: ok: 2 failed: 1
}

func ExampleMapKeys() {
	m := map[string]int{"ada": 1, "grace": 2}

	// MapKeys yields in randomized map order, so sort for a stable result.
	fmt.Println(slices.Sorted(maps.MapKeys(m, strings.ToUpper)))
	// Output: [ADA GRACE]
}

func ExampleMapValues() {
	m := map[string]string{"k1": "ada", "k2": "grace"}

	// MapValues yields in randomized map order, so sort for a stable result.
	fmt.Println(slices.Sorted(maps.MapValues(m, strings.ToUpper)))
	// Output: [ADA GRACE]
}
