package nlp

import (
	"strings"
	"testing"
	"fmt"
	"bytes"
)


const tagging_data = `The/DT boy/N walked/V to/IN the/DT store/N ./.
`

const tagging_data2 = `The/DT boy/N walks/V to/IN the/DT store/N ./.
`

const dep_data = `1	Spot	_	V	VB	_	2	AMOD	_	_
2	on	_	R	RP	_	0	ROOT	_	_
3	.	_	.	.	_	2	P	_	_

`

const dep_data_wrong = `1	Spot	_	V	VBZ	_	2	AMOD	_	_
2	on	_	R	RP	_	0	ROOT	_	_
3	.	_	.	.	_	2	P	_	_

`


func Test_Read(t *testing.T) {
	corpus, err := TagFormat{}.ReadCorpus(strings.NewReader(tagging_data))
	if err != nil {
		t.Errorf("Couldn't parse: %s", err)
	}
	err = CheckSameCorpus(corpus, corpus)
	if err != nil {
		t.Errorf("Same check failed: %s", err)
	}
	corpus2, _ := TagFormat{}.ReadCorpus(strings.NewReader(tagging_data2))
	err = CheckSameCorpus(corpus, corpus2)
	if err == nil {
		t.Errorf("Same check failed.", err)
	}

	change_corpus, _ := TagFormat{}.ReadCorpus(strings.NewReader(tagging_data))
	change_corpus.sentences[0][2].tag_id = change_corpus.lexicon.GetTagId("N")
	err = CheckSameCorpus(corpus, change_corpus)
	if err != nil {
		t.Errorf("Same check failed.")
	}
	results := ScoreTagging(corpus, change_corpus)
	if results.SentencesResult.Total != 1 {
		t.Errorf("Sentence check fail.")
	}
	if results.SentencesResult.NumIncorrect() != 1 {
		t.Errorf("Sentence correct fail.")
	}
}

func Test_CoNLLRead(t *testing.T) {
	fmt.Printf("read")
	corpus, err := CoNLLFormat{}.ReadCorpus(strings.NewReader(dep_data))
	if err != nil {
		t.Errorf("Couldn't parse: %s", err)
	}
	err = CheckSameCorpus(corpus, corpus)
	if err != nil {
		t.Errorf("Same check failed: %s", err)
	}
	if n := corpus.NumSentences(); n != 1 {
		t.Errorf("Corpus length failed: %d", n)
	}

	change_corpus, err := CoNLLFormat{}.ReadCorpus(strings.NewReader(dep_data_wrong))
	if err != nil {
		t.Errorf("Couldn't parse: %s", err)
	}
	results := ScoreTagging(corpus, change_corpus)
	if results.SentencesResult.Total != 1 {
		t.Errorf("Sentence check fail.")
	}
	if results.SentencesResult.NumIncorrect() != 1 {
		t.Errorf("Sentence correct fail.")
	}
	if inc := results.TagsResult.NumIncorrect(); inc != 1 {
		t.Errorf("Tags correct fail %d.", inc)
	}

	var out bytes.Buffer
	TagFormat{}.FormatCorpus(corpus, &out)

	var conll_out bytes.Buffer
	CoNLLFormat{}.FormatCorpus(corpus, &conll_out)
	if dep_data != conll_out.String() {
		t.Errorf("CoNLL match incorrect \n %s \n %s.", dep_data, conll_out.String())
	}

}
