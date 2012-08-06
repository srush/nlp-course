package nlp

import (
	"io"
	"fmt"
	"net/http"
	"html/template"
)

type Page struct {
	Title  string
	Posted string
}

type Results struct {
	Correct int
	Total int
	GoldName string
	TestName string
}

func init() {
	http.HandleFunc("/", handler)
	http.HandleFunc("/get", get_handler)
	http.HandleFunc("/upload", upload_handler)
}

func handler(w http.ResponseWriter, r *http.Request) {
	t, _ := template.New("start").Parse(start)
	p := &Page{Title: "hello"}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	t.Execute(w, p)
}

func get_handler(w http.ResponseWriter, r *http.Request) {
	t, _ := template.New("posted").Parse(posted)
	name := r.FormValue("name")
	p := &Page{Title: "hello", Posted: name}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	t.Execute(w, p)
}

func upload_handler(w http.ResponseWriter, r *http.Request) {
	reader, err := r.MultipartReader()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var gold_corpus Corpus
	var test_corpus Corpus
	var gold_name string
	var test_name string
	for {
		part, err := reader.NextPart()
		if err != nil {
			if err == io.EOF {
				break
			} else {
				http.Error(w, fmt.Sprintf("part: %s", err.Error()), http.StatusInternalServerError)
				return
			}
		}
		name := part.FormName()
		if name == "gold" {
			gold_corpus, err = ParseCorpus(part)
			gold_name = part.FileName()
		} else {
			test_corpus, err = ParseCorpus(part)
			test_name = part.FileName()
		}
	}
	err = CheckSameCorpus(gold_corpus, test_corpus)
	if err != nil {
		http.Error(w, fmt.Sprintf("Corpus check: %s", err.Error()), http.StatusInternalServerError)
		return
	}
	results := ScoreTagging(gold_corpus, test_corpus)
	p := &Results{
		Correct: results.num_correct_tags, 
		Total: results.num_tags, 
		GoldName : gold_name,
		TestName : test_name,
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	scoringTemplate.Execute(w, p)
}

const start = `<form method="post" action="/upload" enctype="multipart/form-data" ><input type="file" name="gold"/><input type="file" name = "test"/><input type=submit></form>`
const posted = `You posted : {{.Posted}}`

const scoring = `{{.GoldName}} {{.TestName}} Your tagger got {{.Correct}} out of {{.Total}} correct.`
var scoringTemplate = template.Must(template.New("scoring").Parse(scoring))