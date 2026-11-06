// Package notify entrega as vagas novas: no terminal ou num webhook do Discord.
package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/marcelin1555/vagas-radar/internal/job"
)

// Notifier entrega uma lista de vagas.
type Notifier interface {
	Send(ctx context.Context, jobs []job.Job) error
}

// Console escreve no terminal.
type Console struct{ Out io.Writer }

func (c Console) Send(_ context.Context, jobs []job.Job) error {
	for _, j := range jobs {
		if _, err := fmt.Fprintln(c.Out, line(j)); err != nil {
			return err
		}
	}
	return nil
}

// Discord manda para um webhook. A API do Discord aceita até 2000 caracteres
// por mensagem, então a lista é dividida em blocos.
type Discord struct {
	URL    string
	Client *http.Client
}

const discordLimit = 1900 // margem abaixo dos 2000 do Discord

func (d Discord) Send(ctx context.Context, jobs []job.Job) error {
	for _, chunk := range Chunks(jobs, discordLimit) {
		body, _ := json.Marshal(map[string]string{"content": chunk})
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, d.URL, bytes.NewReader(body))
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "application/json")
		resp, err := d.Client.Do(req)
		if err != nil {
			// Não repetimos o erro cru: a URL do webhook é um segredo.
			return fmt.Errorf("enviar ao Discord: falha de rede")
		}
		resp.Body.Close()
		if resp.StatusCode/100 != 2 {
			return fmt.Errorf("Discord respondeu status %d", resp.StatusCode)
		}
	}
	return nil
}

func line(j job.Job) string {
	where := j.Location
	if where == "" {
		where = "remoto"
	}
	return fmt.Sprintf("**%s** | %s | %s\n<%s> (via %s)", j.Title, j.Company, where, j.URL, j.Source)
}

// Chunks junta as vagas em textos de no máximo limit caracteres.
func Chunks(jobs []job.Job, limit int) []string {
	var out []string
	var cur strings.Builder
	for _, j := range jobs {
		l := line(j)
		if len(l) > limit {
			l = l[:limit]
		}
		if cur.Len() > 0 && cur.Len()+len(l)+2 > limit {
			out = append(out, cur.String())
			cur.Reset()
		}
		if cur.Len() > 0 {
			cur.WriteString("\n\n")
		}
		cur.WriteString(l)
	}
	if cur.Len() > 0 {
		out = append(out, cur.String())
	}
	return out
}
