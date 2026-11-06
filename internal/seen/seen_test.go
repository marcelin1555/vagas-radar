package seen

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/marcelin1555/vagas-radar/internal/job"
)

func TestUnseenAndMark(t *testing.T) {
	s, err := Open(filepath.Join(t.TempDir(), "s.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	ctx := context.Background()

	a := job.Job{ID: "1", Source: "X"}
	b := job.Job{ID: "2", Source: "X"}

	novas, err := s.Unseen(ctx, []job.Job{a, b, a}) // a repetida na lista
	if err != nil || len(novas) != 2 {
		t.Fatalf("esperava 2 novas, veio %d (%v)", len(novas), err)
	}

	if err := s.Mark(ctx, []job.Job{a}); err != nil {
		t.Fatal(err)
	}
	novas, _ = s.Unseen(ctx, []job.Job{a, b})
	if len(novas) != 1 || novas[0].ID != "2" {
		t.Fatalf("só a vaga 2 deveria sobrar: %+v", novas)
	}
}

func TestSameIDDifferentSourceIsDifferentJob(t *testing.T) {
	s, _ := Open(filepath.Join(t.TempDir(), "s.db"))
	defer s.Close()
	ctx := context.Background()

	s.Mark(ctx, []job.Job{{ID: "1", Source: "A"}})
	novas, _ := s.Unseen(ctx, []job.Job{{ID: "1", Source: "B"}})
	if len(novas) != 1 {
		t.Fatal("o mesmo id em outra fonte é outra vaga")
	}
}

func TestMarkSurvivesReopen(t *testing.T) {
	path := filepath.Join(t.TempDir(), "s.db")
	s, _ := Open(path)
	s.Mark(context.Background(), []job.Job{{ID: "9", Source: "A"}})
	s.Close()

	s2, _ := Open(path)
	defer s2.Close()
	novas, _ := s2.Unseen(context.Background(), []job.Job{{ID: "9", Source: "A"}})
	if len(novas) != 0 {
		t.Fatal("a marca deveria sobreviver a reiniciar o programa")
	}
}
