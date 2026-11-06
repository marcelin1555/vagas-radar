package source

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/marcelin1555/vagas-radar/internal/job"
)

// Remotive lê https://remotive.com/api/remote-jobs. O parâmetro de categoria
// da API não teve efeito nos meus testes, então quem filtra é o job.Filter.
type Remotive struct {
	BaseURL string
	Client  *http.Client
}

// NewRemotive cria a fonte com os valores de produção.
func NewRemotive() *Remotive {
	return &Remotive{BaseURL: "https://remotive.com", Client: http.DefaultClient}
}

func (r *Remotive) Name() string { return "Remotive" }

func (r *Remotive) Fetch(ctx context.Context) ([]job.Job, error) {
	var body struct {
		Jobs []struct {
			ID          int      `json:"id"`
			URL         string   `json:"url"`
			Title       string   `json:"title"`
			Company     string   `json:"company_name"`
			Tags        []string `json:"tags"`
			Location    string   `json:"candidate_required_location"`
			PublishedAt string   `json:"publication_date"`
		} `json:"jobs"`
	}
	if err := getJSON(ctx, r.Client, r.BaseURL+"/api/remote-jobs", &body); err != nil {
		return nil, err
	}
	out := make([]job.Job, 0, len(body.Jobs))
	for _, j := range body.Jobs {
		posted, _ := time.Parse("2006-01-02T15:04:05", j.PublishedAt)
		out = append(out, job.Job{
			ID: strconv.Itoa(j.ID), Source: r.Name(), Title: j.Title, Company: j.Company,
			URL: j.URL, Location: j.Location, Tags: j.Tags, PostedAt: posted,
		})
	}
	return out, nil
}
