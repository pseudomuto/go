package is_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/pseudomuto/go/validate"
	"github.com/pseudomuto/go/validate/is"
)

// The messages under test, named so the tables below stay readable and a
// reworded message shows up in one place.
const (
	notCIDR   = "is not a valid CIDR range"
	notCIDRv4 = "is not a valid CIDR(v4) range"
	notCIDRv6 = "is not a valid CIDR(v6) range"
	notIP     = "is not a valid IP address"
	notIPv4   = "is not a valid IPv4 address"
	notIPv6   = "is not a valid IPv6 address"
)

func TestCIDR(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		in      string
		wantErr string
	}{
		{name: "v4 network", in: "10.0.0.0/8"},
		{name: "v6 network", in: "2001:db8::/32"},
		{name: "v4 default route", in: "0.0.0.0/0"},
		{name: "v6 default route", in: "::/0"},
		{name: "single v4 host", in: "10.0.0.1/32"},
		{name: "host bits set is fine here", in: "10.0.0.5/24"},
		{name: "v6 host bits set is fine here", in: "2001:db8::1/32"},
		{name: "empty", in: "", wantErr: notCIDR},
		{name: "bare address, no mask", in: "10.0.0.0", wantErr: notCIDR},
		{name: "mask wider than v4 allows", in: "10.0.0.0/33", wantErr: notCIDR},
		{name: "mask wider than v6 allows", in: "2001:db8::/129", wantErr: notCIDR},
		{name: "not an address at all", in: "nope", wantErr: notCIDR},
		{name: "leading space", in: " 10.0.0.0/8", wantErr: notCIDR},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			requireCheck(t, is.CIDR(), tt.in, tt.wantErr)
		})
	}
}

func TestCIDRv4(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		in      string
		wantErr string
	}{
		{name: "network address", in: "10.0.0.0/8"},
		{name: "another network address", in: "192.168.1.0/24"},
		{name: "default route", in: "0.0.0.0/0"},
		{name: "single host is its own network", in: "10.0.0.1/32"},
		{name: "host bits set", in: "10.0.0.5/24", wantErr: notCIDRv4},
		{name: "one host bit set", in: "10.0.0.1/8", wantErr: notCIDRv4},
		{name: "v6 network", in: "2001:db8::/32", wantErr: notCIDRv4},
		{name: "v6 default route", in: "::/0", wantErr: notCIDRv4},
		{name: "v4-mapped v6 counts as v4", in: "::ffff:10.0.0.0/104"},
		{name: "bare address, no mask", in: "10.0.0.0", wantErr: notCIDRv4},
		{name: "empty", in: "", wantErr: notCIDRv4},
		{name: "not an address at all", in: "nope", wantErr: notCIDRv4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			requireCheck(t, is.CIDRv4(), tt.in, tt.wantErr)
		})
	}
}

func TestCIDRv6(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		in      string
		wantErr string
	}{
		{name: "network address", in: "2001:db8::/32"},
		{name: "default route", in: "::/0"},
		{name: "single host is its own network", in: "2001:db8::1/128"},
		{name: "host bits set", in: "2001:db8::1/32", wantErr: notCIDRv6},
		{name: "v4 network", in: "10.0.0.0/8", wantErr: notCIDRv6},
		{name: "v4 default route", in: "0.0.0.0/0", wantErr: notCIDRv6},
		{name: "v4-mapped v6 counts as v4", in: "::ffff:10.0.0.0/104", wantErr: notCIDRv6},
		{name: "bare address, no mask", in: "2001:db8::", wantErr: notCIDRv6},
		{name: "empty", in: "", wantErr: notCIDRv6},
		{name: "not an address at all", in: "nope", wantErr: notCIDRv6},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			requireCheck(t, is.CIDRv6(), tt.in, tt.wantErr)
		})
	}
}

// The documented difference between CIDR and the two family checks is that
// only the latter insist on a network address. Pin that down as one statement
// rather than leaving it implied by three tables.
func TestCIDRFamilyChecksRejectHostBitsThatCIDRAllows(t *testing.T) {
	t.Parallel()

	for _, in := range []string{"10.0.0.5/24", "2001:db8::1/32"} {
		require.NoError(t, is.CIDR()(in), "CIDR should not care about host bits")
		require.Error(t, is.CIDRv4()(in))
		require.Error(t, is.CIDRv6()(in))
	}
}

func TestIP(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		in      string
		wantErr string
	}{
		{name: "v4", in: "10.0.0.1"},
		{name: "v4 unspecified", in: "0.0.0.0"},
		{name: "v4 broadcast", in: "255.255.255.255"},
		{name: "v6", in: "2001:db8::1"},
		{name: "v6 loopback", in: "::1"},
		{name: "v6 unspecified", in: "::"},
		{name: "v4-mapped v6", in: "::ffff:10.0.0.1"},
		{name: "empty", in: "", wantErr: notIP},
		{name: "a prefix is not an address", in: "10.0.0.0/8", wantErr: notIP},
		{name: "octet out of range", in: "10.0.0.256", wantErr: notIP},
		{name: "too few octets", in: "10.0.1", wantErr: notIP},
		{name: "leading space", in: " 10.0.0.1", wantErr: notIP},
		{name: "hostname", in: "example.com", wantErr: notIP},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			requireCheck(t, is.IP(), tt.in, tt.wantErr)
		})
	}
}

func TestIPv4(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		in      string
		wantErr string
	}{
		{name: "dotted quad", in: "10.0.0.1"},
		{name: "unspecified", in: "0.0.0.0"},
		{name: "broadcast", in: "255.255.255.255"},
		{name: "v4-mapped v6 is v4 in another spelling", in: "::ffff:10.0.0.1"},
		{name: "v6", in: "2001:db8::1", wantErr: notIPv4},
		{name: "v6 loopback", in: "::1", wantErr: notIPv4},
		{name: "a prefix is not an address", in: "10.0.0.0/8", wantErr: notIPv4},
		{name: "empty", in: "", wantErr: notIPv4},
		{name: "octet out of range", in: "10.0.0.256", wantErr: notIPv4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			requireCheck(t, is.IPv4(), tt.in, tt.wantErr)
		})
	}
}

func TestIPv6(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		in      string
		wantErr string
	}{
		{name: "documentation prefix", in: "2001:db8::1"},
		{name: "loopback", in: "::1"},
		{name: "unspecified", in: "::"},
		{name: "fully written out", in: "2001:0db8:0000:0000:0000:0000:0000:0001"},
		{name: "v4-mapped v6 is treated as v4", in: "::ffff:10.0.0.1", wantErr: notIPv6},
		{name: "v4", in: "10.0.0.1", wantErr: notIPv6},
		{name: "a prefix is not an address", in: "2001:db8::/32", wantErr: notIPv6},
		{name: "empty", in: "", wantErr: notIPv6},
		{name: "too many groups", in: "2001:db8:0:0:0:0:0:0:1", wantErr: notIPv6},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			requireCheck(t, is.IPv6(), tt.in, tt.wantErr)
		})
	}
}

// IPv4 and IPv6 are documented as complements over everything IP accepts,
// which only holds if the v4-mapped case lands on exactly one side.
func TestIPFamilyChecksPartitionIP(t *testing.T) {
	t.Parallel()

	accepted := []string{"10.0.0.1", "0.0.0.0", "255.255.255.255", "2001:db8::1", "::1", "::", "::ffff:10.0.0.1"}

	for _, in := range accepted {
		require.NoError(t, is.IP()(in), "IP rejected %q, so the rest of this says nothing", in)

		isV4 := is.IPv4()(in) == nil
		isV6 := is.IPv6()(in) == nil

		require.NotEqual(t, isV4, isV6, "%q is either both families or neither", in)
	}

	// And neither family accepts what IP rejects.
	for _, in := range []string{"", "nope", "10.0.0.0/8"} {
		require.Error(t, is.IP()(in))
		require.Error(t, is.IPv4()(in))
		require.Error(t, is.IPv6()(in))
	}
}

func ExampleCIDR() {
	fmt.Println(is.CIDR()("10.0.0.0/8"))
	fmt.Println(is.CIDR()("10.0.0.5/24")) // host bits are allowed here
	fmt.Println(is.CIDR()("10.0.0.0"))
	// Output:
	// <nil>
	// <nil>
	// is not a valid CIDR range
}

func ExampleCIDRv4() {
	fmt.Println(is.CIDRv4()("10.0.0.0/8"))
	fmt.Println(is.CIDRv4()("10.0.0.5/24")) // not a network address
	fmt.Println(is.CIDRv4()("2001:db8::/32"))
	// Output:
	// <nil>
	// is not a valid CIDR(v4) range
	// is not a valid CIDR(v4) range
}

func ExampleCIDRv6() {
	fmt.Println(is.CIDRv6()("2001:db8::/32"))
	fmt.Println(is.CIDRv6()("2001:db8::1/32")) // not a network address
	fmt.Println(is.CIDRv6()("10.0.0.0/8"))
	// Output:
	// <nil>
	// is not a valid CIDR(v6) range
	// is not a valid CIDR(v6) range
}

func ExampleIP() {
	fmt.Println(is.IP()("10.0.0.1"))
	fmt.Println(is.IP()("2001:db8::1"))
	fmt.Println(is.IP()("10.0.0.0/8")) // a prefix is not an address
	// Output:
	// <nil>
	// <nil>
	// is not a valid IP address
}

func ExampleIPv4() {
	fmt.Println(is.IPv4()("10.0.0.1"))
	fmt.Println(is.IPv4()("::ffff:10.0.0.1")) // v4 in another spelling
	fmt.Println(is.IPv4()("2001:db8::1"))
	// Output:
	// <nil>
	// <nil>
	// is not a valid IPv4 address
}

func ExampleIPv6() {
	fmt.Println(is.IPv6()("2001:db8::1"))
	fmt.Println(is.IPv6()("10.0.0.1"))
	// Output:
	// <nil>
	// is not a valid IPv6 address
}

// requireCheck runs c against in and asserts on the message, where an empty
// wantErr means the check must pass.
//
// A copy of the helper in the parent package's tests. Two call sites is not
// enough to justify a shared package for eleven lines.
func requireCheck[T any](t *testing.T, c validate.Check[T], in T, wantErr string) {
	t.Helper()

	err := c(in)
	if wantErr == "" {
		require.NoError(t, err)
		return
	}

	require.EqualError(t, err, wantErr)
}
