package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/marcelin1555/vagas-radar/internal/job"
)

func sample(n int) []job.Job {
	var out []job.Job
	for i := 0; i < n; i++ {
		out = append(out, job.Job{
			ID: "x", Source: "Teste", Title: strings.Repeat("T", 100),
			Company: "Acme", URL: "https://exemplo.com/vaga",
		})
	}
	return out
}

func TestChunksRespectLimit(t *testing.T) {
	chunks := Chunks(sample(40), 1900)
	if len(chunks) < 2 {
		t.Fatalf("40 vagas deveriam virar mais de um bloco, vieram %d", len(chunks))
	}
	for _, c := range chunks {
		if len(c) > 1900 {
			t.Fatalf("bloco com %d caracteres passa do limite", len(c))
		}
	}
}

func TestConsoleWritesEachJob(t *testing.T) {
	var buf bytes.Buffer
	if err := (Console{Out: &buf}).Send(context.Background(), sample(2)); err != nil {
		t.Fatal(err)
	}
	if strings.Count(buf.String(), "https://exemplo.com/vaga") != 2 {
		t.Fatalf("saída inesperada: %q", buf.String())
	}
}

func TestDiscordPostsContent(t *testing.T) {
	var got map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&got)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	d := Discord{URL: srv.URL, Client: srv.Client()}
	if err := d.Send(context.Background(), sample(1)); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got["content"], "Acme") {
		t.Fatalf("conteúdo não chegou: %v", got)
	}
}

func TestDiscordReportsBadStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer srv.Close()

	d := Discord{URL: srv.URL, Client: srv.Client()}
	if err := d.Send(context.Background(), sample(1)); err == nil {
		t.Fatal("deveria dar erro com status 400")
	}
}
