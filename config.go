package main

import "os"

type Config struct {
	Port string
}

func NewConfig() *Config {
	port := os.Getenv("PORT")

	if port == "" {
		port = "8080"
	}

	return &Config{Port: port}
}
