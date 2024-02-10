// Copyright (c) 2020-2024 Richard Cooper
//
// This file is a part of quenbyako/ext package.
// See https://github.com/quenbyako/ext/blob/master/LICENSE for details

// Package user allows user account lookups by name or id.
//
// For most Unix systems, this package has two internal implementations of
// resolving user and group ids to names, and listing supplementary group IDs.
// One is written in pure Go and parses /etc/passwd and /etc/group. The other
// is cgo-based and relies on the standard C library (libc) routines such as
// getpwuid_r, getgrnam_r, and getgrouplist.
//
// When cgo is available, and the required routines are implemented in libc
// for a particular platform, cgo-based (libc-backed) code is used.
// This can be overridden by using osusergo build tag, which enforces
// the pure Go implementation.
//
// # Stdlib os/user package extension
//
// This package extends standard library package, so all API is backward
// compatible.
//
// However, biggest change is new field `Groups` in `User` structure, because
// stdlib [User.GroupIds] implementation implicitly calls os functionality,
// which is forbidden, if you need to provide custom user provider (not only
// os-related).
package user

import (
	stduser "os/user"
	"sort"
)

// User represents a user account.
type User struct {
	// Uid is the user ID.
	// On POSIX systems, this is a decimal number representing the uid.
	// On Windows, this is a security identifier (SID) in a string format.
	// On Plan 9, this is the contents of /dev/user.
	Uid string
	// Gid is the primary group ID.
	// On POSIX systems, this is a decimal number representing the gid.
	// On Windows, this is a SID in a string format.
	// On Plan 9, this is the contents of /dev/user.
	Gid string
	// Username is the login name.
	Username string
	// Name is the user's real or display name.
	// It might be blank.
	// On POSIX systems, this is the first (or only) entry in the GECOS field
	// list.
	// On Windows, this is the user's display name.
	// On Plan 9, this is the contents of /dev/user.
	Name string
	// HomeDir is the path to the user's home directory (if they have one).
	HomeDir string
	// GroupIDs is an additional field unlike in stduser, cause standard library
	// can't get groups from custom provider/resource.
	//
	// GroupIDs expected to be sorted for easier searching. So you can use
	// binary search
	GroupIDs []string
}

func Current() (*User, error) {
	u, err := stduser.Current()
	if err != nil {
		return nil, err
	}

	// Quick note, why we shouldn't cache this value: in the runtime, user can
	// change their groups, but can't change uid and gid until they'll exit from
	// all processes. That's why sdtlib user caches all fields in current user,
	// but calls syscall `getgroups` every time.
	//
	// Don't worry, this syscall is super fast, so it's not necessary to cache
	// this value.
	groups, err := u.GroupIds()
	if err != nil {
		return nil, err
	}

	return &User{
		Uid:      u.Uid,
		Gid:      u.Gid,
		Username: u.Username,
		Name:     u.Name,
		HomeDir:  u.HomeDir,
		GroupIDs: groups,
	}, nil
}

// Lookup looks up a user by username. If the user cannot be found, the returned
// error is of type UnknownUserError.
func Lookup(username string) (*User, error) {
	u, err := stduser.Lookup(username)
	if err != nil {
		return nil, err
	}
	groups, err := u.GroupIds()
	if err != nil {
		return nil, err
	}
	sort.Strings(groups)

	return &User{
		Uid:      u.Uid,
		Gid:      u.Gid,
		Username: u.Username,
		Name:     u.Name,
		HomeDir:  u.HomeDir,
		GroupIDs: groups,
	}, nil
}

// LookupId looks up a user by userid. If the user cannot be found, the returned
// error is of type UnknownUserIdError.
func LookupId(uid string) (*User, error) {
	u, err := stduser.LookupId(uid)
	if err != nil {
		return nil, err
	}
	groups, err := u.GroupIds()
	if err != nil {
		return nil, err
	}

	return &User{
		Uid:      u.Uid,
		Gid:      u.Gid,
		Username: u.Username,
		Name:     u.Name,
		HomeDir:  u.HomeDir,
		GroupIDs: groups,
	}, nil
}

// GroupIds returns the list of group IDs that the user is a member of.
//
// DEPRECATED: Use [User.Groups] field, to avoid any problems related to stdlib
// implementation.
func (u *User) GroupIds() ([]string, error) { return u.GroupIDs, nil }

func (u *User) StdUser() *stduser.User {
	return &stduser.User{
		Uid:      u.Uid,
		Gid:      u.Gid,
		Username: u.Username,
		Name:     u.Name,
		HomeDir:  u.HomeDir,
	}
}

// Group represents a grouping of users.
//
// On POSIX systems Gid contains a decimal number representing the group ID.
type Group = stduser.Group

// LookupGroup looks up a group by name. If the group cannot be found, the
// returned error is of type UnknownGroupError.
func LookupGroup(name string) (*Group, error) { return stduser.LookupGroup(name) }

// LookupGroupId looks up a group by groupid. If the group cannot be found, the
// returned error is of type UnknownGroupIdError.
func LookupGroupId(gid string) (*Group, error) { return stduser.LookupGroupId(gid) }

type (
	// UnknownUserIdError is returned by LookupId when a user cannot be found.
	UnknownUserIdError = stduser.UnknownUserIdError

	// UnknownUserError is returned by Lookup when a user cannot be found.
	UnknownUserError = stduser.UnknownUserError

	// UnknownGroupIdError is returned by LookupGroupId when a group cannot be
	// found.
	UnknownGroupIdError = stduser.UnknownGroupIdError

	// UnknownGroupError is returned by LookupGroup when a group cannot be
	// found.
	UnknownGroupError = stduser.UnknownGroupError
)
