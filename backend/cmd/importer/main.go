// Command importer loads the hasaneyldrm exercises dataset (text is MIT)
// into PostgreSQL. Idempotent: re-running updates instead of duplicating.
//
// Usage:
//
//	importer -url https://raw.githubusercontent.com/hasaneyldrm/exercises-dataset/main/data/exercises.json
//	importer -file exercises.json
//
// DATABASE_URL is required (same as the API).
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"miyano/internal/adapter/postgres"
	"miyano/internal/config"
)

// datasetRecord mirrors one entry of data/exercises.json.
type datasetRecord struct {
	ID               string            `json:"id"`
	Name             string            `json:"name"`
	Category         string            `json:"category"`
	BodyPart         string            `json:"body_part"`
	Target           string            `json:"target"`
	MuscleGroup      string            `json:"muscle_group"`
	Equipment        string            `json:"equipment"`
	Instructions     map[string]string `json:"instructions"`
	InstructionSteps map[string][]string `json:"instruction_steps"`
	Image            string            `json:"image"`
	GifURL           string            `json:"gif_url"`
	Attribution      string            `json:"attribution"`
}

const defaultURL = "https://raw.githubusercontent.com/hasaneyldrm/exercises-dataset/main/data/exercises.json"

// langs controls which translations are imported (es first, en as fallback info).
var langs = []string{"es", "en"}

func main() {
	file := flag.String("file", "", "path to exercises.json (alternative to -url)")
	url := flag.String("url", defaultURL, "URL of exercises.json")
	flag.Parse()

	cfg := config.Load()
	if err := cfg.Validate(); err != nil {
		log.Fatalf("invalid configuration: %v", err)
	}

	data, err := fetch(*file, *url)
	if err != nil {
		log.Fatalf("fetching dataset: %v", err)
	}
	var records []datasetRecord
	if err := json.Unmarshal(data, &records); err != nil {
		log.Fatalf("parsing dataset: %v", err)
	}
	log.Printf("parsed %d exercises", len(records))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	pool, err := postgres.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database unavailable: %v", err)
	}
	defer pool.Close()

	nEx, nTr, err := importAll(ctx, pool, records)
	if err != nil {
		log.Fatalf("import failed: %v", err)
	}
	log.Printf("done: %d exercises, %d translations imported", nEx, nTr)
}

func fetch(file, url string) ([]byte, error) {
	if file != "" {
		// #nosec G304 -- file path comes from the operator's own -file flag (local CLI tool).
		return os.ReadFile(file)
	}
	if !strings.HasPrefix(url, "http") {
		return nil, fmt.Errorf("unsupported source: %q", url)
	}
	// #nosec G107 -- url comes from the operator's own -url flag (local CLI tool, 5min outer timeout).
	resp, err := http.Get(url) //nolint:noctx // one-shot CLI with its own outer timeout
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d fetching %s", resp.StatusCode, url)
	}
	return io.ReadAll(resp.Body)
}

func importAll(ctx context.Context, pool *pgxpool.Pool, records []datasetRecord) (int, int, error) {
	batch := &pgx.Batch{}
	queuedTr := 0

	for _, r := range records {
		batch.Queue(`
			INSERT INTO exercises (id, name, category, body_part, target, muscle_group, equipment, image_url, gif_url, attribution)
			VALUES ($1, $2, $3, $4, nullif($5, ''), nullif($6, ''), $7, nullif($8, ''), nullif($9, ''), nullif($10, ''))
			ON CONFLICT (id) DO UPDATE SET
				name = EXCLUDED.name, category = EXCLUDED.category, body_part = EXCLUDED.body_part,
				target = EXCLUDED.target, muscle_group = EXCLUDED.muscle_group, equipment = EXCLUDED.equipment,
				image_url = EXCLUDED.image_url, gif_url = EXCLUDED.gif_url, attribution = EXCLUDED.attribution`,
			r.ID, r.Name, r.Category, r.BodyPart, r.Target, r.MuscleGroup, r.Equipment, r.Image, r.GifURL, r.Attribution)

		for _, lang := range langs {
			steps, hasSteps := r.InstructionSteps[lang]
			text, hasText := r.Instructions[lang]
			if !hasSteps && !hasText {
				continue
			}
			var stepsJSON any
			if hasSteps {
				raw, err := json.Marshal(steps)
				if err != nil {
					return 0, 0, fmt.Errorf("marshaling steps for %s/%s: %w", r.ID, lang, err)
				}
				stepsJSON = raw
			}
			batch.Queue(`
				INSERT INTO exercise_translations (exercise_id, lang, instructions, instruction_steps)
				VALUES ($1, $2, nullif($3, ''), $4)
				ON CONFLICT (exercise_id, lang) DO UPDATE SET
					instructions = EXCLUDED.instructions, instruction_steps = EXCLUDED.instruction_steps`,
				r.ID, lang, text, stepsJSON)
			queuedTr++
		}
	}

	br := pool.SendBatch(ctx, batch)
	defer br.Close()

	for i := 0; i < len(records)+queuedTr; i++ {
		if _, err := br.Exec(); err != nil {
			return 0, 0, fmt.Errorf("batch statement %d failed: %w", i, err)
		}
	}
	return len(records), queuedTr, nil
}
