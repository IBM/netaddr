package netaddr

import (
	"net"
)

// IPSet is a set of IP addresses
type IPSet struct {
	tree *ipTree
}

// InsertNet ensures this IPSet has the entire given IP network
func (me *IPSet) InsertNet(net *net.IPNet) {
	if net == nil {
		return
	}

	newNet := net
	for {
		newNode := &ipTree{net: newNet}
		me.tree = me.tree.insert(newNode)

		if me.tree != newNode && newNode.up == nil {
			break
		}

		// The new node was inserted. See if it can be combined with the previous and/or next ones
		prev := newNode.prev()
		if prev != nil {
			if ok, n := canCombineNets(prev.net, newNet); ok {
				newNet = n
			}
		}
		next := newNode.next()
		if next != nil {
			if ok, n := canCombineNets(newNet, next.net); ok {
				newNet = n
			}
		}
		if newNet == newNode.net {
			break
		}
	}
}

// RemoveNet ensures that all of the IPs in the given network are removed from
// the set if present.
func (me *IPSet) RemoveNet(net *net.IPNet) {
	if net == nil {
		return
	}

	me.tree = me.tree.removeNet(net)
}

// ContainsNet returns true iff this IPSet contains all IPs in the given network
func (me *IPSet) ContainsNet(net *net.IPNet) bool {
	if me == nil || net == nil {
		return false
	}
	return me.tree.contains(&ipTree{net: net})
}

// Insert ensures this IPSet has the given IP
func (me *IPSet) Insert(ip net.IP) {
	me.InsertNet(ipToNet(ip))
}

// Remove ensures this IPSet does not contain the given IP
func (me *IPSet) Remove(ip net.IP) {
	me.RemoveNet(ipToNet(ip))
}

// Contains returns true iff this IPSet contains the the given IP address
func (me *IPSet) Contains(ip net.IP) bool {
	return me.ContainsNet(ipToNet(ip))
}

// Union computes the union of this IPSet and another set. It returns the
// result as a new set.
func (me *IPSet) Union(other *IPSet) (newSet *IPSet) {
	newSet = &IPSet{}
	me.tree.walk(func(node *ipTree) {
		newSet.InsertNet(node.net)
	})
	other.tree.walk(func(node *ipTree) {
		newSet.InsertNet(node.net)
	})
	return
}

// Difference computes the set difference between this IPSet and another one
// It returns the result as a new set.
func (me *IPSet) Difference(other *IPSet) (newSet *IPSet) {
	newSet = &IPSet{}
	me.tree.walk(func(node *ipTree) {
		newSet.InsertNet(node.net)
	})
	other.tree.walk(func(node *ipTree) {
		newSet.RemoveNet(node.net)
	})
	return
}

// GetIPs retrieves a slice of the first IPs in the set ordered by address up
// to the given limit.
func (me *IPSet) GetIPs(limit int) (ips []net.IP) {
	if limit == 0 {
		limit = int(^uint(0) >> 1) // MaxInt
	}
	for node := me.tree.first(); node != nil; node = node.next() {
		ips = append(ips, expandNet(node.net, limit-len(ips))...)
	}
	return
}
