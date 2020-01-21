package poi

import "unicode/utf8"

type calculator struct {
	indel, sub int
}

// https://en.wikibooks.org/wiki/Algorithm_Implementation/Strings/Levenshtein_distance#C
func (c *calculator) Dist(s1, s2 string) int {
	l := utf8.RuneCountInString(s1)
	m := make([]int, l+1)
	for i := 1; i <= l; i++ {
		m[i] = i * c.indel
	}
	lastdiag, x, y := 0, 1, 1
	for _, rx := range s2 {
		m[0], lastdiag, y = x*c.indel, (x-1)*c.indel, 1
		for _, ry := range s1 {
			m[y], lastdiag = min3(m[y]+c.indel, m[y-1]+c.indel, lastdiag+c.subCost(rx, ry)), m[y]
			y++
		}
		x++
	}
	return m[l]
}

func (c *calculator) subCost(r1, r2 rune) int {
	if r1 == r2 {
		return 0
	}
	return c.sub
}

func min3(a, b, c int) int {
	return min(a, min(b, c))
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

var defaultCalculator = &calculator{1, 1}

// Dist is a convenience function for a levenshtein distance calculator with equal costs.
func LevenshteinDistance(s1, s2 string) int {
	return defaultCalculator.Dist(s1, s2)
}

func LevenshteinRatio(s1, s2 string) float64 {
	max := float64(max(len(s1), len(s2)))
	return float64(LevenshteinDistance(s1, s2)) / max
}
