package acl

import "testing"

func newTestUser(name string, groups ...string) *User {
	return &User{
		name:   name,
		groups: groups,
	}
}

func checkErr(t *testing.T, got, expected error) {
	if got == nil {
		if expected != nil {
			t.Fatalf("expected '%s' but got nil", expected)
			return
		}
		return
	}

	if expected == nil {
		t.Fatalf("unexpected error '%s'", got)
		return
	}
}

func compareACL(a, b *ACL) bool {
	if !compareSlices(a.allowed.users, b.allowed.users) {
		return false
	}

	if !compareSlices(a.allowed.groups, b.allowed.groups) {
		return false
	}

	return true
}

func compareSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	for _, i := range a {
		var match bool
		for _, j := range b {
			if i == j {
				match = true
				break
			}
		}

		if !match {
			return false
		}
	}

	return true
}
