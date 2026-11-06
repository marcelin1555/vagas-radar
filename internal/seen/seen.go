// Package seen lembra quais vagas já foram avisadas, para não repetir.
package seen

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/marcelin1555/vagas-radar/internal/job"
	_ "modernc.org/sqlite"
)

// Store guarda as chaves das vagas já enviadas em SQLite.
type Store struct{ db *sql.DB }

// Open abre (e cria, se preciso) o banco.
func Open(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("abrir banco: %w", err)
	}
	db.SetMaxOpenConns(1)
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS seen (
		key     TEXT PRIMARY KEY,
		seen_at TEXT NOT NULL
	)`)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("criar tabela: %w", err)
	}
	return &Store{db: db}, nil
}

// Close fecha o banco.
func (s *Store) Close() error { return s.db.Close() }

// Unseen devolve só as vagas que ainda não foram marcadas, sem repetir
// a mesma vaga duas vezes na lista.
func (s *Store) Unseen(ctx context.Context, jobs []job.Job) ([]job.Job, error) {
	var out []job.Job
	dup := map[string]bool{}
	for _, j := range jobs {
		k := j.Key()
		if dup[k] {
			continue
		}
		dup[k] = true
		var one int
		err := s.db.QueryRowContext(ctx, `SELECT 1 FROM seen WHERE key = ?`, k).Scan(&one)
		if err == sql.ErrNoRows {
			out = append(out, j)
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("consultar: %w", err)
		}
	}
	return out, nil
}

// Mark registra as vagas como já enviadas. Só chame depois de o aviso sair,
// senão uma falha de rede faria a vaga se perder.
func (s *Store) Mark(ctx context.Context, jobs []job.Job) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	now := time.Now().UTC().Format(time.RFC3339)
	for _, j := range jobs {
		if _, err := tx.ExecContext(ctx,
			`INSERT OR IGNORE INTO seen (key, seen_at) VALUES (?, ?)`, j.Key(), now); err != nil {
			return fmt.Errorf("marcar: %w", err)
		}
	}
	return tx.Commit()
}
