package sqlitevss

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	sqlite "github.com/mattn/go-sqlite3"
	"github.com/nlpodyssey/cybertron/pkg/models/bert"
	"github.com/nlpodyssey/cybertron/pkg/tasks"
	"github.com/nlpodyssey/cybertron/pkg/tasks/textencoding"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/vectorstores"
)

// #cgo linux,amd64 LDFLAGS:  -Wl,-undefined,dynamic_lookup -lstdc++
// #cgo darwin,amd64 LDFLAGS: -Wl,-undefined,dynamic_lookup -lomp
// #cgo darwin,arm64 LDFLAGS: -Wl,-undefined,dynamic_lookup -lomp
import "C"

var (
	// ErrMissingTextKey is returned in SimilaritySearch if a vector
	// from the query is missing the text key.
	ErrMissingTextKey = errors.New("missing text key in vector metadata")
	// ErrEmbedderWrongNumberVectors is returned when if the embedder returns a number
	// of vectors that is not equal to the number of documents given.
	ErrEmbedderWrongNumberVectors = errors.New(
		"number of vectors from embedder does not match number of documents",
	)
	// ErrEmptyResponse is returned if the API gives an empty response.
	ErrEmptyResponse         = errors.New("empty response")
	ErrInvalidResponse       = errors.New("invalid response")
	ErrInvalidScoreThreshold = errors.New(
		"score threshold must be between 0 and 1")
	ErrInvalidFilter = errors.New("invalid filter")
)

// SqliteVSS is a wrapper client around sqlite-vss exposing langchaingo compatible methods
type SqliteVSS struct {
	db  *sql.DB
	enc textencoding.Interface
}

var _ vectorstores.VectorStore = (*SqliteVSS)(nil)

func New(addr string) (*SqliteVSS, error) {

	conf := &tasks.Config{
		ModelsDir: os.TempDir(),
		ModelName: "sentence-transformers/all-MiniLM-L6-v2",
	}

	enc, err := tasks.Load[textencoding.Interface](conf)
	if err != nil {
		return nil, err
	}

	store := &SqliteVSS{nil, enc}
	// defer tasks.Finalize(m)

	dir, err := filepath.Abs("sqlite-vss/dist/release")
	if err != nil {
		return nil, err
	}

	sql.Register("sqlite-vss", &sqlite.SQLiteDriver{
		ConnectHook: func(conn *sqlite.SQLiteConn) error {
			if err := conn.RegisterFunc("st_encode", store.encode, true); err != nil {
				return err
			}
			return nil
		},
		Extensions: []string{
			filepath.Join(dir, "vector0"),
			filepath.Join(dir, "vss0"),
		},
	})

	db, err := sql.Open("sqlite-vss", addr)
	if err != nil {
		return nil, err
	}
	store.db = db

	query := `
CREATE TABLE IF NOT EXISTS docs(id INTEGER PRIMARY KEY AUTOINCREMENT, content TEXT NOT NULL, embeddings TEXT);
CREATE VIRTUAL TABLE IF NOT EXISTS vss_docs USING vss0(embeddings(384));`
	_, err = db.ExecContext(context.Background(), query)
	if err != nil {
		return nil, err
	}

	return store, nil
}

func (s *SqliteVSS) encode(text string) string {
	result, err := s.enc.Encode(context.Background(), text, int(bert.MeanPooling))
	if err != nil {
		panic(err)
	}
	vec := result.Vector.Data().F32()
	vecBytes, err := json.Marshal(vec)
	if err != nil {
		println(err)
	}
	return string(vecBytes)
}

func (s *SqliteVSS) AddDocuments(ctx context.Context, docs []schema.Document, _ ...vectorstores.Option) error {
	// tx, err := s.db.BeginTx(ctx, nil)
	// if err != nil {
	// 	return err
	// }

	count := len(docs)
	println("documents count:", count)

	for _, doc := range docs {
		// if i > 0 && i%100 == 0 {
		// 	perc := (float32(i) / float32(count)) * 100.0
		// 	fmt.Printf("- progress: %d / %d - %.2f%%\n", i, count, perc)
		// }

		res, err := s.db.ExecContext(ctx, "INSERT INTO docs(content) VALUES (?)", doc.PageContent)
		if err != nil {
			return err
		}

		id, err := res.LastInsertId()
		if err != nil {
			return err
		}

		_, err = s.db.ExecContext(ctx, "INSERT INTO vss_docs(rowid, embeddings) VALUES (?, st_encode(?))", id, doc.PageContent)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *SqliteVSS) SimilaritySearch(ctx context.Context, query string, numDocuments int, options ...vectorstores.Option) ([]schema.Document, error) {
	println("-- similarity search, query:", query)
	// data := s.encode(query)

	stmt := `WITH
matches AS (
	SELECT rowid, distance
	FROM vss_docs
	WHERE vss_search(embeddings, st_encode(?))
	LIMIT 5
),
final AS (
	SELECT docs.rowid, docs.content, matches.distance
	FROM matches
	LEFT JOIN docs ON docs.id = matches.rowid
	GROUP BY docs.content
)
SELECT rowid, distance, content
FROM final
ORDER BY distance`

	rows, err := s.db.QueryContext(ctx, stmt, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var docs []schema.Document
	for rows.Next() {
		var rowid int64
		var content string
		var distance float32
		if err = rows.Scan(&rowid, &distance, &content); err != nil {
			return nil, err
		}
		doc := schema.Document{PageContent: content}
		docs = append(docs, doc)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return docs, nil
}
