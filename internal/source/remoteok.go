package source

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/marcelin1555/vagas-radar/internal/job"
)

// RemoteOK lê https://remoteok.com/api.
type RemoteOK struct {
	BaseURL string
	Client  *http.Client
}

// NewRemoteOK cria a fonte com os valores de produção.
func NewRemoteOK() *RemoteOK {
	return &RemoteOK{BaseURL: "https://remoteok.com", Client: http.DefaultClient}
}

func (r *RemoteOK) Name() string { return "Remote OK" }

func (r *RemoteOK) Fetch(ctx context.Context) ([]job.Job, error) {
	// O primeiro item da lista é um aviso legal, não uma vaga. Lemos cada
	// item solto e pulamos os que não têm cargo.
	var raw []json.RawMessage
	if err := getJSON(ctx, r.Client, r.BaseURL+"/api", &raw); err != nil {
		return nil, err
	}
	out := make([]job.Job, 0, len(raw))
	for _, item := range raw {
		var j struct {
			ID       string   `json:"id"`
			URL      string   `json:"url"`
			Position string   `json:"position"`
			Company  string   `json:"company"`
			Tags     []string `json:"tags"`
			Location string   `json:"location"`
			Date     string   `json:"date"`
		}
		if err := json.Unmarshal(item, &j); err != nil || j.Position == "" {
			continue
		}
		posted, _ := time.Parse(time.RFC3339, j.Date)
		out = append(out, job.Job{
			ID: j.ID, Source: r.Name(), Title: j.Position, Company: j.Company,
			URL: j.URL, Location: j.Location, Tags: j.Tags, PostedAt: posted,
		})
	}
	return out, nil
}
