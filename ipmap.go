package netaddr

import (
	"fmt"
	"net"

	"github.com/ecbaldwin/trie"
)

// IPMap is a structure that maps IP prefixes to values. For example, you can
// insert the following values and they will all exist as distinct prefix/value
// pairs in the map.
//
// 10.0.0.0/16 -> 1
// 10.0.0.0/24 -> 1
// 10.0.0.0/32 -> 2
//
// The map supports looking up values based on a longest prefix match and also
// supports efficient aggregation of prefix/value pairs based on equality of
// values. See the README.md file for a more detailed discussion..
type IPMap struct {
	length uint
	trie   trie.Trie
}

// NewIPv4Map returns a new map where the prefixes are 4-byte IPv4 prefixes.
func NewIPv4Map() *IPMap {
	return &IPMap{
		length: net.IPv4len,
	}
}

// NewIPv6Map returns a new map where the prefixes are 16-byte IPv6 prefixes.
func NewIPv6Map() *IPMap {
	return &IPMap{
		length: net.IPv6len,
	}
}

// Size returns the number of exact prefixes stored in the map
func (m *IPMap) Size() int {
	return m.trie.Size()
}

// InsertPrefix inserts the given prefix with the given value into the map
func (m *IPMap) InsertPrefix(prefix *net.IPNet, value interface{}) error {
	if prefix == nil {
		return fmt.Errorf("cannot insert nil prefix")
	}
	if uint(len(prefix.IP)) != m.length {
		return fmt.Errorf("cannot insert prefix with length %d in map with length %d", len(prefix.IP), m.length)
	}

	return m.trie.Insert(
		prefixToKey(prefix),
		value,
	)
}

// Insert is a convenient alternative to InsertPrefix that treats the given IP
// address as a host prefix (i.e. /32 for IPv4 and /128 for IPv6)
func (m *IPMap) Insert(ip net.IP, value interface{}) error {
	if uint(len(ip)) != m.length {
		return fmt.Errorf("cannot insert IP with length %d in map with length %d", len(ip), m.length)
	}

	return m.trie.Insert(
		ipToKey(ip),
		value,
	)
}

// InsertOrUpdatePrefix inserts the given prefix with the given value into the map.
// If the prefix already existed, it updates the associated value in place.
func (m *IPMap) InsertOrUpdatePrefix(prefix *net.IPNet, value interface{}) error {
	if prefix == nil {
		return fmt.Errorf("cannot insert nil prefix")
	}
	if uint(len(prefix.IP)) != m.length {
		return fmt.Errorf("cannot insert prefix with length %d in map with length %d", len(prefix.IP), m.length)
	}

	return m.trie.InsertOrUpdate(
		prefixToKey(prefix),
		value,
	)
}

// InsertOrUpdate is a convenient alternative to InsertOrUpdatePrefix that treats
// the given IP address as a host prefix (i.e. /32 for IPv4 and /128 for IPv6)
func (m *IPMap) InsertOrUpdate(ip net.IP, value interface{}) error {
	if uint(len(ip)) != m.length {
		return fmt.Errorf("cannot insert IP with length %d in map with length %d", len(ip), m.length)
	}

	return m.trie.InsertOrUpdate(
		ipToKey(ip),
		value,
	)
}

// GetPrefix returns the value in the map associated with the given network prefix
// with an exact match: both the IP and the prefix length must match. If an
// exact match is not found, found is false and value is nil and should be
// ignored.
func (m *IPMap) GetPrefix(prefix *net.IPNet) (interface{}, bool) {
	if prefix == nil {
		return nil, false
	}
	if uint(len(prefix.IP)) != m.length {
		return nil, false
	}

	key := prefixToKey(prefix)
	match, _, value := m.trie.Match(key)

	if match == trie.MatchExact {
		return value, true
	}

	return nil, false
}

// Get is a convenient alternative to GetPrefix that treats the given IP address
// as a host prefix (i.e. /32 for IPv4 and /128 for IPv6)
func (m *IPMap) Get(ip net.IP) (interface{}, bool) {
	if uint(len(ip)) != m.length {
		return nil, false
	}

	key := ipToKey(ip)
	match, _, value := m.trie.Match(key)

	if match == trie.MatchExact {
		return value, true
	}

	return nil, false
}

// GetOrInsertPrefix returns the value associated with the given prefix if it
// already exists. If it does not exist, it inserts it with the given value and
// returns that.
func (m *IPMap) GetOrInsertPrefix(prefix *net.IPNet, value interface{}) (interface{}, error) {
	if prefix == nil {
		return nil, fmt.Errorf("cannot insert nil prefix")
	}
	if uint(len(prefix.IP)) != m.length {
		return nil, fmt.Errorf("cannot insert prefix with length %d in map with length %d", len(prefix.IP), m.length)
	}

	key := prefixToKey(prefix)
	return m.trie.GetOrInsert(key, value)
}

// GetOrInsert is a convenient alternative to GetOrInsertPrefix that treats the
// given IP address as a host prefix (i.e. /32 for IPv4 and /128 for IPv6)
func (m *IPMap) GetOrInsert(ip net.IP, value interface{}) (interface{}, error) {
	if uint(len(ip)) != m.length {
		return nil, fmt.Errorf("cannot insert IP with length %d in map with length %d", len(ip), m.length)
	}

	key := ipToKey(ip)
	return m.trie.GetOrInsert(key, value)
}

// MatchPrefix returns the value in the map associated with the given network
// prefix using a longest prefix match. If a match is found, it returns a
// *net.IPNet representing the longest prefix matched. If a match is *not*
// found, the prefix returned is nil; value is also nil and should be ignored.
func (m *IPMap) MatchPrefix(prefix *net.IPNet) (*net.IPNet, interface{}) {
	if prefix == nil {
		return nil, false
	}
	if uint(len(prefix.IP)) != m.length {
		return nil, false
	}

	key := prefixToKey(prefix)
	match, matchKey, value := m.trie.Match(key)

	if match == trie.MatchNone {
		return nil, false
	}

	return keyToPrefix(matchKey, m.length), value
}

// Match is a convenient alternative to MatchPrefix that treats the given IP
// address as a host prefix (i.e. /32 for IPv4 and /128 for IPv6)
func (m *IPMap) Match(ip net.IP) (*net.IPNet, interface{}) {
	if uint(len(ip)) != m.length {
		return nil, false
	}

	key := ipToKey(ip)
	match, matchKey, value := m.trie.Match(key)

	if match == trie.MatchNone {
		return nil, false
	}

	return keyToPrefix(matchKey, m.length), value
}

// RemovePrefix removes the given prefix from the map with its associated value.
// Only a prefix with an exact match will be removed.
func (m *IPMap) RemovePrefix(prefix *net.IPNet) {
	if prefix == nil {
		return
	}
	if uint(len(prefix.IP)) != m.length {
		return
	}

	m.trie.Delete(prefixToKey(prefix))
}

// Remove is a convenient alternative to RemovePrefix that treats the given IP
// address as a host prefix (i.e. /32 for IPv4 and /128 for IPv6)
func (m *IPMap) Remove(ip net.IP) {
	if uint(len(ip)) != m.length {
		return
	}

	m.trie.Delete(ipToKey(ip))
}

// Callback is the signature of the callback functions that can be passed to
// Iterate or Aggregate to handle each prefix/value combination.
type Callback func(prefix *net.IPNet, value interface{}) bool

// Iterate invokes the given callback function for each prefix/value pair in
// the map in lexigraphical order.
func (m *IPMap) Iterate(callback Callback) bool {
	return m.trie.Iterate(trieCallback(m, callback))
}

// Aggregate invokes then given callback function for each prefix/value pair in
// the map, aggregated by value, in lexigraphical order.
//
// 1. The values stored must be comparable to be aggregable. Prefixes get
//    aggregated only where their values compare equal.
// 2. The set of prefix/value pairs visited is the minimal set such that any
//    longest prefix match against the aggregated set will always return the
//    same value as the same match against the non-aggregated set.
// 3. The aggregated and non-aggregated sets of prefixes may be disjoint.
func (m *IPMap) Aggregate(callback Callback) bool {
	return m.trie.Aggregate(trieCallback(m, callback))
}

func ipToKey(ip net.IP) *trie.Key {
	return &trie.Key{
		Length: uint(8 * len(ip)),
		Bits:   ip,
	}
}

func prefixToKey(prefix *net.IPNet) *trie.Key {
	ones, _ := prefix.Mask.Size()
	return &trie.Key{
		Length: uint(ones),
		Bits:   prefix.IP,
	}
}

func keyToPrefix(key *trie.Key, length uint) *net.IPNet {
	// The trie implementation may not store a full 4 or 16 bytes if the prefix
	// length is shorter. But, we want the full size when creating a net.IP.
	ip := NewIP(int(length))
	copy(ip, key.Bits)

	return &net.IPNet{
		IP: ip,
		Mask: net.CIDRMask(
			int(key.Length),
			8*int(length),
		),
	}
}

func trieCallback(m *IPMap, callback Callback) trie.Callback {
	return func(key *trie.Key, value interface{}) bool {
		return callback(keyToPrefix(key, m.length), value)
	}
}
