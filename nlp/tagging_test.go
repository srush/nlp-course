package nlp

import (
	"strings"
	"testing"
)

const tagging_data = `The/DT boy/N walked/V to/IN the/DT store/N ./.
`

const tagging_data2 = `The/DT boy/N walks/V to/IN the/DT store/N ./.
`

func Test_Read(t *testing.T) {
	corpus, err := ParseCorpus(strings.NewReader(tagging_data))
	if err != nil {
		t.Errorf("Couldn't parse: %s", err)
	}
	err = CheckSameCorpus(corpus, corpus)
	if err != nil {
		t.Errorf("Same check failed: %s", err)
	}
	corpus2, _ := ParseCorpus(strings.NewReader(tagging_data2))
	err = CheckSameCorpus(corpus, corpus2)
	if err == nil {
		t.Errorf("Same check failed.", err)
	}

	change_corpus, _ := ParseCorpus(strings.NewReader(tagging_data))
	change_corpus.sentences[0][2].tag_id = change_corpus.lexicon.GetTagId("N")
	err = CheckSameCorpus(corpus, change_corpus)
	if err != nil {
		t.Errorf("Same check failed.")
	}
	results := ScoreTagging(corpus, change_corpus)
	if results.num_sentences != 1 {
		t.Errorf("Sentence check fail.")
	}
	if results.NumIncorrect() != 1 {
		t.Errorf("Sentence correct fail.")
	}
}
