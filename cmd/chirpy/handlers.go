package main

import (
	"log"
	"fmt"
	"strings"
	"errors"
	"net/http"
	"encoding/json"
	"database/sql"

	"github.com/google/uuid"

	"github.com/dubbersthehoser/httpserver/internal/database"
)

func somethingError(err error, w http.ResponseWriter) bool {
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, err = w.Write([]byte(`{"error":"Something went wrong"}`))
		if err != nil {
			log.Fatal(err)
		}
		return true
	}
	return false
}

func (a *apiConfig) adminHandler(w http.ResponseWriter, r *http.Request) {
	r.Header.Add("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	body := `
<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
  </body>
</html>
`
	_, err := w.Write([]byte(fmt.Sprintf(body, a.fileserverHits.Load())))
	if err != nil {
		log.Fatal(err)
	}
}

func (a *apiConfig) adminResetHandler(w http.ResponseWriter, r *http.Request) {

	var err error
	if a.Platform != "dev" {
		w.WriteHeader(http.StatusForbidden)
		_, err = w.Write([]byte("Metrics and Database Reset: Unsuccsessful: Forbidden"))
		if err != nil {
			log.Fatal(err)
		}
		return
	}


	_ = a.fileserverHits.Swap(0)

	err = a.DBQ.DeleteAllUsers(r.Context())
	if err != nil {
		log.Fatal(err)
	}

	r.Header.Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write([]byte("Metrics and Database Reset: Succsessful"))
	if err != nil {
		log.Fatal(err)
	}
}

func (a *apiConfig) addUserHandler(w http.ResponseWriter, r *http.Request) {

	type params struct { 
		Email string `json:"email"`
	}
	
	p := params{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&p)
	if somethingError(err, w) {
		log.Printf("unable to decode: %s", r.Body)
		return
	}

	user, err := a.DBQ.CreateUser(r.Context(), p.Email)
	if somethingError(err, w) {
		log.Printf("unable to create user: %#v", err)
		return
	}

	jdata, err := json.Marshal(&user)
	if somethingError(err, w) {
		log.Printf(" unable to marshal user json: %#v", user)
		return
	}

	w.WriteHeader(http.StatusCreated)
	_, err = w.Write(jdata)
	if err != nil {
		log.Fatal(err)
	}
}

func (a *apiConfig) GetAChirpHandler(w http.ResponseWriter, r *http.Request) {
	
	chirpID := r.PathValue("ChirpID")

	log.Println(chirpID)

	chirpUUID, err := uuid.Parse(chirpID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, err = w.Write([]byte(`{"error": "Invalid chirp id"}`))
		if err != nil {
			log.Fatal(err)
		}
		return
	}


	chirp, err := a.DBQ.GetAChirp(r.Context(), chirpUUID)
	if errors.Is(err, sql.ErrNoRows) {
		w.WriteHeader(http.StatusNotFound)
		_, err = w.Write([]byte(`{"error": "Chirp id not found"}`))
		if err != nil {
			log.Fatal(err)
		}
		return
	} else if err != nil {
		log.Fatal(err)
	}

	jData, err := json.Marshal(&chirp)
	if err != nil {
		log.Fatal(err)
	}

	w.WriteHeader(http.StatusOK)
	_, err = w.Write(jData)
	if err != nil {
		log.Fatal(err)
	}
}

func (a *apiConfig) GetAllChirpsHandler(w http.ResponseWriter, r *http.Request) {
	
	chirps, err := a.DBQ.GetAllChirps(r.Context())
	if somethingError(err, w) {
		return
	}

	jData, err := json.Marshal(&chirps)
	if err != nil {
		log.Fatal(err)
	}
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(jData)
	if err != nil {
		log.Fatal(err)
	}
}

func (a *apiConfig) CreateChirpHandler(w http.ResponseWriter, r *http.Request) {

	type params struct {
		Body string `json:"body"`
		UserID string `json:"user_id"`
	}

	decoder := json.NewDecoder(r.Body)
	p := params{}
	err := decoder.Decode(&p)
	if somethingError(err, w) {
		return
	}

	if len(p.Body) > ChirpSize {
		w.WriteHeader(http.StatusBadRequest)
		_, err = w.Write([]byte(`{"error":"Chirp is too long"}`))
		if err != nil {
			log.Fatal(err)
		}
		return
	}
	
	words := strings.Split(p.Body, " ")

	for i, word := range words {
		switch strings.ToLower(word) {
			case "kerfuffle", "sharbert", "fornax":
				words[i] = "****"
		}
	}

	p.Body = strings.Join(words, " ")

	userID, err := uuid.Parse(p.UserID)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, err = w.Write([]byte(`{"error":"invalid user_id"}`))
		return
	}

	qParams := database.CreateChirpParams{
		UserID: userID,
		Body: p.Body,
	}

	chirp, err := a.DBQ.CreateChirp(r.Context(), qParams)

	if err != nil {
		log.Fatal(err)
	}

	data, _ :=  json.Marshal(&chirp)

	w.WriteHeader(http.StatusCreated)
	_, err = w.Write(data)
	if err != nil {
		log.Fatal(err)
	}
}

func ReadinessHandler(w http.ResponseWriter, r *http.Request) {
	r.Header.Add("Content-Type",  "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte("OK"))
	if err != nil {
		log.Fatal(err)
	}
}

