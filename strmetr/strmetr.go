// Package strmeter measures the distances between two strings.
package strmetr

// An ad-hoc function for a percentage difference between two strings.
func Approx(str1, str2 string) float64 {
	var sum1 float64
	for _, chr := range str1 {
		sum1 += float64(chr)
	}
	var sum2 float64
	for _, chr := range str2 {
		sum2 += float64(chr)
	}
	if sum1 > sum2 {
		return 100 - (float64(sum2/sum1) * 100)
	} else if sum2 > sum1 {
		return 100 - (float64(sum1/sum2) * 100)
	}
	return 0
}
