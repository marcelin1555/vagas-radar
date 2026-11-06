// Package source busca vagas em APIs públicas de vagas remotas.
package source

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/marcelin1555/vagas-radar/internal/job"
)

const (
	userAgent = "vagas-radar/0.1 (projeto pessoal de estudo)"
	maxBody   = 10 << 20 // 10 MiB
)

// Source é qualquer lugar de onde dá para buscar vagas.
// Para somar uma fonte nova, basta implementar esta interface.
type Source interface {
	Name() string
	Fetch(ctx context.Context) ([]job.Job, error)
}

// getJSON faz um GET e decodifica a resposta em v.
func getJSON(ctx context.Context, c *http.Client, url string, v any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", userAgent)
	resp, err := c.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("status %d em %s", resp.StatusCode, req.URL.Host)
	}
	return json.NewDecoder(io.LimitReader(resp.Body, maxBody)).Decode(v)
}

// Result é o que uma fonte devolveu: vagas ou o erro.
type Result struct {
	Source string
	Jobs   []job.Job
	Err    error
}

// FetchAll busca em todas as fontes ao mesmo tempo. Uma fonte fora do ar não
// derruba as outras: cada uma volta com o seu próprio erro.
func FetchAll(ctx context.Context, sources []Source, perSource time.Duration) []Result {
	results := make([]Result, len(sources))
	var wg sync.WaitGroup
	for i, s := range sources {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ctx, cancel := context.WithTimeout(ctx, perSource)
			defer cancel()
			jobs, err := s.Fetch(ctx)
			results[i] = Result{Source: s.Name(), Jobs: jobs, Err: err}
		}()
	}
	wg.Wait()
	return results
}
