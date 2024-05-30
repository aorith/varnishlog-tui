package util

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// ParseVarnishlogArgs sanitizes the script arguments.
func ParseVarnishlogArgs(input string) string {
	var result strings.Builder
	scanner := bufio.NewScanner(strings.NewReader(input))

	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)

		// Skip comment lines or empty lines
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "#") {
			continue
		}

		result.WriteString(line + "\n")
	}

	return strings.TrimSpace(result.String())
}

// ConvertUnixTimestamp converts a Unix timestamp string (integer or fractional) to a time.Time object
func ConvertUnixTimestamp(timestampStr string) (time.Time, error) {
	// Check if the timestamp contains a decimal point
	if strings.Contains(timestampStr, ".") {
		// Split the string into seconds and fractional parts
		parts := strings.SplitN(timestampStr, ".", 2)
		if len(parts) != 2 {
			return time.Time{}, fmt.Errorf("invalid timestamp format: %s", timestampStr)
		}

		// Convert the seconds part to an int64
		seconds, err := strconv.ParseInt(parts[0], 10, 64)
		if err != nil {
			return time.Time{}, fmt.Errorf("error converting seconds part: %v", err)
		}

		// Convert the fractional part to nanoseconds
		fractionalPart := parts[1]
		nanoseconds, err := strconv.ParseInt(fractionalPart+strings.Repeat("0", 9-len(fractionalPart)), 10, 64)
		if err != nil {
			return time.Time{}, fmt.Errorf("error converting fractional part: %v", err)
		}

		return time.Unix(seconds, nanoseconds), nil
	}

	// Integer timestamp format
	timestampInt, err := strconv.ParseInt(timestampStr, 10, 64)
	if err != nil {
		return time.Time{}, fmt.Errorf("error converting timestamp: %v", err)
	}

	return time.Unix(timestampInt, 0), nil
}

// Sum calculates the sum of the elements in the given slice.
func Sum(arr []int) (total int) {
	for _, v := range arr {
		total += v
	}
	return total
}
