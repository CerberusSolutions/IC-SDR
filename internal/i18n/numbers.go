package i18n

import (
	"fmt"
	"strings"
)

// Number formatting differs between the two languages only in how digits are
// grouped. Spanish groups with a full stop (446.193.750) and British English
// groups with a comma (446,193,750).
//
// Decimal separators are deliberately left alone. IC-SDR already prints every
// decimal value with Go's "." through %f, which is correct for British
// English, and rewriting the Spanish side to use a comma would change output
// the original author did not ask to have changed.

// GroupSeparator returns the digit grouping separator for the active language.
func GroupSeparator() string {
	if Current() == English {
		return ","
	}
	return "."
}

// FormatDialFrequency renders a frequency in hertz for the main VFO dial,
// grouped in threes for the active language.
//
// The dial is digit-addressable: clicking a digit selects the decade it tunes.
// That hit test counts only characters in the range 0-9 and steps over
// everything else, so swapping the separator changes the readout without
// affecting which digit a click selects.
func FormatDialFrequency(hz int64) string {
	sign := ""
	if hz < 0 {
		sign, hz = "-", -hz
	}
	separator := GroupSeparator()
	return fmt.Sprintf("%s%d%s%03d%s%03d", sign, hz/1_000_000, separator, (hz/1_000)%1_000, separator, hz%1_000)
}

// GroupDigits inserts the active language's grouping separator into a whole
// number every three digits.
func GroupDigits(value int64) string {
	sign := ""
	if value < 0 {
		sign, value = "-", -value
	}
	digits := fmt.Sprintf("%d", value)
	separator := GroupSeparator()
	var builder strings.Builder
	builder.WriteString(sign)
	leading := len(digits) % 3
	if leading == 0 {
		leading = 3
	}
	builder.WriteString(digits[:leading])
	for index := leading; index < len(digits); index += 3 {
		builder.WriteString(separator)
		builder.WriteString(digits[index : index+3])
	}
	return builder.String()
}
