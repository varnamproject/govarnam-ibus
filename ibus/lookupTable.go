package ibus

/**
 * goibus - golang implementation of libibus
 * Copyright Sarim Khan, 2016
 * Copyright Nguyen Tran Hau, 2021
 * https://github.com/sarim/goibus
 * Licensed under Mozilla Public License 1.1 ("MPL")
 *
 * Derivative Changes: Add new functions for lookup table modification
 * Copyright Subin Siby, 2021
 */

import (
	"github.com/godbus/dbus/v5"
)

type LookupTable struct {
	Name          string
	Attachments   map[string]dbus.Variant
	PageSize      uint32
	CursorPos     uint32
	CursorVisible bool
	Round         bool
	Orientation   int32
	Candidates    []dbus.Variant
	Labels        []dbus.Variant
}

func NewLookupTable() *LookupTable {
	lt := &LookupTable{}
	lt.Name = "IBusLookupTable"
	lt.PageSize = 9
	lt.CursorPos = 0
	lt.CursorVisible = true
	lt.Round = true
	lt.Orientation = ORIENTATION_SYSTEM

	return lt
}

func (lt *LookupTable) AppendCandidate(text string) {
	t := NewText(text)
	lt.Candidates = append(lt.Candidates, dbus.MakeVariant(*t))
}

func (lt *LookupTable) AppendLabel(label string) {
	l := NewText(label)
	lt.Labels = append(lt.Labels, dbus.MakeVariant(*l))
}

func (lt *LookupTable) CursorUp() {
	if lt.CursorPos == 0 && lt.Round {
		lt.CursorPos = uint32(len(lt.Candidates)) - 1
	} else {
		lt.CursorPos--
	}
}

func (lt *LookupTable) CursorDown() {
	if lt.CursorPos == uint32(len(lt.Candidates))-1 && lt.Round {
		lt.CursorPos = uint32(0)
	} else {
		lt.CursorPos++
	}
}

func (lt *LookupTable) Clear() {
	lt.Candidates = []dbus.Variant{}
	lt.Labels = []dbus.Variant{}
	lt.CursorPos = 0
}
