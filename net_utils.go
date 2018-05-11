package netaddr

import (
	"bytes"
	"net"
)

// containsNet returns true if net2 is a subset of net1. To be clear, it
// returns true if net1 == net2 also.
func containsNet(net1, net2 *net.IPNet) bool {
	// If the two networks are different IP versions, return false
	if len(net1.IP) != len(net2.IP) {
		return false
	}
	if !net1.Contains(net2.IP) {
		return false
	}
	if !net1.IP.Equal(net2.IP) {
		return true
	}
	return bytes.Compare(net1.Mask, net2.Mask) <= 0
}

// netDifference returns the set difference a - b. It returns the list of CIDRs
// in order from largest to smallest. They are *not* sorted by network IP.
func netDifference(a, b *net.IPNet) (result []*net.IPNet) {
	// If the two networks are different IP versions, return a
	if len(a.IP) != len(b.IP) {
		return []*net.IPNet{a}
	}

	// If b contains a then the difference is empty
	if containsNet(b, a) {
		return
	}
	// If a doesn't contain b then the difference is equal to a
	if !containsNet(a, b) {
		return []*net.IPNet{a}
	}

	// If two nets overlap then one must contain the other. At this point, we
	// know a contains b and b is smaller than a. Cut a in half and recurse on
	// the one that overlaps
	first, second := divideNetInHalf(a)
	if bytes.Compare(b.IP, second.IP) < 0 {
		return append([]*net.IPNet{second}, netDifference(first, b)...)
	}
	return append([]*net.IPNet{first}, netDifference(second, b)...)
}

// divideNetInHalf returns the given net as two equally sized halves
func divideNetInHalf(n *net.IPNet) (a, b *net.IPNet) {
	// Get the size of the original netmask
	ones, bits := n.Mask.Size()

	// Netmask has one more 1. Net is half the size of original.
	mask := net.CIDRMask(ones+1, bits)

	// Create a new IP to fill in for the second half
	ip := net.ParseIP("::")
	if bits == 32 {
		ip = net.ParseIP("0.0.0.0").To4()
	}
	// Fill in the new IP
	for i := 0; i < bits/8; i++ {
		// Puts a 1 in the new bit since this is the second half
		extraOne := mask[i] ^ n.Mask[i]
		// New IP is the same as old IP with the extra one at the end
		ip[i] = mask[i] & (n.IP[i] | extraOne)
	}

	a = &net.IPNet{IP: n.IP, Mask: mask}
	b = &net.IPNet{IP: ip, Mask: mask}
	return
}

// canCombineNets returns true if the two networks, a and b, can be combined
// into one larger cidr twice the size. If true, it returns the combined
// network.
func canCombineNets(a, b *net.IPNet) (ok bool, newNet *net.IPNet) {
	if a.IP.Equal(b.IP) {
		return
	}
	if bytes.Compare(a.Mask, b.Mask) != 0 {
		return
	}
	ones, bits := a.Mask.Size()
	newNet = &net.IPNet{IP: a.IP, Mask: net.CIDRMask(ones-1, bits)}
	if newNet.Contains(b.IP) {
		ok = true
		return
	}
	return
}

// ipToNet converts the given IP to a /32 or /128 network depending on the type
// of address.
func ipToNet(ip net.IP) *net.IPNet {
	size := 8 * len(ip)
	return &net.IPNet{IP: ip, Mask: net.CIDRMask(size, size)}
}

// incrementIP returns the given IP + 1
func incrementIP(ip net.IP) (result net.IP) {
	result = net.ParseIP("::")
	if len(ip) == 4 {
		result = net.ParseIP("0.0.0.0").To4()
	}

	carry := true
	for i := len(ip) - 1; i >= 0; i-- {
		result[i] = ip[i]
		if carry {
			result[i]++
			if result[i] != 0 {
				carry = false
			}
		}
	}
	return
}

// expandNet returns a slice containing all of the IPs in the given net up to
// the given limit
func expandNet(n *net.IPNet, limit int) []net.IP {
	ones, bits := n.Mask.Size()

	size := limit
	max := 1 << 30
	if bits-ones < 30 {
		max = 1 << uint(bits-ones)
	}
	if max < size {
		size = max
	}
	result := make([]net.IP, size)
	next := n.IP
	for i := 0; i < size; i++ {
		result[i] = next[:]
		next = incrementIP(next)
	}
	return result
}
