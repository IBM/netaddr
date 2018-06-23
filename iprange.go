package netaddr

// IPRange contains a single contiguous range of IP addresses. A valid range
// cannot be empty, it must have an least one IP address. For this reason, they
// should be created with the CreateRange method below.
type IPRange struct {
}

// CreateRange creates a new range given the two IP addresses passed in. The
// first address must be less than or equal to the last. The two addresses must
// be from the IP family (i.e. ipv4 or ipv6)
func CreateRange(first, last net.IP) (r *IPRange, err error) {
}

// First returns the first IP in the range.
func (s *IPRange) First() net.IP {
}

// First returns the last IP in the range.
func (s *IPRange) Last() net.IP {
}

// ToSet converts the range into a IPSet
func (s *IPRange) ToSet() (set *IPSet) {
}
