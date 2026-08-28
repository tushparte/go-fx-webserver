package main

import (
	"os"
	"time"
)

type Config struct {
	Port string

	ReadTimeout       time.Duration
	ReadHeaderTimeout time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration

	DSN string
}

func NewConfig() *Config {
	port := os.Getenv("PORT")

	if port == "" {
		port = "8080"
	}

	dsn := os.Getenv("MYSQL_DSN")
	if dsn == "" {
		dsn = "root:root@tcp(127.0.0.1:3306)/myapp?parseTime=true"
	}

	return &Config{
		Port:              port,
		ReadTimeout:       5 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       120 * time.Second,
		DSN:               dsn,
	}
}
