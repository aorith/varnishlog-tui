package util

import (
	"fmt"
	"strings"
)

const histogramLength = 60

// GenerateHistogram generates an ASCII histogram based on the given headers and RowValues.
//
//	headers := []string{"TxId", "Type", "Reason", "Duration"}
//	rowValues := RowValues{
//		Rows: []RowValue{
//			{Row: []string{"26", "sess", "HTTP/1"}, Value: DurationValue(0)},
//			{Row: []string{"27", "req", "rxreq"}, Value: DurationValue(818000)},
//			{Row: []string{"28", "bereq", "fetch"}, Value: DurationValue(317000)},
//			{Row: []string{"29_1", "req", "esi"}, Value: DurationValue(361000)},
//			{Row: []string{"30", "bereq", "pass"}, Value: DurationValue(131000)},
//			{Row: []string{"31_2", "req", "esi"}, Value: DurationValue(165000)},
//			{Row: []string{"32*", "bereq", "fetch"}, Value: DurationValue(321000)},
//		},
//		Total: DurationValue(2113000),
//	}
//
//	histogram := GenerateHistogram(headers, rowValues)
//	if histogram != expectedOutput {
//		t.Errorf("Expected output:\n%s\n\nGot:\n%s", expectedOutput, histogram)
//	}
func GenerateHistogram(headers []string, rowValues RowValues) string {
	var (
		maxValue     int64
		maxLens      = make([]int, len(headers))
		lenHistChars int
		perc         int
		s            strings.Builder
	)

	// Initialize max lengths for each column
	for i, header := range headers {
		maxLens[i] = len(header)
	}

	// Total column is placed on the last header
	totalStr := rowValues.Total.String()
	if len(totalStr) > maxLens[len(headers)-1] {
		maxLens[len(headers)-1] = len(totalStr)
	}

	for _, rowValue := range rowValues.Rows {
		for j, col := range rowValue.Row {
			if len(col) > maxLens[j] {
				maxLens[j] = len(col)
			}
		}
	}

	for _, rowValue := range rowValues.Rows {
		value := rowValue.Value.Value()
		if value > maxValue {
			maxValue = value
		}
		valueStr := rowValue.Value.String()
		if len(valueStr) > maxLens[len(headers)-1] { // Values are on the last header
			maxLens[len(headers)-1] = len(valueStr)
		}
	}

	// Separator for the "table"
	separator := strings.Repeat("-", Sum(maxLens)+((len(headers)+2)*3)+int(histogramLength)+1) + "\n"

	// Headers
	s.WriteRune('\n')
	for i, header := range headers {
		s.WriteString(fmt.Sprintf("%-*s | ", maxLens[i], header))
	}
	s.WriteString("Perc | Histogram\n")
	s.WriteString(separator)

	// Rows
	for _, rowValue := range rowValues.Rows {
		value := rowValue.Value.Value()
		row := rowValue.Row
		if rowValues.Total.Value() > 0 {
			perc = int(value * 100 / rowValues.Total.Value())
		} else {
			perc = 0
		}
		if maxValue > 0 {
			lenHistChars = int(value * int64(histogramLength) / maxValue)
		} else {
			lenHistChars = 0
		}
		for j, col := range row {
			s.WriteString(fmt.Sprintf("%-*s | ", maxLens[j], col))
		}
		s.WriteString(fmt.Sprintf("%-*s | ", maxLens[len(headers)-1], rowValue.Value.String()))
		s.WriteString(fmt.Sprintf("%3d%% |", perc))
		if lenHistChars > 0 {
			s.WriteString(fmt.Sprintf(" %s\n", strings.Repeat("#", lenHistChars)))
		} else {
			s.WriteRune('\n')
		}
	}

	// Total
	s.WriteString(separator)
	totalHeader := fmt.Sprintf("%-*s", Sum(maxLens)-maxLens[len(maxLens)-1]-3+(len(headers)*3), "Total")
	s.WriteString(fmt.Sprintf("%s%s\n", totalHeader, totalStr))

	return s.String()
}
