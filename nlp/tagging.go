package nlp

import (
	"strings"
	"fmt"
	"io"
	"bufio"
)

type Token struct {
	word   string
	tag_id int
}

type Sentence []Token

type Lexicon struct {
	tag_map         map[int]string
	reverse_tag_map map[string]int
	tag_counter     int
}

func NewLexicon() *Lexicon {
	return &Lexicon{
		tag_map:         map[int]string{},
		reverse_tag_map: map[string]int{},
		tag_counter:     0,
	}
}

func (lexicon Lexicon) TagCount() int {
	return len(lexicon.tag_map)
}

func (lexicon Lexicon) GetTagId(tag string) int {
	return lexicon.reverse_tag_map[tag]
}

func (lexicon Lexicon) GetTag(tag_id int) string {
	return lexicon.tag_map[tag_id]
}

func (lexicon *Lexicon) UpdateTagMap(tag string) int {
	if _, ok := lexicon.reverse_tag_map[tag]; !ok {
		lexicon.reverse_tag_map[tag] = lexicon.tag_counter
		lexicon.tag_counter++
	}
	return lexicon.reverse_tag_map[tag]
}

func (lexicon *Lexicon) FinishLexicon() {
	for tag, tag_id := range lexicon.reverse_tag_map {
		lexicon.tag_map[tag_id] = tag
	}
}

type Corpus struct {
	sentences []Sentence
	lexicon   *Lexicon
}

func (token Token) Word() string {
	return token.word
}

func (token Token) TagId() int {
	return token.tag_id
}

func (lexicon Lexicon) Tag(token Token) string {
	return lexicon.tag_map[token.tag_id]
}

type TaggingResults struct {
	num_sentences         int
	num_tags              int
	num_correct_tags      int
	num_correct_sentences int
	num_tag               []int
	num_correct_tag       []int
}

func (results TaggingResults) NumIncorrect() int {
	return results.num_tags - results.num_correct_tags
}

type ScoringError struct {
	error string
}

func (err ScoringError) Error() string {
	return err.error
}

type ParseError struct {
	error string
}

func (err ParseError) Error() string {
	return err.error
}

func CheckSameSentence(sent1 Sentence, sent2 Sentence) bool {
	return len(sent1) == len(sent2)
}

func CheckSameCorpus(gold Corpus, test Corpus) error {
	if len(gold.sentences) != len(test.sentences) {
		return ScoringError{"Number of sentences differ."}
	}
	// if len(gold.lexicon.tag_map) != len(test.lexicon.tag_map) {
	// 	return ScoringError{"Number of tags differ."}
	// }
	// for i, tag := range gold.lexicon.tag_map {
	// 	if tag != test.lexicon.tag_map[i] {
	// 		return ScoringError{"Tag map differs."}
	// 	}
	// }
	for i, test_sent := range test.sentences {
		gold_sent := gold.sentences[i]
		if !CheckSameSentence(gold_sent, test_sent) {
			return ScoringError{"Sentence size differs."}
		}
		for j, test_token := range test_sent {
			gold_token := gold_sent[j]
			if gold_token.word != test_token.word {
				return ScoringError{"Words differ."}
			}
		}
	}
	return nil
}

func ScoreTagging(gold Corpus, test Corpus) (results TaggingResults) {
	results.num_sentences = len(gold.sentences)
	results.num_tag = make([]int, gold.lexicon.TagCount())
	results.num_correct_tag = make([]int, gold.lexicon.TagCount())
	for i, test_sentence := range test.sentences {
		gold_sentence := gold.sentences[i]
		sentence_correct_tags := 0
		for j, test_token := range test_sentence {
			gold_token := gold_sentence[j]

			results.num_tags++
			results.num_tag[gold_token.tag_id]++
			if gold.lexicon.GetTag(gold_token.tag_id) == test.lexicon.GetTag(test_token.tag_id) {
				results.num_correct_tags++
				results.num_correct_tag[gold_token.tag_id]++
			}
		}
		if sentence_correct_tags == len(test_sentence) {
			results.num_correct_sentences++
		}
	}
	return
}

func ParseSentence(sent string, lexicon *Lexicon) (sentence Sentence, err error) {
	var word_tag string
	reader := strings.NewReader(sent)
	for {
		_, er := fmt.Fscanf(reader, "%s", &word_tag)
		if er != nil {
			break
		}
		split_token := strings.Split(word_tag, "/")
		id := lexicon.UpdateTagMap(split_token[1])
		sentence = append(sentence, Token{word: split_token[0], tag_id: id})
	}
	return
}

func ParseCorpus(reader io.Reader) (corpus Corpus, err error) {
	buf_reader := bufio.NewReader(reader)
	lexicon := NewLexicon()
	for {
		sent, err := buf_reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return corpus, err
			}
		}
		sentence, err := ParseSentence(sent, lexicon)
		if err != nil {
			return corpus, err
		}
		corpus.sentences = append(corpus.sentences, sentence)
	}
	lexicon.FinishLexicon()
	corpus.lexicon = lexicon
	return
}
