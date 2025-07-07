package main

import (
	"log"
	"fmt"
	"strings"
	"time"
	"errors"
	"net/http"
	"encoding/json"
	"database/sql"

	"github.com/google/uuid"

	"github.com/dubbersthehoser/httpserver/internal/database"
	"github.com/dubbersthehoser/httpserver/internal/auth"
)

type ReturnToUser struct {
	ID uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email string `json:"email"`
	Token string `json:"token"`
	RefreshToken string `json:"refresh_token"`
}

func somethingError(err error, w http.ResponseWriter) bool {
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, err = w.Write([]byte(`{"error":"Something went wrong"}`))
		if err != nil {
			log.Fatal(err)
		}
		log.Println(err)
		return true
	}
	return false
}

func fatalError(err error, w http.ResponseWriter) bool {
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, err = w.Write([]byte(`{"error":"Ouch!! Going down )-:"}`))
		if err != nil {
			log.Fatal(err)
		}
		log.Fatal(err)
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
		Password string `json:"password"`
	}
	
	p := params{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&p)
	if somethingError(err, w) {
		log.Printf("unable to decode: %s", r.Body)
		return
	}

	passhash, err := auth.HashPassword(p.Password)
	if somethingError(err, w) {
		log.Printf("unable to hash password: %s", p.Password)
		return
	}

	qParams := database.CreateUserParams{
		Email: p.Email,
		HashedPassword: passhash,

	}

	user, err := a.DBQ.CreateUser(r.Context(), qParams)
	if somethingError(err, w) {
		log.Printf("unable to create user: %#v", err)
		return
	}

	ruser := ReturnToUser{
		ID: user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email: user.Email,
	}

	jdata, err := json.Marshal(&ruser)
	if somethingError(err, w) {
		log.Printf("unable to marshal user json: %#v", ruser)
		return
	}

	w.WriteHeader(http.StatusCreated)
	_, err = w.Write(jdata)
	if err != nil {
		log.Fatal(err)
	}
}

func (a *apiConfig) LoginUserHandler(w http.ResponseWriter, r *http.Request) {
	type params struct {
		Email string `json:"email"`
		Password string `json:"password"`
	//	ExpiresInSeconds int64 `json:"expires_in_seconds"`
	}

	p := params{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&p)
	if somethingError(err, w) {
		log.Printf("unable to decode: %s", r.Body)
		return
	}

	// Get user from DB
	user, err := a.DBQ.GetUserByEmailWithPassword(r.Context(), p.Email)
	if errors.Is(err, sql.ErrNoRows) {
		w.WriteHeader(http.StatusNotFound)
		_, err = w.Write([]byte(`{"error": "User not found"}`))
		if err != nil {
			log.Fatal(err)
		}
		return
	} else if fatalError(err, w) {
		return
	}

	// Check Password
	if err := auth.CheckPasswordHash(user.HashedPassword, p.Password); err != nil {
		log.Printf("login: %s != %s", user.HashedPassword, p.Password)
		w.WriteHeader(http.StatusUnauthorized)
		_, err = w.Write([]byte(`{"error": "Invalid password"}`))
		if err != nil {
			log.Fatal(err)
		}
		return
	}


	//expires := time.Duration(p.ExpiresInSeconds)
	//if expires == 0 || expires > time.Hour {
	//	expires = time.Hour
	//}


	// Create JWT
	jwtExpires := time.Hour
	token, err := auth.MakeJWT(user.ID, a.JWTSecret, jwtExpires)
	if somethingError(err, w) {
		return
	}

	// Create Refresh Token
	refreshToken, err := auth.MakeRefreshToken()
	if err != nil {
		log.Panic(err)
	}
	refreshExpires := time.Now().Add((time.Hour * 24) * 60)

	qParams := database.CreateRefreshTokenParams{
		Token: refreshToken,
		ExpiresAt: refreshExpires,
		UserID: user.ID,
	}
	_, err = a.DBQ.CreateRefreshToken(r.Context(), qParams)
	if somethingError(err, w) {
		return
	}

	ruser := ReturnToUser{
		ID: user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email: user.Email,
		Token: token,
		RefreshToken: refreshToken,
	}

	jData, err := json.Marshal(&ruser)
	if somethingError(err, w) {
		log.Printf("unable to json.Marshal(user): %#v", ruser)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, err = w.Write(jData)
	if err != nil {
		log.Fatal(err)
	}
}

func (a *apiConfig) RefreshToken(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := auth.GetBearerToken(r.Header)
	if somethingError(err, w) {
		return
	}

	tok, err := a.DBQ.GetRefreshToken(r.Context(), refreshToken)
	if errors.Is(err, sql.ErrNoRows) {
		w.WriteHeader(http.StatusUnauthorized)
		_, err := w.Write([]byte(`{"error":"Unauthorized"}`))
		if err != nil {
			log.Fatal(err)
		}
		return
	} else if somethingError(err, w) {
		return
	}

	currTime := time.Now()
	if !currTime.Before(tok.ExpiresAt) {
		w.WriteHeader(http.StatusUnauthorized)
		_, err := w.Write([]byte(`{"error":"Unauthorized"}`))
		if err != nil {
			log.Fatal(err)
		}
		return
	}

	if tok.RevokedAt.Valid {
		w.WriteHeader(http.StatusUnauthorized)
		_, err := w.Write([]byte(`{"error":"Unauthorized"}`))
		if err != nil {
			log.Fatal(err)
		}
		return
	}

	// Create New JWT
	jwtExpires := time.Hour
	token, err := auth.MakeJWT(tok.UserID, a.JWTSecret, jwtExpires)
	if somethingError(err, w) {
		return
	}

	r.Header.Add("Content-Type", "application/json; charset=utf-8")
	ret := fmt.Sprintf(`{"token":"%s"}`, token)

	w.WriteHeader(http.StatusOK)
	_, err = w.Write([]byte(ret))
	if err != nil {
		log.Fatal(err)
	}
}
func (a *apiConfig) RevokeToken(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := auth.GetBearerToken(r.Header)
	if somethingError(err, w) {
		return
	}

	err = a.DBQ.RevokeToken(r.Context(), refreshToken)
	if somethingError(err, w) {
		return
	}
	w.WriteHeader(http.StatusNoContent)
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

	// Get JWT Token From Header
	token, err := auth.GetBearerToken(r.Header)
	if somethingError(err, w) {
		return
	}

	// Validate JWT
	uid, err := auth.ValidateJWT(token, a.JWTSecret)
	if err != nil {
		log.Printf("%s: token=%s\n\theader=", err, token, r.Header.Get("Authorization"))
		w.WriteHeader(http.StatusUnauthorized)
		_, err = w.Write([]byte(`{"error":"Invalid Token"}`))
		if err != nil {
			log.Fatal(err)
		}
		return
	}

	// Check The Size of Chirp
	if len(p.Body) > ChirpSize {
		w.WriteHeader(http.StatusBadRequest)
		_, err = w.Write([]byte(`{"error":"Chirp is too long"}`))
		if err != nil {
			log.Fatal(err)
		}
		return
	}

	// Senser Chirp
	words := strings.Split(p.Body, " ")
	for i, word := range words {
		switch strings.ToLower(word) {
			case "kerfuffle", "sharbert", "fornax":
				words[i] = "****"
		}
	}
	p.Body = strings.Join(words, " ")

	// Create Chirp
	qParams := database.CreateChirpParams{
		UserID: uid,
		Body: p.Body,
	}
	chirp, err := a.DBQ.CreateChirp(r.Context(), qParams)
	if err != nil {
		log.Fatal(err)
	}

	// Return to Client
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

