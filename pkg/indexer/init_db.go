package indexer

import (
	"context"
	"database/sql"
	_ "embed"
	"errors"
	"strings"

	"github.com/hugr-lab/query-engine/pkg/data-sources/sources"
	"github.com/hugr-lab/query-engine/pkg/db"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/marcboeker/go-duckdb/v2"
	_ "github.com/marcboeker/go-duckdb/v2"
)

//go:embed schema.sql
var initSchema string

func InitDB(ctx context.Context, path string, vectorSize int) error {
	dbType := db.SDBDuckDB
	if strings.HasPrefix(path, "postgres://") {
		dbType = db.SDBPostgres
	}
	return createDB(ctx, dbType, path, vectorSize)
}

type dbInitParams struct {
	DBVersion         string
	VectorSize        int
	EmbeddingsEnabled bool
	EmbeddingModel    string
}

func createDB(ctx context.Context, dbType db.ScriptDBType, dbPath string, vectorSize int) error {
	initSQL, err := db.ParseSQLScriptTemplate(dbType, initSchema, dbInitParams{
		DBVersion:  dbVersion,
		VectorSize: vectorSize,
	})
	if err != nil {
		return err
	}
	switch dbType {
	case db.SDBPostgres:
		// try to create the database (need to connect to the postgres database)
		dbDSN, err := sources.ParseDSN(dbPath)
		if err != nil {
			return err
		}
		dbName := dbDSN.DBName
		dbDSN.DBName = "postgres"
		d, err := sql.Open("pgx", dbDSN.String())
		if err != nil {
			return err
		}
		_, err = d.Exec("CREATE DATABASE \"" + dbName + "\";")
		d.Close()
		if err != nil {
			return err
		}
		d, err = sql.Open("pgx", dbPath)
		if err != nil {
			return err
		}
		_, err = d.ExecContext(ctx, initSQL)
		return err
	case db.SDBDuckDB:
		if strings.HasPrefix(dbPath, "s3://") {
			return errors.New("database is in readonly mode (s3)")
		}
		conn, err := duckdb.NewConnector(dbPath, nil)
		if err != nil {
			return err
		}
		defer conn.Close()
		d := sql.OpenDB(conn)
		defer d.Close()
		_, err = d.ExecContext(ctx, initSQL)
		return err
	default:
		return errors.New("unsupported database type")
	}
}
