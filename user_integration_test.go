package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/testcontainers/testcontainers-go"
	tcmysql "github.com/testcontainers/testcontainers-go/modules/mysql"
)

// startMySQLContainer spins up a real MySQL instance in Docker and returns
// a DSN plus a cleanup func.
func startMySQLContainer(t *testing.T) (dsn string, cleanup func()) {
	t.Helper()
	ctx := context.Background()

	container, err := tcmysql.Run(
		ctx,
		"mysql:8.0",
		tcmysql.WithDatabase("testdb"),
		tcmysql.WithUsername("root"),
		tcmysql.WithPassword("password"),
	)
	if err != nil {
		t.Fatalf("failed to start mysql container: %v", err)
	}

	connStr, err := container.ConnectionString(ctx, "parseTime=true")
	if err != nil {
		t.Fatalf("failed to get connection string: %v", err)
	}

	cleanup = func() {
		if err := testcontainers.TerminateContainer(container); err != nil {
			t.Logf("failed to terminate container: %v", err)
		}
	}

	return connStr, cleanup
}

// applySchema runs schema.sql against the DB so the users table exists.
func applySchema(t *testing.T, db *sql.DB) {
	t.Helper()
	schema, err := os.ReadFile("schema.sql")
	if err != nil {
		t.Fatalf("failed to read schema.sql: %v", err)
	}
	if _, err := db.Exec(string(schema)); err != nil {
		t.Fatalf("failed to apply schema: %v", err)
	}
}

func TestUserCreateAndGet_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	dsn, cleanup := startMySQLContainer(t)
	defer cleanup()

	// Connect directly to run migrations before building the app.
	setupDB, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	defer setupDB.Close()

	// Wait for MySQL to actually be ready to accept connections.
	deadline := time.Now().Add(30 * time.Second)
	for {
		if err := setupDB.Ping(); err == nil {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("mysql did not become ready in time: %v", err)
		}
		time.Sleep(500 * time.Millisecond)
	}

	applySchema(t, setupDB)

	// Build the real app components directly (no fx.Invoke of the actual
	// HTTP listener — we just want the mux, tested via httptest).
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	repo := NewUserRepository(setupDB)
	mux := NewMux(repo)

	server := httptest.NewServer(mux)
	defer server.Close()

	// --- POST /users ---
	newUser := User{ID: "u1", Name: "Ada Lovelace"}
	body, _ := json.Marshal(newUser)

	resp, err := http.Post(server.URL+"/users", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("POST /users failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}

	// --- GET /users/{id} ---
	getResp, err := http.Get(fmt.Sprintf("%s/users/%s", server.URL, newUser.ID))
	if err != nil {
		t.Fatalf("GET /users/%s failed: %v", newUser.ID, err)
	}
	defer getResp.Body.Close()

	if getResp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", getResp.StatusCode)
	}

	var got User
	if err := json.NewDecoder(getResp.Body).Decode(&got); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if got.ID != newUser.ID || got.Name != newUser.Name {
		t.Fatalf("expected %+v, got %+v", newUser, got)
	}

	logger.Info("integration test passed", "user", got)
}
