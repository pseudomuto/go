package validate

import (
	"errors"
	"net"
)

// IsCIDR reports strings that are not a CIDR prefix, of either family.
//
// It accepts whatever [net.ParseCIDR] accepts, which means host bits may be set:
// "10.0.0.5/24" passes here. [IsCIDRv4] and [IsCIDRv6] are stricter on that
// point as well as pinning the family.
func IsCIDR() Check[string] {
	return func(s string) error {
		if _, _, err := net.ParseCIDR(s); err != nil {
			return errors.New("is not a valid CIDR range")
		}

		return nil
	}
}

// IsCIDRv4 reports strings that are not an IPv4 network address in CIDR form.
//
// Two conditions, both required. The prefix must be IPv4, and it must name the
// network rather than a host inside it, so every bit outside the mask has to be
// zero. "10.0.0.0/24" passes; "10.0.0.5/24" does not, even though
// [net.ParseCIDR] is happy with it. [IsCIDR] is the loose version.
//
// Family is decided after [net.ParseCIDR] normalises the address, so an
// IPv4-mapped IPv6 prefix such as "::ffff:10.0.0.0/104" counts as IPv4 here and
// is rejected by [IsCIDRv6].
func IsCIDRv4() Check[string] {
	return func(s string) error {
		ip, ipn, err := net.ParseCIDR(s)
		if err != nil || ip.To4() == nil || !ipn.IP.Equal(ip) {
			return errors.New("is not a valid CIDR(v4) range")
		}

		return nil
	}
}

// IsCIDRv6 reports strings that are not an IPv6 network address in CIDR form.
//
// The mirror of [IsCIDRv4]: the prefix must be IPv6, and it must name the
// network rather than a host inside it. "2001:db8::/32" passes;
// "2001:db8::1/32" does not.
//
// IPv4-mapped prefixes such as "::ffff:10.0.0.0/104" are rejected, since
// [net.ParseCIDR] normalises them to IPv4.
func IsCIDRv6() Check[string] {
	return func(s string) error {
		ip, ipn, err := net.ParseCIDR(s)
		if err != nil || ip.To4() != nil || !ipn.IP.Equal(ip) {
			return errors.New("is not a valid CIDR(v6) range")
		}

		return nil
	}
}

// IsIP reports strings that are not an IP address, of either family.
//
// It accepts whatever [net.ParseIP] accepts, so both "10.0.0.1" and "2001:db8::1"
// pass. A prefix such as "10.0.0.0/8" does not; that is [IsCIDR]'s job.
func IsIP() Check[string] {
	return func(s string) error {
		if ip := net.ParseIP(s); ip == nil {
			return errors.New("is not a valid IP address")
		}

		return nil
	}
}

// IsIPv4 reports strings that are not an IPv4 address.
//
// An IPv4-mapped IPv6 address such as "::ffff:10.0.0.1" passes, because it is
// IPv4 in a different spelling. [IsIPv6] rejects the same string, so the two are
// complements over everything [IsIP] accepts.
func IsIPv4() Check[string] {
	return func(s string) error {
		if ip := net.ParseIP(s); ip == nil || ip.To4() == nil {
			return errors.New("is not a valid IPv4 address")
		}

		return nil
	}
}

// IsIPv6 reports strings that are not an IPv6 address.
//
// IPv4-mapped addresses such as "::ffff:10.0.0.1" are rejected here and accepted
// by [IsIPv4]; see there for why.
func IsIPv6() Check[string] {
	return func(s string) error {
		if ip := net.ParseIP(s); ip == nil || ip.To4() != nil {
			return errors.New("is not a valid IPv6 address")
		}

		return nil
	}
}
