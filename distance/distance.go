// Package distance measures the distances between two strings.
package distance

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
		return 1 - sum2/sum1
	} else if sum2 > sum1 {
		return 1 - sum1/sum2
	}
	return 0
}
