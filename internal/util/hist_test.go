package util

import (
	"testing"
)

// TestGenerateDurationHistogram tests the GenerateHistogram function.
func TestGenerateDurationHistogram(t *testing.T) {
	headers := []string{"TxId", "Type", "Reason", "Duration"}
	rowValues := RowValues{
		Rows: []RowValue{
			{Row: []string{"26", "sess", "HTTP/1"}, Value: DurationValue(0)},
			{Row: []string{"27", "req", "rxreq"}, Value: DurationValue(818000)},
			{Row: []string{"28", "bereq", "fetch"}, Value: DurationValue(317000)},
			{Row: []string{"29_1", "req", "esi"}, Value: DurationValue(361000)},
			{Row: []string{"30", "bereq", "pass"}, Value: DurationValue(131000)},
			{Row: []string{"31_2", "req", "esi"}, Value: DurationValue(165000)},
			{Row: []string{"32*", "bereq", "fetch"}, Value: DurationValue(321000)},
		},
		Total: DurationValue(2113000),
	}

	expectedOutput := `
TxId | Type  | Reason | Duration | Perc | Histogram
------------------------------------------------------------------------------------------------------
26   | sess  | HTTP/1 | 0s       |   0% |
27   | req   | rxreq  | 818µs    |  38% | ############################################################
28   | bereq | fetch  | 317µs    |  15% | #######################
29_1 | req   | esi    | 361µs    |  17% | ##########################
30   | bereq | pass   | 131µs    |   6% | #########
31_2 | req   | esi    | 165µs    |   7% | ############
32*  | bereq | fetch  | 321µs    |  15% | #######################
------------------------------------------------------------------------------------------------------
Total                   2.113ms
`

	histogram := GenerateHistogram(headers, rowValues)
	if histogram != expectedOutput {
		t.Errorf("Expected output:\n%s\n\nGot:\n%s", expectedOutput, histogram)
	}
}

// TestGenerateSizesHistogram tests the GenerateHistogram function.
func TestGenerateSizesHistogram(t *testing.T) {
	headers := []string{"Type", "Direction", "Size"}
	rowValues := RowValues{
		Rows: []RowValue{
			{Row: []string{"Body", "Transmitted"}, Value: SizeValue(923)},
			{Row: []string{"Header", "Transmitted"}, Value: SizeValue(8180)},
		},
		Total: SizeValue(9103),
	}

	expectedOutput := `
Type   | Direction   | Size    | Perc | Histogram
----------------------------------------------------------------------------------------------------
Body   | Transmitted | 923B    |  10% | ######
Header | Transmitted | 7.988KB |  89% | ############################################################
----------------------------------------------------------------------------------------------------
Total                  8.890KB
`

	histogram := GenerateHistogram(headers, rowValues)
	if histogram != expectedOutput {
		t.Errorf("Expected output:\n%s\n\nGot:\n%s", expectedOutput, histogram)
	}
}
