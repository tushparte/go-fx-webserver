package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/testcontainers/testcontainers-go"
	tcmysql "github.com/testcontainers/testcontainers-go/modules/mysql"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

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

func waitForReady(t *testing.T, db *sql.DB) {
	t.Helper()
	deadline := time.Now().Add(30 * time.Second)
	for {
		if err := db.Ping(); err == nil {
			return
		}
		if time.Now().After(deadline) {
			t.Fatal("mysql did not become ready in time")
		}
		time.Sleep(500 * time.Millisecond)
	}
}

func TestUserCreateAndGet_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	dsn, cleanup := startMySQLContainer(t)
	defer cleanup()

	setupDB, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	defer setupDB.Close()

	waitForReady(t, setupDB)
	applySchema(t, setupDB)

	// Test config: fixed DSN from the container, port 0 so the OS
	// picks a free port (avoids clashing with anything else running).
	cfg := &Config{
		Port:              "0",
		DSN:               dsn,
		ReadTimeout:       5 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	var listener net.Listener

	app := fxtest.New(
		t,
		fx.Supply(cfg), // overrides NewConfig — no NewConfig in this list
		fx.Provide(
			NewLogger,
			NewDB,
			NewUserRepository,
			NewMux,
			NewListener,
			NewHTTPServer,
		),
		fx.Populate(&listener), // pulls the constructed net.Listener out of the graph
		fx.Invoke(func(*http.Server) {}),
	)

	app.RequireStart()
	defer app.RequireStop()

	baseURL := fmt.Sprintf("http://%s", listener.Addr().String())

	// --- POST /users ---
	newUser := User{ID: "u1", Name: "Ada Lovelace"}
	body, _ := json.Marshal(newUser)

	resp, err := http.Post(baseURL+"/users", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("POST /users failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}

	// --- GET /users/{id} ---
	getResp, err := http.Get(fmt.Sprintf("%s/users/%s", baseURL, newUser.ID))
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
}
