// Copyright 2011 The LevelDB-Go and Pebble Authors. All rights reserved. Use
// of this source code is governed by a BSD-style license that can be found in
// the LICENSE file.

package db

import (
	"fmt"
	"math/rand"
	"testing"
)

func TestDefAppendSeparator(t *testing.T) {
	testCases := []struct {
		a, b, want string
	}{
		// Examples from the doc comments.
		{"black", "blue", "blb"},
		{"green", "", "green"},
		// Non-empty b values. The C++ Level-DB code calls these separators.
		{"", "2", ""},
		{"1", "2", "1"},
		{"1", "29", "2"},
		{"13", "19", "14"},
		{"13", "99", "2"},
		{"135", "19", "14"},
		{"1357", "19", "14"},
		{"1357", "2", "14"},
		{"13\xff", "14", "13\xff"},
		{"13\xff", "19", "14"},
		{"1\xff\xff", "19", "1\xff\xff"},
		{"1\xff\xff", "2", "1\xff\xff"},
		{"1\xff\xff", "9", "2"},
		// Empty b values. The C++ Level-DB code calls these successors.
		{"", "", ""},
		{"1", "", "1"},
		{"11", "", "11"},
		{"11\xff", "", "11\xff"},
		{"1\xff", "", "1\xff"},
		{"1\xff\xff", "", "1\xff\xff"},
		{"\xff", "", "\xff"},
		{"\xff\xff", "", "\xff\xff"},
		{"\xff\xff\xff", "", "\xff\xff\xff"},
	}
	for _, tc := range testCases {
		t.Run("", func(t *testing.T) {
			got := string(DefaultComparer.Separator(nil, []byte(tc.a), []byte(tc.b)))
			if got != tc.want {
				t.Errorf("a, b = %q, %q: got %q, want %q", tc.a, tc.b, got, tc.want)
			}
		})
	}
}

func getRandKey() []byte {
	res := make([]byte, rand.Intn(30))
	if _, err := rand.Read(res); err != nil {
		panic(err)
	}
	return res
}

func TestAbbreviatedKey(t *testing.T) {
	testCases := []struct {
		k  []byte
		ak uint64
	}{
		{[]byte{}, 0x0000000000000000},
		{[]byte{0x00}, 0x0000000000000000},
		{[]byte{0xab}, 0xab00000000000000},
		{[]byte{0xab, 0xcd}, 0xabcd000000000000},
		{[]byte{0x00, 0xab, 0xcd}, 0x00abcd0000000000},
		{[]byte{0x00, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88}, 0x0011223344556677},
	}
	for _, tc := range testCases {
		t.Run(string(fmt.Sprintf("%v", tc.k)), func(t *testing.T) {
			actual := DefaultComparer.AbbreviatedKey(tc.k)
			if actual != tc.ak {
				t.Fatalf(
					"expected AbbreviatedKey(%v) = %x, got %x",
					tc.k, tc.ak, actual,
				)
			}
		})
	}
}

func TestAbbreviatedKeyRandom(t *testing.T) {
	for i := 0; i < 100; i++ {
		a, b := getRandKey(), getRandKey()
		aabbrev := DefaultComparer.AbbreviatedKey(a)
		babbrev := DefaultComparer.AbbreviatedKey(b)
		cmp := DefaultComparer.Compare(a, b)
		if aabbrev < babbrev && cmp >= 0 {
			t.Fatalf("%v >= %v but %016x < %016x", a, b, aabbrev, babbrev)
		}
		if aabbrev > babbrev && cmp <= 0 {
			t.Fatalf("%v <= %v but %016x > %016x", a, b, aabbrev, babbrev)
		}
	}
}
