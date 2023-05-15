package action

import "testing"

func TestAlgo(t *testing.T) {
	count := logarithmic(8)
	t.Log(count)
}
func logarithmic(n float64) int {
	count := 0
	for n >= 1 {
		n = n / 2
		count++
	}
	return count
}
