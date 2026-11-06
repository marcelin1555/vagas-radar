package source

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/marcelin1555/vagas-radar/internal/job"
)

func TestRemotiveParsesJobs(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("User-Agent") == "" {
			t.Error("deveria mandar User-Agent")
		}
		w.Write([]byte(`{"jobs":[{"id":42,"url":"https://x/42","title":"Python Dev","company_name":"Acme",
			"tags":["python"],"candidate_required_location":"Worldwide","publication_date":"2026-11-04T16:53:35"}]}`))
	}))
	defer srv.Close()

	jobs, err := (&Remotive{BaseURL: srv.URL, Client: srv.Client()}).Fetch(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(jobs) != 1 || jobs[0].ID != "42" || jobs[0].Title != "Python Dev" || jobs[0].Company != "Acme" {
		t.Fatalf("parse errado: %+v", jobs)
	}
	if jobs[0].PostedAt.IsZero() {
		t.Fatal("data não foi lida")
	}
}

func TestRemoteOKSkipsLegalNotice(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`[{"last_updated":1,"legal":"aviso"},
			{"id":"7","url":"https://x/7","position":"Go Engineer","company":"Beta","tags":["golang"],
			 "location":"Anywhere","date":"2026-11-05T22:37:39+00:00"}]`))
	}))
	defer srv.Close()

	jobs, err := (&RemoteOK{BaseURL: srv.URL, Client: srv.Client()}).Fetch(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(jobs) != 1 || jobs[0].Title != "Go Engineer" || jobs[0].Source != "Remote OK" {
		t.Fatalf("parse errado: %+v", jobs)
	}
}

func TestFetchErrorOnBadStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "fora do ar", http.StatusServiceUnavailable)
	}))
	defer srv.Close()

	if _, err := (&Remotive{BaseURL: srv.URL, Client: srv.Client()}).Fetch(context.Background()); err == nil {
		t.Fatal("deveria dar erro com status 503")
	}
}

type fake struct {
	name string
	jobs []job.Job
	err  error
	wait time.Duration
}

func (f fake) Name() string { return f.name }
func (f fake) Fetch(ctx context.Context) ([]job.Job, error) {
	select {
	case <-time.After(f.wait):
		return f.jobs, f.err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func TestFetchAllIsolatesFailures(t *testing.T) {
	sources := []Source{
		fake{name: "boa", jobs: []job.Job{{ID: "1"}}},
		fake{name: "quebrada", err: errors.New("caiu")},
		fake{name: "lenta", wait: time.Second},
	}
	res := FetchAll(context.Background(), sources, 100*time.Millisecond)

	if len(res) != 3 {
		t.Fatalf("esperava 3 resultados, veio %d", len(res))
	}
	if res[0].Err != nil || len(res[0].Jobs) != 1 {
		t.Fatalf("fonte boa deveria funcionar: %+v", res[0])
	}
	if res[1].Err == nil {
		t.Fatal("fonte quebrada deveria trazer erro")
	}
	if !errors.Is(res[2].Err, context.DeadlineExceeded) {
		t.Fatalf("fonte lenta deveria estourar o tempo, veio %v", res[2].Err)
	}
}
