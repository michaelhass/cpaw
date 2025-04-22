package models

import "slices"

type Role string

const (
	AdminRole Role = "admin"
	UserRole  Role = "user"
)

var allRoles = []Role{AdminRole, UserRole}

func AllRoles() []Role {
	return allRoles
}

func (r Role) isValid() bool {
	return slices.Contains(allRoles, r)
}
