// Package job define a vaga e o filtro que decide se ela interessa.
package job

import (
	"regexp"
	"strings"
	"time"
)

// Job é uma vaga vinda de qualquer fonte.
type Job struct {
	ID       string
	Source   string
	Title    string
	Company  string
	URL      string
	Location string
	Tags     []string
	PostedAt time.Time
}

// Key identifica a vaga de forma única entre fontes.
func (j Job) Key() string { return j.Source + ":" + j.ID }

// Filter escolhe as vagas pelo título e pelas tags.
type Filter struct {
	include []*regexp.Regexp
	exclude []*regexp.Regexp
}

// NewFilter cria o filtro. Cada termo casa como palavra inteira, sem
// diferenciar maiúsculas: "go" casa com "Go Developer", mas não com "Django".
// Lista de include vazia aceita tudo.
func NewFilter(include, exclude []string) *Filter {
	return &Filter{include: compile(include), exclude: compile(exclude)}
}

func compile(terms []string) []*regexp.Regexp {
	var out []*regexp.Regexp
	for _, t := range terms {
		t = strings.TrimSpace(t)
		if t == "" {
			continue
		}
		out = append(out, regexp.MustCompile(`(?i)(^|[^\p{L}\p{N}])`+regexp.QuoteMeta(t)+`($|[^\p{L}\p{N}])`))
	}
	return out
}

// Match diz se a vaga interessa.
func (f *Filter) Match(j Job) bool {
	text := j.Title + " " + strings.Join(j.Tags, " ")
	for _, re := range f.exclude {
		// Exclusão olha só o título: uma tag "senior" solta não deve
		// derrubar uma vaga júnior.
		if re.MatchString(j.Title) {
			return false
		}
	}
	if len(f.include) == 0 {
		return true
	}
	for _, re := range f.include {
		if re.MatchString(text) {
			return true
		}
	}
	return false
}

// Split quebra uma lista separada por vírgula em termos.
func Split(s string) []string {
	var out []string
	for _, p := range strings.Split(s, ",") {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}
