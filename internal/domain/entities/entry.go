package entities

import "time"

type EntryType int

const (
	EntryTypeFile EntryType = iota + 1
	EntryTypeLink
	EntryTypeDir
)

type SortType int

const (
	SortTypeName SortType = iota + 1
	SortTypeSize
	SortTypeDate
)

type Entry struct {
	Type                 EntryType
	Permissions          string
	Name                 string
	OwnerUser            string
	OwnerGroup           string
	SizeInBytes          uint64
	NumHardLinks         int
	LastModificationDate time.Time
}
