package nlp

import (
	"strings"
	"fmt"
	"io"
	"bufio"
)

type Token struct {
	index      int
	word       string
	word_id       int
	tag        string
	tag_id     int
	label      string
	label_id      int
	category   string
	head_index int
}

type Sentence []Token

type dynamicStringMap struct {
	forward_map         []string
	reverse_map         map[string]int
	counts              []int
	counter             int	
}

func newDynamicStringMap() dynamicStringMap {
	return dynamicStringMap {
		forward_map:         make([]string, 0),
		reverse_map: map[string]int{},
		counts: make([]int, 0),
		counter:     0,
	}
}

func (dsm dynamicStringMap) TypesCount() int {
	return dsm.counter
}

func (dsm dynamicStringMap) TypeIdCount(type_id int) int {
	return dsm.counts[type_id]
}

func (dsm dynamicStringMap) TypeId(typ string) int {
	return dsm.reverse_map[typ]
}

func (dsm dynamicStringMap) Type(id int) string {
	return dsm.forward_map[id]
}

func (dsm *dynamicStringMap) UpdateTypeMap(typ string) int {
	if _, ok := dsm.reverse_map[typ]; !ok {
		dsm.reverse_map[typ] = dsm.counter
		dsm.forward_map = append(dsm.forward_map, typ)
		dsm.counts = append(dsm.counts, 0)
		dsm.counter++
	}
	id := dsm.reverse_map[typ]
	dsm.counts[id]++
	return id
}


type Lexicon struct {
	tags dynamicStringMap
	words dynamicStringMap
	labels dynamicStringMap
}

func NewLexicon() *Lexicon {
	return &Lexicon{
		tags:  newDynamicStringMap(),
		words: newDynamicStringMap(),
		labels: newDynamicStringMap(),
	}
}

func (lexicon Lexicon) TagCount() int {
	return lexicon.tags.TypesCount()
}

func (lexicon Lexicon) GetTagId(tag string) int {
	return lexicon.tags.TypeId(tag)
}

func (lexicon Lexicon) GetTag(tag_id int) string {
	return lexicon.tags.Type(tag_id)
}

func (corpus Corpus) NumSentences() int {
	return len(corpus.sentences)
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
	return lexicon.tags.Type(token.tag_id)
}

type HammingResult struct {
	Name    string
	Correct int
	Total   int
}

func (result HammingResult) Percent() float64 {
	return float64(result.Correct) / float64(result.Total)
}

type TaggingResults struct {
	HammingResult
	SentencesResult HammingResult
	TagsResult      HammingResult
	TagResults      []HammingResult
}

func (result HammingResult) NumIncorrect() int {
	return result.Total - result.Correct
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
	results.SentencesResult.Total = len(gold.sentences)
	results.TagResults = make([]HammingResult, gold.lexicon.TagCount())
	for i, test_sentence := range test.sentences {
		gold_sentence := gold.sentences[i]
		sentence_correct := true
		for j, test_token := range test_sentence {
			gold_token := gold_sentence[j]

			results.TagsResult.Total++
			results.TagResults[gold_token.tag_id].Total++
			results.TagResults[gold_token.tag_id].Name = gold.lexicon.GetTag(gold_token.tag_id)
			if gold.lexicon.GetTag(gold_token.tag_id) == test.lexicon.GetTag(test_token.tag_id) {
				results.TagsResult.Correct++
				results.TagResults[gold_token.tag_id].Correct++
			} else {
				sentence_correct = false
			}
		}
		if sentence_correct {
			results.SentencesResult.Correct++
		}
	}
	return
}

type CorpusFormatter interface {
	// Read in a corpus in this format. 
	ReadCorpus(reader io.Reader) (corpus Corpus, err error)

	// Write out a corpus in this format.
	FormatCorpus(corpus Corpus, writer io.Writer)
}

type TagFormat struct {}

func (tag_format TagFormat) ReadSentence(sent string, lexicon *Lexicon) (sentence Sentence, err error) {
	var word_tag string
	reader := strings.NewReader(sent)
	for {
		_, er := fmt.Fscanf(reader, "%s", &word_tag)
		if er != nil {
			break
		}
		split_token := strings.Split(word_tag, "/")
		id := lexicon.tags.UpdateTypeMap(split_token[1])
		word_id := lexicon.words.UpdateTypeMap(split_token[0])
		sentence = append(sentence, Token{word: split_token[0], tag_id: id, tag: split_token[1], word_id : word_id})
	}
	return
}

func (tag_format TagFormat) ReadCorpus(reader io.Reader) (corpus Corpus, err error) {
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
		sentence, err := tag_format.ReadSentence(sent, lexicon)
		if err != nil {
			return corpus, err
		}
		corpus.sentences = append(corpus.sentences, sentence)
	}
	corpus.lexicon = lexicon
	return
}

func (token Token) ToTagString() string {
	return fmt.Sprintf("%s/%s",
		token.word,
		token.tag)
}

func (sentence Sentence) ToTagString() string {
	words := make([]string, 0)
	for _, token := range sentence {
		words = append(words, token.ToTagString())
	}
	return strings.Join(words, " ")
}

func (format TagFormat) FormatCorpus(corpus Corpus, writer io.Writer) {
	for _, sent := range corpus.sentences {
		fmt.Fprintf(writer, "%s\n", sent.ToTagString())
	}
}

type CoNLLFormat struct {} 

func (token Token) ToCoNLLString() string {
	return fmt.Sprintf("%d %s _ %s %s _ %d %s _ _",
		token.index,
		token.word,
		token.category,
		token.tag,
		token.head_index,
		token.label)
}

func (sentence Sentence) ToCoNLLString() string {
	words := make([]string, 0)
	for _, token := range sentence {
		words = append(words, token.ToCoNLLString())
	}
	return strings.Join(words, "\n")
}

func (format CoNLLFormat) ReadToken(line string, lexicon *Lexicon) (token Token, err error) {
	var temp string
	fmt.Sscanf(line, "%d %s %s %s %s %s %d %s %s %s",
		&token.index,
		&token.word,
		&temp,
		&token.category,
		&token.tag,
		&temp,
		&token.head_index,
		&token.label,
		&temp,
		&temp)
	token.tag_id = lexicon.tags.UpdateTypeMap(token.tag)
	token.word_id = lexicon.words.UpdateTypeMap(token.word)
	token.label_id = lexicon.labels.UpdateTypeMap(token.label)
	return
}

func (format CoNLLFormat) FormatCorpus(corpus Corpus, writer io.Writer) {
	for _, sent := range corpus.sentences{
		fmt.Fprintf(writer, "%s\n\n", sent.ToCoNLLString())
	}
}

func (format CoNLLFormat) ReadCorpus(reader io.Reader) (corpus Corpus, err error) {
	buf_reader := bufio.NewReader(reader)
	lexicon := NewLexicon()
	var sentence Sentence
	for {
		line, err := buf_reader.ReadString('\n')
		if line == "\n" {
			corpus.sentences = append(corpus.sentences, sentence)
			sentence = make(Sentence, 0)
			continue
		}
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return corpus, err
			}
		}
		token, err := format.ReadToken(line, lexicon)
		sentence = append(sentence, token)
		if err != nil {
			return corpus, err
		}
	}
	corpus.lexicon = lexicon
	return
}

func ReadCorpus(reader io.Reader, file_name string) (corpus Corpus, err error) {
	formatter := FormatterFromFile(file_name)
	return formatter.ReadCorpus(reader)
}

var formatter = map[string]CorpusFormatter {
	"conll" : CoNLLFormat{},
	"tag" : TagFormat{},
}

func Formatter(formatter_string string) CorpusFormatter {
	return formatter[formatter_string]
}

func FormatterFromFile(file_name string) CorpusFormatter {
	split := strings.Split(file_name, ".")
	return Formatter(split[len(split) - 1])
}