package acl

import (
	"fmt"
	"testing"
)

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

func TestNewRule(t *testing.T) {
	var tests = []struct {
		input string
		rule  Rule
		err   string
	}{
		{
			"download /path/test/dir -user !*",
			Rule{
				"/path/test/dir",
				PermissionScopeDownload,
				&ACL{
					collection{false, []string{"user"}, nil},
					collection{true, nil, nil},
				},
			},
			"",
		},
		{
			"download /path/test/dir !-user *",
			Rule{
				"/path/test/dir",
				PermissionScopeDownload,
				&ACL{
					collection{true, nil, nil},
					collection{false, []string{"user"}, nil},
				},
			},
			"",
		},
		{
			"notexist /path/test/dir !-user *",
			Rule{},
			"unknown permission scope 'notexist'",
		},
		{
			"bad",
			Rule{},
			"rule requires minimum of 3 fields",
		},
		{
			"bad line",
			Rule{},
			"rule requires minimum of 3 fields",
		},
		{
			"download /path/test !-*",
			Rule{
				"/path/test",
				PermissionScopeDownload,
				nil,
			},
			"bad user '*'",
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.input,
			func(t *testing.T) {
				rule, err := NewRule(tt.input)
				if err != nil && len(tt.err) == 0 {
					t.Fatalf("expected nil but got: '%s'", err)
				}

				if err != nil && tt.err != err.Error() {
					t.Fatalf("expected '%s' but got: '%s'", tt.err, err)
				}

				if err == nil && len(tt.err) > 0 {
					t.Fatalf("expected '%s' but got nil", tt.err)
				}

				if tt.rule.path != rule.path {
					t.Errorf("expected path to be '%s' but got '%s'", tt.rule.path, rule.path)
				}

				if tt.rule.scope != rule.scope {
					t.Errorf("expected scope to be '%s' but got '%s'", tt.rule.scope, rule.scope)
				}

				if tt.rule.acl != nil && rule.acl != nil {
					if !compareACL(tt.rule.acl, rule.acl) {
						t.Error("acl do not match")
					}
				}
			},
		)
	}
}

func TestNewPermissions(t *testing.T) {
	var tests = []struct {
		lines []string
		err   string
	}{
		{
			[]string{},
			"",
		},
		{
			[]string{
				"download /dir/a *",
				"download /dir/b !*",
			},
			"",
		},
		{
			[]string{
				"download /dir/a *",
				"download /dir/a !*",
			},
			"path '/dir/a' for scope 'download' already exists",
		},
	}

	for idx, tt := range tests {
		t.Run(
			fmt.Sprintf("%d", idx),
			func(t *testing.T) {
				var rules []Rule
				for _, l := range tt.lines {
					r, err := NewRule(l)
					if err != nil {
						t.Fatalf("unable to parse rule '%s': %s", l, err)
					}
					rules = append(rules, r)
				}
				_, err := NewPermissions(rules)
				if err != nil && len(tt.err) == 0 {
					t.Fatalf("expected nil but got: '%s'", err)
				}

				if err != nil && len(tt.err) > 0 && err.Error() != tt.err {
					t.Fatalf("expected '%s' but got: '%s'", tt.err, err)
				}

				if err == nil && len(tt.err) > 0 {
					t.Fatalf("expected '%s' but got nil", tt.err)
				}
			},
		)
	}
}

func TestPermissionsCheck(t *testing.T) {
	var tests = []struct {
		input    string
		path     string
		scope    PermissionScope
		user     TestUser
		expected bool
	}{
		{
			"download /dir/a *",
			"/dir/a",
			PermissionScopeDownload,
			TestUser{"user", nil},
			true,
		},
		{
			"download /dir/a !*",
			"/dir/a",
			PermissionScopeDownload,
			TestUser{"user", nil},
			false,
		},
		{
			"download /dir/a -user !*",
			"/dir/a",
			PermissionScopeDownload,
			TestUser{"user", nil},
			true,
		},
		{
			"download /dir/a =group !*",
			"/dir/a",
			PermissionScopeDownload,
			TestUser{"user", []string{"group"}},
			true,
		},
		{
			"download / =group !*",
			"/dir/a",
			PermissionScopeDownload,
			TestUser{"user", []string{"group"}},
			true,
		},
		{
			"download / =group !*",
			"/dir/a",
			PermissionScopeUpload,
			TestUser{"user", []string{"group"}},
			false,
		},
		{
			"download /some/path =group !*",
			"/dir/a",
			PermissionScopeDownload,
			TestUser{"user", []string{"group"}},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.input,
			func(t *testing.T) {
				r, err := NewRule(tt.input)
				if err != nil {
					t.Fatalf("unable to parse rule '%s': %s", tt.input, err)
				}

				p, err := NewPermissions([]Rule{r})
				if err != nil {
					t.Fatalf("unable to create Permissions: %s", err)
				}

				allowed := p.Allowed(tt.scope, tt.path, tt.user)
				if allowed != tt.expected {
					t.Errorf("expected %t got %t", tt.expected, allowed)
				}
			},
		)
	}
}
