package service

import (
	"unicode"
)

func Luhn(code string) bool {
	if len(code) == 0 {
		return false
	}

	for _, char := range code {
		if !unicode.IsDigit(char) {
			return false
		}
	}

	digits := make([]int, len(code))
	for i, char := range code {
		digits[i] = int(char - '0')
	}

	sum := 0
	isSecond := false

	for i := len(digits) - 1; i >= 0; i-- {
		digit := digits[i]

		if isSecond {
			digit *= 2
			if digit > 9 {
				digit = digit - 9
			}
		}

		sum += digit
		isSecond = !isSecond
	}

	return sum%10 == 0
}
