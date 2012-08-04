package nlp
import ( 
	"strings"


type Token struct {
	word string
	tag_id int
}

type Sentence []Token;

type Lexicon struct {
	tag_map map[int]string
	reverse_tag_map map[string]int
}

type Corpus struct {
	sentences []Sentence 
	lexicon *Lexicon
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
	num_sentences int
	num_tags int
	num_correct_tags int
	num_correct_sentences int
	num_tag []int
	num_correct_tag []int
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

func CheckSameCorpus(gold Corpus, test Corpus) (error) {
	if len(gold.sentences) != len(test.sentences) {
		return ScoringError { "Number of sentences differ." }
	}
	if len(gold.lexicon.tag_map) != len(test.lexicon.tag_map) {
		return ScoringError { "Number of tags differ." }
	}
	for i, tag := range gold.lexicon.tag_map {
		if tag != test.lexicon.tag_map[i] {
			return ScoringError { "Tag map differs." }
		}
	}
	for i, test_sent := range test.sentences {
		gold_sent := gold.sentences[i] 
		if !CheckSameSentence(gold_sent, test_sent) {
			return ScoringError { "Sentence size differs." }
		}
		for j, test_token := range test_sent {
			gold_token := gold_sent[j]
			if gold_token.word != test_token.word {
				return ScoringError { "Words differ." }
			}
		}
	}
}

func ScoreTagging(gold Corpus, test Corpus) (results TaggingResults) {
	results.num_sentences = len(gold.sentences)
	for i, test_sentence := range test.sentences {
		gold_sentence := gold.sentences[i] 
		sentence_correct_tags := 0
		for j, test_token := range test_sentence {
			gold_token := gold_sentence[j]

			results.num_tags++
			results.num_tag[gold_token.tag_id]++
			if gold_token.tag_id == test_token.tag_id {
				results.num_correct_tags++
				results.num_correct_tag[gold_token.tag_id]++
			}
		}
		if sentence_correct_tags == len(test_sentence) {
			results.num_correct_sentences++
		}
	}
}

func ParseSentence(sent string, reverse_tag_map *map[string]int, counter *int) (sentence Sentence, err error) {
	tokens := strings.Split(sent, " ")
	for i, token_str := range tokens {
		split_token := strings.Split(token_str, "/")
		if len(split_token) != 2 {
			return nil, ParseError{ "Couldn't parse token." }
		}
		if (*reverse_tag_map)[split_token[1]] {
			(*reverse_tag_map)[split_token[1]] = *counter
			(*counter)++
		}
		id := (*reverse_tag_map)[split_token[1]]
		sentence = append(sentence, Token{word: split_token[0], tag_id: id})
	}
}

func ParseCorpus(bufio) (corpus Corpus, err error) {
	tag_counter := 0
	for {
		sentence, err := ParseSentence(sent, &corpus.lexicon.reverse_tag_map, &tag_counter)
		if err != nil { 
			return nil, err
		}
		corpus.sentences = append(corpus.sentences, sentence)
	}
	for tag, tag_id := range corpus.lexicon.reverse_tag_map {
		corpus.lexicon.tag_map[tag_id] = tag
	}
}