package util

import (
	"fmt"
	"time"
)

const (
	Byte SizeValue = 1
	KB             = Byte * 1024
	MB             = KB * 1024
	GB             = MB * 1024
	TB             = GB * 1024
	PB             = TB * 1024
)

// ValueProvider is an interface that provides a numerical value and string representation for different types.
type ValueProvider interface {
	Value() int64
	String() string
}

// DurationValue is a custom type based on time.Duration to implement ValueProvider.
type DurationValue time.Duration

// Value returns the duration in nanoseconds.
func (d DurationValue) Value() int64 {
	return int64(d)
}

// String returns the string representation of the duration.
func (d DurationValue) String() string {
	return time.Duration(d).String()
}

// SizeValue is a custom type based on int64 to handle sizes.
type SizeValue int64

// Value returns the size in bytes.
func (s SizeValue) Value() int64 {
	return int64(s)
}

// String returns the string representation of the size.
func (s SizeValue) String() string {
	switch {
	case s >= PB:
		return fmt.Sprintf("%.3fPB", float64(s)/float64(PB))
	case s >= TB:
		return fmt.Sprintf("%.3fTB", float64(s)/float64(TB))
	case s >= GB:
		return fmt.Sprintf("%.3fGB", float64(s)/float64(GB))
	case s >= MB:
		return fmt.Sprintf("%.3fMB", float64(s)/float64(MB))
	case s >= KB:
		return fmt.Sprintf("%.3fKB", float64(s)/float64(KB))
	default:
		return fmt.Sprintf("%dB", s)
	}
}

// RowValue holds a row of data and its corresponding value.
type RowValue struct {
	Row   []string
	Value ValueProvider
}

// RowValues holds multiple RowValue and a total value.
type RowValues struct {
	Rows  []RowValue
	Total ValueProvider
}
