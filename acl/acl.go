// Package acl provides primitives for creating and checking permissions
// based on user, group and flags
package acl

import (
	"errors"
	"strings"
)

var ErrPermissionDenied = errors.New("permission denied")

// User is an interface used to check against an ACL
type User interface {
	Name() string
	Groups() []string
	Flags() []string
}

// collection is a container for the three different permission types,
// users, group and flags. Provides utilities for checking if the collection
// contains a provided entity
type collection struct {
	all bool

	users  []string
	groups []string
	flags  []string
}

// ACL provides utilities for checking if a subject has permission to perform
// on an object
type ACL struct {
	allowed collection
	blocked collection
}

// Takes in a string that describes the permissions for an object. Returns an ACL with
// a method for checking permissions. An entity is a user with the following attributes:
// - name
// - list of groups
// - list of flags
//
// When describing permissions use the following (glftpd) syntax:
// - `-` prefix describes a user, i.e. `-userName`
// - `=` prefix describes a group, i.e. `=groupName`
// - no prefix describes a flag, i.e. `1` (currently no restrictions on legnth)
// - `!` prefix denotes that the preceding permission is blocked, i.e. `!-userName` would
// not be allowed
//
// Currently the order of checking is:
// - blocked users
// - blocked groups
// - blocked flags
// - allowed user
// - allowed groups
// - allowed flags
// - blocked all (!*)
// - allowed all (*)
//
// The default is to block permission
func NewFromString(s string) (*ACL, error) {
	if len(s) == 0 {
		return nil, errors.New("no input string given")
	}

	var a ACL

	fields := strings.Fields(strings.ToLower(s))

	var c *collection

	for _, f := range fields {
		if len(f) == 0 {
			continue
		}

		c = &a.allowed

		if f[0] == '!' {
			if len(f) <= 1 {
				return nil, errors.New("expected string after '!'")
			}

			c = &a.blocked

			f = f[1:]
		}

		switch f[0] {
		case '-':
			// user specific acl
			if len(f) <= 1 {
				return nil, errors.New("expected string after '-'")
			}

			f = f[1:]

			if f == "*" {
				c.all = true
			} else {
				c.users = append(c.users, f)
			}

		case '=':
			// group specific acl
			if len(f) <= 1 {
				return nil, errors.New("expected string after '='")
			}

			f = f[1:]

			if f == "*" {
				c.all = true
			} else {
				c.groups = append(c.groups, f)
			}

		default:
			if f == "*" {
				c.all = true
			} else {
				c.groups = append(c.groups, f)
			}
		}
	}

	return &a, nil
}

// has checks to see if the slice contains the provided element (lower cased)
func (c *collection) has(s []string, e string) bool {
	e = strings.ToLower(e)
	for idx := range s {
		if s[idx] == e {
			return true
		}
	}
	return false
}

// hasUser checks to see if the users slices contains the fgiven user
func (c *collection) hasUser(u string) bool {
	return c.has(c.users, u)
}

// hasGroup checks to see if the groups slice contains given group
func (c *collection) hasGroup(g string) bool {
	return c.has(c.groups, g)
}

// hasFlag checks to see if the flags slice contains given flag
func (c *collection) hasFlag(f string) bool {
	return c.has(c.flags, f)
}

// UserAllowed checks to see if given User is allowed or blocked. Default is to
// block access
func (a *ACL) Allowed(u User) bool {
	// check blocked lists
	if a.blocked.hasUser(u.Name()) {
		return false
	}

	groups := u.Groups()
	for idx := range groups {
		if a.blocked.hasGroup(groups[idx]) {
			return false
		}
	}

	flags := u.Flags()
	for idx := range flags {
		if a.blocked.hasFlag(flags[idx]) {
			return false
		}
	}

	// check allowed lists
	if a.allowed.hasUser(u.Name()) {
		return true
	}

	for idx := range groups {
		if a.allowed.hasGroup(groups[idx]) {
			return true
		}
	}

	for idx := range flags {
		if a.allowed.hasFlag(flags[idx]) {
			return true
		}
	}

	// fall back to all flags
	if a.blocked.all {
		return false
	}

	if a.allowed.all {
		return true
	}

	// default is to block access
	return false
}
