// Copyright (c) 2020-2024 Richard Cooper
//
// This file is a part of quenbyako/ext package.
// See https://github.com/quenbyako/ext/blob/master/LICENSE for details

package user

import (
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
)

// CheckGroupExists is a wrapper for LookupGroup which also returns hinter error
// with hint that function caller must create provided group and add themself to
// it.
func GetGroupWithHint(users Provider, name string) (*Group, error) { //cover:ignore // alias
	group, err := users.LookupGroup(name)
	if e := UnknownGroupError(""); errors.As(err, &e) {
		return nil, ErrGroupNotExist(name)
	} else if err != nil {
		return nil, err
	}

	return group, nil
}

func AssertUsersInGroup(p Provider, groups map[string][]string) (errs error) { //cover:ignore // alias
	for group, users := range groups {
		g, err := GetGroupWithHint(p, group)
		if err != nil {
			errs = errors.Join(errs, err)
			continue
		}

		for _, user := range users {
			if u, err := p.Lookup(user); err != nil {
				errs = errors.Join(errs, err)
			} else if !slices.Contains(u.GroupIDs, g.Gid) {
				errs = errors.Join(err, ErrUserNotInGroup{User: user, ExpectedGroup: group})
			}
		}
	}

	return errs
}

// returns errdefs.Multi only.
func AssertCurrentUserInGroups(p Provider, groups ...string) (errs error) { //cover:ignore // alias
	u, err := p.Current()
	if err != nil {
		return errors.Join(errs, err)
	}

	for _, group := range groups {
		if g, err := GetGroupWithHint(p, group); err != nil {
			errs = errors.Join(errs, err)
		} else if !slices.Contains(u.GroupIDs, g.Gid) {
			errs = errors.Join(err, ErrUserNotInGroup{User: u.Username, ExpectedGroup: group})
		}
	}

	return errs
}

// Provider is a special interface which allows you to get unix-compatible
// user data from any source you want (not only your local machine).
//
// Provider MUST return same errors as [os/user] package.
type Provider interface {
	Current() (*User, error)
	Lookup(username string) (*User, error)
	LookupID(uid string) (*User, error)
	LookupGroup(name string) (*Group, error)
	LookupGroupID(gid string) (*Group, error)
}

// OSProvider is a default UserProvider, based on stdlib, and providing data
// from local host.
type OSProvider struct{}

var _ Provider = OSProvider{}

func (OSProvider) Current() (*User, error)                  { return Current() }
func (OSProvider) Lookup(username string) (*User, error)    { return Lookup(username) }
func (OSProvider) LookupID(uid string) (*User, error)       { return LookupId(uid) }
func (OSProvider) LookupGroup(name string) (*Group, error)  { return LookupGroup(name) }
func (OSProvider) LookupGroupID(gid string) (*Group, error) { return LookupGroupId(gid) }

type TestProvider struct {
	CurrentID int
	Users     map[int]*User
	Groups    map[int]string
	UserGIDs  map[int][]string
}

func (p *TestProvider) Current() (*User, error) {
	return p.LookupID(strconv.Itoa(p.CurrentID))
}

func (p *TestProvider) Lookup(username string) (*User, error) {
	for _, u := range p.Users {
		if u.Username == username {
			return u, nil
		}
	}
	return nil, UnknownUserError(username)
}

func (p *TestProvider) LookupID(uid string) (*User, error) {
	id, err := strconv.Atoi(uid)
	if err != nil {
		return nil, err
	}

	u, ok := p.Users[id]
	if !ok {
		return nil, UnknownUserIdError(id)
	}

	// if you didn't provide uid in users, this line will do it for you 😘
	u.Uid = uid

	return u, nil
}

func (p *TestProvider) LookupGroup(name string) (*Group, error) {
	for id, group := range p.Groups {
		if group == name {
			return &Group{
				Gid:  strconv.Itoa(id),
				Name: group,
			}, nil
		}
	}

	return nil, UnknownGroupError(name)
}

func (p *TestProvider) LookupGroupID(gid string) (*Group, error) {
	id, err := strconv.Atoi(gid)
	if err != nil {
		return nil, err
	}

	if name, ok := p.Groups[id]; ok {
		return &Group{Gid: gid, Name: name}, nil
	}

	return nil, UnknownGroupIdError(gid)
}

func (p *TestProvider) UserGroupIDs(u *User) ([]string, error) {
	id, err := strconv.Atoi(u.Uid)
	if err != nil {
		return nil, err
	}

	if ids, ok := p.UserGIDs[id]; ok {
		return ids, nil
	}

	return nil, UnknownUserIdError(id)
}

// ResolveHomedir resolves user home directory in paths, e.g. if path contains
// `~` symbol in prefix (like '~/.local/yourapp/config.yaml'), this function
// finds user directory and replaces to homedir.
//
// NOTE: this function doesn't make path absolute, so you need to do it
// manually.
func ResolveHomedir(users Provider, path string) (string, error) {
	// paths like '~abcd' are invalid, '~/abcd' or '~' are only valid.
	if path != "~" && !strings.HasPrefix(path, "~/") {
		return path, nil
	}

	u, err := users.Current()
	if err != nil {
		return "", err
	}

	return u.HomeDir + path[len("~"):], nil
}

// ErrGroupNotExist shows that group is not exist in the system or in the
// provider. It extends stdlib [os/user.UnknownGroupError]
type ErrGroupNotExist string

func (e ErrGroupNotExist) Error() string { return e.Unwrap().Error() }
func (e ErrGroupNotExist) Unwrap() error { return UnknownGroupError(e) }

type ErrUserNotInGroup struct {
	User          string
	ExpectedGroup string
}

func (e ErrUserNotInGroup) Error() string {
	return fmt.Sprintf("user %q expected to be in %q group", e.User, e.ExpectedGroup)
}
