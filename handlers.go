package main

import (
	"encoding/json"
	"net/http"
)

func NewMux(userRepo *UserRepository) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	mux.HandleFunc("POST /users", func(w http.ResponseWriter, r *http.Request) {
		var u User
		if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
			http.Error(w, "invalid body", http.StatusBadRequest)
			return
		}

		if err := userRepo.Create(r.Context(), &u); err != nil {
			http.Error(w, "failed to create user", http.StatusInternalServerError)
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(u)
	})

	mux.HandleFunc("GET /users/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		user, err := userRepo.GetByID(r.Context(), id)
		if err != nil {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(user)
	})

	return mux
}
