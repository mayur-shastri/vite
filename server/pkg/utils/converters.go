package utils

import (
	"fmt"
	"strconv"
)

func StringToUint(s string) (uint, error) {
	// Atoi parses a signed integer, so we first parse it as int...
	i, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("could not parse string to int: %w", err)
	}

	// ...then we check if it's negative. GORM IDs are unsigned.
	if i < 0 {
		return 0, fmt.Errorf("id cannot be negative")
	}

	return uint(i), nil
}
