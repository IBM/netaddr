package netaddr

import (
	"errors"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateNilTree(t *testing.T) {
	var tree *ipTree
	assert.Equal(t, []error{}, tree.validate())
}

func TestValidateNoNetwork(t *testing.T) {
	tree := &ipTree{}
	assert.Equal(t, []error{
		errors.New("each node in tree must have a network"),
	}, tree.validate())
}

func TestValidateBadCidrBadUp(t *testing.T) {
	_, ten24, _ := net.ParseCIDR("10.0.0.0/24")
	ten24.IP = net.ParseIP("10.0.0.1")
	tree := &ipTree{
		net: ten24,
		up:  &ipTree{},
	}
	assert.Equal(t, []error{
		errors.New("root up must be nil"),
		errors.New("cidr invalid: 10.0.0.1/24"),
	}, tree.validate())
}

func TestValidateBadLinkageLeft(t *testing.T) {
	tree := &ipTree{
		net: TenOne24,
		left: &ipTree{
			net: Ten24,
		},
	}
	assert.Equal(t, []error{
		errors.New("linkage error: left.up node must equal node"),
		errors.New("nodes must be combined: 10.0.0.0/24, 10.0.1.0/24"),
	}, tree.validate())
}

func TestValidateBadLinkageRight(t *testing.T) {
	tree := &ipTree{
		net: Ten24,
		right: &ipTree{
			net: TenOne24,
		},
	}
	assert.Equal(t, []error{
		errors.New("linkage error: right.up node must equal node"),
		errors.New("nodes must be combined: 10.0.0.0/24, 10.0.1.0/24"),
	}, tree.validate())
}

func TestValidateOutOfOrder(t *testing.T) {
	tree := &ipTree{}
	tree.left = &ipTree{
		up:  tree,
		net: TenTwo24,
	}
	tree.net = TenOne24
	tree.right = &ipTree{
		up:  tree,
		net: Ten24,
	}
	tree.right.right = &ipTree{
		up:  tree.right,
		net: Ten24,
	}
	assert.Equal(t, []error{
		errors.New("nodes must be in order: 10.0.2.0 !< 10.0.1.0"),
		errors.New("nodes must be in order: 10.0.1.0 !< 10.0.0.0"),
		errors.New("nodes must be combined: 10.0.1.0/24, 10.0.0.0/24"),
		errors.New("nodes must be in order: 10.0.0.0 !< 10.0.0.0"),
	}, tree.validate())
}
