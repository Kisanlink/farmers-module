package testutils

// StringPtr converts a string to *string
func StringPtr(s string) *string {
	return &s
}

// IntPtr converts an int to *int
func IntPtr(i int) *int {
	return &i
}

// Float64Ptr converts a float64 to *float64
func Float64Ptr(f float64) *float64 {
	return &f
}
