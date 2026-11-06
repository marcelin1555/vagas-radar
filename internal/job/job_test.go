package job

import "testing"

func TestFilterMatch(t *testing.T) {
	f := NewFilter([]string{"python", "go", "ai"}, []string{"senior", "lead"})

	casos := []struct {
		nome string
		j    Job
		want bool
	}{
		{"python no título", Job{Title: "Junior Python Developer"}, true},
		{"go como palavra", Job{Title: "Go Backend Engineer"}, true},
		{"go não casa com Django", Job{Title: "Django Developer"}, false},
		{"casa pela tag", Job{Title: "Backend Developer", Tags: []string{"Golang", "ai"}}, true},
		{"excluída por senior", Job{Title: "Senior Python Developer"}, false},
		{"tag senior não exclui", Job{Title: "Python Developer", Tags: []string{"senior"}}, true},
		{"sem relação", Job{Title: "Graphic Designer"}, false},
		{"maiúsculas", Job{Title: "PYTHON ENGINEER"}, true},
	}
	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			if got := f.Match(c.j); got != c.want {
				t.Fatalf("Match(%q) = %v, esperava %v", c.j.Title, got, c.want)
			}
		})
	}
}

func TestEmptyIncludeAcceptsAll(t *testing.T) {
	f := NewFilter(nil, []string{"senior"})
	if !f.Match(Job{Title: "Qualquer coisa"}) {
		t.Fatal("sem include deveria aceitar")
	}
	if f.Match(Job{Title: "Senior qualquer"}) {
		t.Fatal("exclude deveria valer mesmo sem include")
	}
}

func TestSplit(t *testing.T) {
	got := Split(" python, go ,, lua ")
	if len(got) != 3 || got[0] != "python" || got[1] != "go" || got[2] != "lua" {
		t.Fatalf("Split errado: %v", got)
	}
}
