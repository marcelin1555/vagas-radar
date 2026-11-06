// Comando radar: busca vagas remotas, filtra e avisa só das novas.
package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/marcelin1555/vagas-radar/internal/job"
	"github.com/marcelin1555/vagas-radar/internal/notify"
	"github.com/marcelin1555/vagas-radar/internal/seen"
	"github.com/marcelin1555/vagas-radar/internal/source"
)

type options struct {
	db       string
	interval time.Duration
	include  string
	exclude  string
	max      int
	dryRun   bool
}

func main() {
	var o options
	flag.StringVar(&o.db, "db", "radar.db", "arquivo do SQLite com as vagas já avisadas")
	flag.DurationVar(&o.interval, "interval", 0, "de quanto em quanto tempo repetir (ex: 30m). 0 roda uma vez")
	flag.StringVar(&o.include, "include", "python,go,golang,lua,luau,roblox,ai,llm,machine learning,junior,intern,estágio",
		"termos que interessam, separados por vírgula")
	flag.StringVar(&o.exclude, "exclude", "senior,sr,staff,principal,lead,manager,director,head",
		"termos que descartam a vaga pelo título")
	flag.IntVar(&o.max, "max", 15, "máximo de vagas por aviso")
	flag.BoolVar(&o.dryRun, "dry-run", false, "mostra no terminal e não marca as vagas como vistas")
	flag.Parse()

	log := slog.New(slog.NewTextHandler(os.Stderr, nil))
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := run(ctx, o, log); err != nil {
		log.Error("falha", "err", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, o options, log *slog.Logger) error {
	store, err := seen.Open(o.db)
	if err != nil {
		return err
	}
	defer store.Close()

	// O webhook é segredo: vem só de variável de ambiente, nunca de flag
	// (flags aparecem na lista de processos).
	var out notify.Notifier = notify.Console{Out: os.Stdout}
	if url := os.Getenv("DISCORD_WEBHOOK_URL"); url != "" && !o.dryRun {
		out = notify.Discord{URL: url, Client: &http.Client{Timeout: 15 * time.Second}}
	}

	sources := []source.Source{source.NewRemotive(), source.NewRemoteOK()}
	filter := job.NewFilter(job.Split(o.include), job.Split(o.exclude))

	once := func() {
		if err := cycle(ctx, sources, filter, store, out, o, log); err != nil {
			log.Error("ciclo falhou", "err", err)
		}
	}
	once()
	if o.interval <= 0 {
		return nil
	}
	t := time.NewTicker(o.interval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-t.C:
			once()
		}
	}
}

func cycle(ctx context.Context, sources []source.Source, filter *job.Filter,
	store *seen.Store, out notify.Notifier, o options, log *slog.Logger) error {

	var matched []job.Job
	for _, r := range source.FetchAll(ctx, sources, 30*time.Second) {
		if r.Err != nil {
			log.Warn("fonte falhou", "fonte", r.Source, "err", r.Err)
			continue
		}
		n := 0
		for _, j := range r.Jobs {
			if filter.Match(j) {
				matched = append(matched, j)
				n++
			}
		}
		log.Info("fonte lida", "fonte", r.Source, "vagas", len(r.Jobs), "combinam", n)
	}

	fresh, err := store.Unseen(ctx, matched)
	if err != nil {
		return err
	}
	if len(fresh) == 0 {
		log.Info("nenhuma vaga nova")
		return nil
	}
	if len(fresh) > o.max {
		// As que sobraram continuam "não vistas" e saem no próximo ciclo.
		fresh = fresh[:o.max]
	}
	if err := out.Send(ctx, fresh); err != nil {
		return fmt.Errorf("avisar: %w", err)
	}
	if o.dryRun {
		return nil
	}
	return store.Mark(ctx, fresh)
}
