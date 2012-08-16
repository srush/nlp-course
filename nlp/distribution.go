package nlp

type MultinomialCounts struct {
	counts map[int]int
}

func NewCounts() (counts MultinomialCounts) {
	counts.counts = make(map[int]int)
	return
}

func (counts *MultinomialCounts) Inc(key int) {
	if _, ok := counts.counts[key]; ok {
		counts.counts[key] = 0
	}
	counts.counts[key]++
}

type Multinomial struct {
	distribution map[int]float64
}

func (multi Multinomial) Prob(key int) float64 {
	return multi.distribution[key]
}

func (counts MultinomialCounts) MaximumLikelihood() (multinomial Multinomial) {
	multinomial.distribution = make(map[int]float64)
	sum := 0.0
	for _, count := range counts.counts {
		sum += float64(count)
	}
	for key, count := range counts.counts {
		multinomial.distribution[key] = float64(count) / sum
	}
	return
}
