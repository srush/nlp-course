package nlp

struct HmmConfiguration {
	unknown_threshold int 	
	order int
}

func EstimateFromCorpus(corpus Corpus) HMM{
	tag_counts := corpus.tags.TypeCount()
	word_counts := corpus.words.TypeCount()
}