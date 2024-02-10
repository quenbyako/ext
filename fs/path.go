// Copyright (c) 2020-2024 Richard Cooper
//
// This file is a part of quenbyako/ext package.
// See https://github.com/quenbyako/ext/blob/master/LICENSE for details

package fs

import (
	"errors"
	stdpath "path"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

func PreparePath(homedirResolver func(string) (string, error), workdir string, path string) (res string, err error) {
	if !filepath.IsAbs(workdir) {
		return "", errors.New("working directory is not absolute")
	}

	// `~/some/path` is an exception, dealing with it firstly
	if homedirResolver == nil {
		// fallback, if you didn't provide user provider
		if path == "~" || strings.HasPrefix(path, "~/") {
			return "", errors.New("path contains user's directory alias, which is forbidden")
		}
	} else if path, err = homedirResolver(path); err != nil {
		return "", err
	}

	if !filepath.IsAbs(path) {
		path = filepath.Join(workdir, path)
	}

	var ok bool
	if res, ok = NormalizePath(path); ok {
		return res, nil
	}

	return "", errors.New("invalid path")
}

// NormalizePath works like filepath.Clean, but it prepares path for fs
func NormalizePath(path string) (string, bool) {
	p := path
	// already valid?
	if ValidPath(p) {
		return p, true
	}
	// valid string at least?
	if !utf8.ValidString(p) || p == "" {
		return p, false
	}
	// have .. or . ?
	p = stdpath.Clean(p)
	// last item is path is separator?
	if p[len(p)-1] == '/' {
		p = p[:len(p)-1]
	}
	// is absolute?
	if p[0] == '/' {
		p = p[1:]
	}
	// soooooo?
	if ValidPath(p) {
		return p, true
	}

	return path, false
}

func Abs(workdir, path string) string {
	if stdpath.IsAbs(path) {
		return stdpath.Clean(path)
	}

	return stdpath.Join(workdir, path)
}

// Rel returns a relative path that is lexically equivalent to targpath when
// joined to basepath with an intervening separator. That is,
// Join(basepath, Rel(basepath, targpath)) is equivalent to targpath itself.
// On success, the returned path will always be relative to basepath,
// even if basepath and targpath share no elements.
// An error is returned if targpath can't be made relative to basepath or if
// knowing the current working directory would be necessary to compute it.
// Rel calls Clean on the result.
func Rel(basepath string, targpath string) (string, error) {
	base := stdpath.Clean(basepath)
	targ := stdpath.Clean(targpath)
	if targ == base {
		return ".", nil
	}

	panic("unimplemented")
}
