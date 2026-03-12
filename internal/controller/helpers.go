package controller

import "time"

// convertFlexDateTime converts FlexDateTime to *time.Time.
// Returns nil if the time is zero.
func convertFlexDateTime(fd *FlexDateTime) *time.Time {
	if fd == nil {
		return nil
	}

	t := time.Time(*fd)
	if t.IsZero() {
		return nil
	}

	return &t
}

// convertFlexInt converts FlexInt to int.
func convertFlexInt(fi FlexInt) int {
	return int(fi)
}

// convertFlexIntToPtr converts FlexInt to *int.
// Returns nil if the value is zero.
func convertFlexIntToPtr(fi FlexInt) *int {
	if int(fi) == 0 {
		return nil
	}

	i := int(fi)

	return &i
}

// convertFlexFloat converts FlexFloat to float64.
func convertFlexFloat(ff FlexFloat) float64 {
	return float64(ff)
}

// convertFlexFloatToPtr converts FlexFloat to *float64.
// Returns nil if the value is zero.
func convertFlexFloatToPtr(ff FlexFloat) *float64 {
	if float64(ff) == 0 {
		return nil
	}

	f := float64(ff)

	return &f
}

// convertFlexBoolToPtr converts FlexBool to *bool.
// Always returns a pointer (false is a valid value).
func convertFlexBoolToPtr(fb FlexBool) *bool {
	b := bool(fb)
	return &b
}
