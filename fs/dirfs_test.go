// Copyright (c) 2020-2024 Richard Cooper
//
// This file is a part of quenbyako/ext package.
// See https://github.com/quenbyako/ext/blob/master/LICENSE for details

package fs

import (
	"reflect"
	"testing"
)

func TestNormalizeDirFS(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		name   string
		want   string
		wantOk bool
	}{
		{"", "/", true},
		{"/", "/", true},
		{"/a/b/c", "/a/b/c", true},
		{"/a/", "/a", true},
		{"a/b/c", "", false},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got, gotOk := normalizeDirFS(tt.name)
			assertEqual(t, tt.want, got)
			assertEqual(t, tt.wantOk, gotOk)
		})
	}
}

func assertEqual[T any](t *testing.T, want, got T) {
	if !reflect.DeepEqual(want, got) {
		t.Logf("want: %#v, got %#v", want, got)
		t.FailNow()
	}
}
