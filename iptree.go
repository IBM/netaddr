package netaddr

import (
	"bytes"
	"math/big"
	"net"
)

type ipTree struct {
	net             *net.IPNet
	left, right, up *ipTree
}

// setLeft helps maintain the bidirectional relationships in the tree. Always
// use it to set the left child of a node.
func (me *ipTree) setLeft(child *ipTree) {
	if me.left != nil && me == me.left.up {
		me.left.up = nil
	}
	me.left = child
	if child != nil {
		child.up = me
	}
}

// setRight helps maintain the bidirectional relationships in the tree. Always
// use it to set the right child of a node.
func (me *ipTree) setRight(child *ipTree) {
	if me.right != nil && me == me.right.up {
		me.right.up = nil
	}
	me.right = child
	if child != nil {
		child.up = me
	}
}

// trimLeft trims CIDRs that overlap top from the left child
func (me *ipTree) trimLeft(top *ipTree) *ipTree {
	if me == nil {
		return nil
	}

	if containsNet(top.net, me.net) {
		return me.left.trimLeft(top)
	}
	me.setRight(me.right.trimLeft(top))
	return me
}

// trimRight trims CIDRs that overlap top from the right child
func (me *ipTree) trimRight(top *ipTree) *ipTree {
	if me == nil {
		return nil
	}

	if containsNet(top.net, me.net) {
		return me.right.trimRight(top)
	}
	me.setLeft(me.left.trimRight(top))
	return me
}

// insert adds the given node to the tree if its CIDR is not already in the
// set. The new node's CIDR is added in the correct spot and any existing
// subsets are removed from the tree. This method does not optimize the tree by
// adding CIDRs that can be combined.
func (me *ipTree) insert(newNode *ipTree) *ipTree {
	if me == nil {
		return newNode
	}

	if containsNet(me.net, newNode.net) {
		return me
	}

	if containsNet(newNode.net, me.net) {
		// Replace the current top node and trim the tree
		newNode.setLeft(me.left.trimLeft(newNode))
		newNode.setRight(me.right.trimRight(newNode))

		// Check the left-most leaf to see if it can be combined with this one
		return newNode
	}

	if bytes.Compare(newNode.net.IP, me.net.IP) < 0 {
		me.setLeft(me.left.insert(newNode))
	} else {
		me.setRight(me.right.insert(newNode))
	}
	return me
}

// contains returns true if the given IP is in the set.
func (me *ipTree) contains(newNode *ipTree) bool {
	if me == nil || newNode == nil {
		return false
	}

	if containsNet(me.net, newNode.net) {
		return true
	}
	if containsNet(newNode.net, me.net) {
		return false
	}
	if bytes.Compare(newNode.net.IP, me.net.IP) < 0 {
		return me.left.contains(newNode)
	}
	return me.right.contains(newNode)
}

// remove takes out the node and adjusts the tree recursively
func (me *ipTree) remove() *ipTree {
	replaceMe := func(newChild *ipTree) *ipTree {
		if me.up != nil {
			if me == me.up.left {
				me.up.setLeft(newChild)
			} else {
				me.up.setRight(newChild)
			}
		}
		return newChild
	}

	if me.left != nil && me.right != nil {
		next := me.next()
		me.net = next.net
		next.remove()
		return me
	}
	if me.left != nil {
		return replaceMe(me.left)
	}
	if me.right != nil {
		return replaceMe(me.right)
	}
	return replaceMe(nil)
}

// removeNet removes all of the IPs in the given net from the set
func (me *ipTree) removeNet(net *net.IPNet) (top *ipTree) {
	if me == nil {
		return
	}
	// If net starts before me.net, recursively remove net from the left
	if bytes.Compare(net.IP, me.net.IP) < 0 {
		me.left = me.left.removeNet(net)
	}

	// If any CIDRs in `net - me.net` come after me.net, remove net from
	// the right
	diff := netDifference(net, me.net)
	for _, n := range diff {
		if bytes.Compare(me.net.IP, n.IP) < 0 {
			me.right = me.right.removeNet(net)
			break
		}
	}

	top = me
	if containsNet(net, me.net) {
		// Remove the current node
		top = me.remove()
	} else if containsNet(me.net, net) {
		diff = netDifference(me.net, net)
		me.net = diff[0]
		for _, n := range diff[1:] {
			top = top.insert(&ipTree{net: n})
		}
	}
	return
}

// first returns the first node in the tree or nil if there are none. It is
// always the left-most node.
func (me *ipTree) first() *ipTree {
	if me == nil {
		return nil
	}
	if me.left == nil {
		return me
	}
	return me.left.first()
}

// next returns the node following the given one in order or nil if it is the last.
func (me *ipTree) next() *ipTree {
	if me.right != nil {
		next := me.right
		for next.left != nil {
			next = next.left
		}
		return next
	}

	next := me
	for next.up != nil {
		if next.up.left == next {
			return next.up
		}
		next = next.up
	}
	return nil
}

// prev returns the node preceding the given one in order or nil if it is the first.
func (me *ipTree) prev() *ipTree {
	if me.left != nil {
		prev := me.left
		for prev.right != nil {
			prev = prev.right
		}
		return prev
	}

	prev := me
	for prev.up != nil {
		if prev.up.right == prev {
			return prev.up
		}
		prev = prev.up
	}
	return nil
}

// walk visits all of the nodes in order by passing each node, in turn, to the
// given visit function.
func (me *ipTree) walk(visit func(*ipTree)) {
	if me == nil {
		return
	}
	me.left.walk(visit)
	visit(me)
	me.right.walk(visit)
}

// size returns the number of IPs in the set.
// It isn't efficient and only meant for testing.
func (me *ipTree) size() *big.Int {
	s := big.NewInt(0)
	if me != nil {
		ones, bits := me.net.Mask.Size()
		s.Lsh(big.NewInt(1), uint(bits-ones))
		s.Add(s, me.left.size())
		s.Add(s, me.right.size())
	}
	return s
}

// height returns the length of the maximum path from top node to leaf
// It isn't efficient and only meant for testing.
func (me *ipTree) height() uint {
	if me == nil {
		return 0
	}

	s := me.left.height()
	if s < me.right.height() {
		s = me.right.height()
	}
	return s + 1
}

// numNodes Return the number of nodes in the underlying tree It isn't
// efficient and only meant for testing.
func (me *ipTree) numNodes() int {
	if me == nil {
		return 0
	}
	return 1 + me.left.numNodes() + me.right.numNodes()
}
