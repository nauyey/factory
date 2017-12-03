package utils

import "strings"

// StringSliceContains check whether a string  is in a slice
func StringSliceContains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// StringSliceTrim go through the string slice, trim each intem string by cutset,
// and return a new string slice.
func StringSliceTrim(slice []string, cutset string) []string {
	sliceTrimed := make([]string, len(slice))
	for i, s := range slice {
		sliceTrimed[i] = strings.Trim(s, cutset)
	}
	return sliceTrimed
}

// StringSliceToLower go through the string slice, make earch item string lower case,
// and return a new string slice.
func StringSliceToLower(slice []string) []string {
	sliceLowercase := make([]string, len(slice))
	for i, s := range slice {
		sliceLowercase[i] = strings.ToLower(s)
	}
	return sliceLowercase
}
