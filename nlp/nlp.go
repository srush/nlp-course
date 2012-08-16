package nlp

import (
	"io"
	"fmt"
	"net/http"
	ttemplate "text/template"
	htemplate "html/template"
	"encoding/json"
	"appengine"
	"errors"
)

type Page struct {
	Title  string
	Posted string
}

type Results struct {
	GoldName string         `json:"gold_name"`
	TestName string         `json:"test_name"`
	Results  TaggingResults `json:"results"`
}

type Conversions struct {
	GoldName string         `json:"gold_name"`
	TestName string         `json:"test_name"`
	Results  TaggingResults `json:"results"`
}

func init() {
	http.HandleFunc("/", handler)
	http.HandleFunc("/get", get_handler)
	http.HandleFunc("/upload", upload_handler)
	http.HandleFunc("/convert", convert_handler)
}

func handler(w http.ResponseWriter, r *http.Request) {
	p := &Page{Title: "hello"}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	startTemplate.Execute(w, p)
}

func get_handler(w http.ResponseWriter, r *http.Request) {
	t, _ := htemplate.New("posted").Parse(posted)
	name := r.FormValue("name")
	p := &Page{Title: "hello", Posted: name}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	t.Execute(w, p)
}

func http_error(w http.ResponseWriter, command string, err error) {
	http.Error(w, fmt.Sprintf("%s: %s", command, err.Error()),
		http.StatusInternalServerError)
}

func convert_handler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	var corpus Corpus
	var output_formatter CorpusFormatter
	reader, err := r.MultipartReader()
	if err != nil {
		http_error(w, "file", err)
		return
	}
	for {
		part, err := reader.NextPart()
		if err != nil {
			if err == io.EOF {
				break
			} else {
				http_error(w, "part", err)
				return
			}
		}
		name := part.FormName()
		switch name {
		case "corpus":
			file_name := part.FileName()
			corpus, err = ReadCorpus(part, file_name)
		case "outputformat":
			var format string
			fmt.Fscanf(part, "%s", &format)
			c.Infof("Formatter: %s", format)
			output_formatter = Formatter(format)
		}
		if err != nil {
			http_error(w, "Corpus parsing error", err)
			return
		}
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	c.Infof("Formatter: %v", output_formatter)
	output_formatter.FormatCorpus(corpus, w)
}

func upload_handler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	reader, err := r.MultipartReader()
	if err != nil {
		http_error(w, "file", err)
		return
	}
	var typ string
	// Read in the files.
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
				http_error(w, "part", err)
				return
			}
		}
		name := part.FormName()

		switch name {
		case "gold":
			gold_name = part.FileName()
			gold_corpus, err = ReadCorpus(part, gold_name)
		case "test":
			test_name = part.FileName()
			test_corpus, err = ReadCorpus(part, test_name)
		case "type":
			fmt.Fscanf(part, "%s", &typ)
		}
		if err != nil {
			http_error(w, "Corpus parsing error", err)
			return
		}
	}

	// Error check the corpus.
	if test_corpus.NumSentences() == 0 {
		http_error(w, "corpus check", errors.New("Test corpus blank."))
	}
	if gold_corpus.NumSentences() == 0 {
		http_error(w, "corpus check", errors.New("Gold corpus blank."))
	}
	err = CheckSameCorpus(gold_corpus, test_corpus)
	if err != nil {
		http_error(w, "corpus check", err)
		return
	}

	// Score the tagging.
	results := ScoreTagging(gold_corpus, test_corpus)
	p := &Results{
		Results:  results,
		GoldName: gold_name,
		TestName: test_name,
	}
	c.Infof("Type: %s", typ)
	switch typ {
	case "json":
		b, err := json.Marshal(p)
		if err != nil {
			http_error(w, "json err", err)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Write(b)
	default:
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		tagResultTxtTemplate.Execute(w, p)
	case "html":
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		tagResultTemplate.Execute(w, p)
	}
}

const posted = `You posted : {{.Posted}}`

const start = `
<form method="post" action="/upload" enctype="multipart/form-data">
<input type="file" name="gold"/>
<input type="file" name = "test"/>
<input type="hidden" name="html" value="true">
<input type=submit>
</form>
`

var startTemplate = htemplate.Must(htemplate.New("start").Parse(start))

const tagResultHtml = `
<html>
<title>
</title>
<body>
Tag Results

Gold File: {{.GoldName}}
Test File: {{.TestName}}

{{with .Results}} 
{{.NumCorrectTags}}
{{.NumTags}}
<table>
<tr><th>Name</th><th>Correct</th><th>Total</th><th>Percent</th></tr>
{{range .TagResults}}
<tr><td>{{.Name}}</td><td>{{.Correct}}</td><td>{{.Total}}</td><td>{{.Percent}}</td></tr>
{{end}}
</table>
{{end}}
</body>
</html>
`

var tagResultTemplate = htemplate.Must(htemplate.New("tag_result").Parse(tagResultHtml))

const tagResultTxt = `
Tag Results

Test file: {{.TestName}}
Gold file: {{.GoldName}}

{{with .Results}} 
Tags
----
{{with .TagsResult }} 
Correct:  {{printf "%5d" .Correct}}
Total:    {{printf "%5d" .Total}}
Accuracy: {{printf "%0.3f" .Percent}}
{{end}}

Sentences
---------
{{with .SentencesResult }} 
Correct:  {{printf "%5d" .Correct}}
Total:    {{printf "%5d" .Total}}
Accuracy: {{printf "%0.3f" .Percent}}
{{end}}

Tag Accuracy
Name   | Correct | Total | Percent
----------------------------------
{{range .TagResults}}
{{printf "%6s" .Name}} | {{printf "%6d" .Correct}} | {{printf "%6d" .Total}} | {{printf "%0.3f" .Percent}}
{{end}}
{{end}}
`

var tagResultTxtTemplate = ttemplate.Must(ttemplate.New("tag_result").Parse(tagResultTxt))
