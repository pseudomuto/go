package validate_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/pseudomuto/go/validate"
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

func TestIsCIDR(t *testing.T) {
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

			requireCheck(t, validate.IsCIDR(), tt.in, tt.wantErr)
		})
	}
}

func TestIsCIDRv4(t *testing.T) {
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

			requireCheck(t, validate.IsCIDRv4(), tt.in, tt.wantErr)
		})
	}
}

func TestIsCIDRv6(t *testing.T) {
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

			requireCheck(t, validate.IsCIDRv6(), tt.in, tt.wantErr)
		})
	}
}

// The documented difference between IsCIDR and the two family checks is that
// only the latter insist on a network address. Pin that down as one statement
// rather than leaving it implied by three tables.
func TestCIDRFamilyChecksRejectHostBitsThatIsCIDRAllows(t *testing.T) {
	t.Parallel()

	for _, in := range []string{"10.0.0.5/24", "2001:db8::1/32"} {
		require.NoError(t, validate.IsCIDR()(in), "IsCIDR should not care about host bits")
		require.Error(t, validate.IsCIDRv4()(in))
		require.Error(t, validate.IsCIDRv6()(in))
	}
}

func TestIsIP(t *testing.T) {
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

			requireCheck(t, validate.IsIP(), tt.in, tt.wantErr)
		})
	}
}

func TestIsIPv4(t *testing.T) {
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

			requireCheck(t, validate.IsIPv4(), tt.in, tt.wantErr)
		})
	}
}

func TestIsIPv6(t *testing.T) {
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

			requireCheck(t, validate.IsIPv6(), tt.in, tt.wantErr)
		})
	}
}

// IsIPv4 and IsIPv6 are documented as complements over everything IsIP accepts,
// which only holds if the v4-mapped case lands on exactly one side.
func TestIPFamilyChecksPartitionIsIP(t *testing.T) {
	t.Parallel()

	accepted := []string{"10.0.0.1", "0.0.0.0", "255.255.255.255", "2001:db8::1", "::1", "::", "::ffff:10.0.0.1"}

	for _, in := range accepted {
		require.NoError(t, validate.IsIP()(in), "IsIP rejected %q, so the rest of this says nothing", in)

		isV4 := validate.IsIPv4()(in) == nil
		isV6 := validate.IsIPv6()(in) == nil

		require.NotEqual(t, isV4, isV6, "%q is either both families or neither", in)
	}

	// And neither family accepts what IsIP rejects.
	for _, in := range []string{"", "nope", "10.0.0.0/8"} {
		require.Error(t, validate.IsIP()(in))
		require.Error(t, validate.IsIPv4()(in))
		require.Error(t, validate.IsIPv6()(in))
	}
}

func ExampleIsCIDR() {
	fmt.Println(validate.IsCIDR()("10.0.0.0/8"))
	fmt.Println(validate.IsCIDR()("10.0.0.5/24")) // host bits are allowed here
	fmt.Println(validate.IsCIDR()("10.0.0.0"))
	// Output:
	// <nil>
	// <nil>
	// is not a valid CIDR range
}

func ExampleIsCIDRv4() {
	fmt.Println(validate.IsCIDRv4()("10.0.0.0/8"))
	fmt.Println(validate.IsCIDRv4()("10.0.0.5/24")) // not a network address
	fmt.Println(validate.IsCIDRv4()("2001:db8::/32"))
	// Output:
	// <nil>
	// is not a valid CIDR(v4) range
	// is not a valid CIDR(v4) range
}

func ExampleIsCIDRv6() {
	fmt.Println(validate.IsCIDRv6()("2001:db8::/32"))
	fmt.Println(validate.IsCIDRv6()("2001:db8::1/32")) // not a network address
	fmt.Println(validate.IsCIDRv6()("10.0.0.0/8"))
	// Output:
	// <nil>
	// is not a valid CIDR(v6) range
	// is not a valid CIDR(v6) range
}

func ExampleIsIP() {
	fmt.Println(validate.IsIP()("10.0.0.1"))
	fmt.Println(validate.IsIP()("2001:db8::1"))
	fmt.Println(validate.IsIP()("10.0.0.0/8")) // a prefix is not an address
	// Output:
	// <nil>
	// <nil>
	// is not a valid IP address
}

func ExampleIsIPv4() {
	fmt.Println(validate.IsIPv4()("10.0.0.1"))
	fmt.Println(validate.IsIPv4()("::ffff:10.0.0.1")) // v4 in another spelling
	fmt.Println(validate.IsIPv4()("2001:db8::1"))
	// Output:
	// <nil>
	// <nil>
	// is not a valid IPv4 address
}

func ExampleIsIPv6() {
	fmt.Println(validate.IsIPv6()("2001:db8::1"))
	fmt.Println(validate.IsIPv6()("10.0.0.1"))
	// Output:
	// <nil>
	// is not a valid IPv6 address
}
