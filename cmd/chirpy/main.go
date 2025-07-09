package main

import (
	"os"
	"log"
	"net/http"
	"sync/atomic"
	"database/sql"
	_ "github.com/lib/pq"

	"github.com/joho/godotenv"

	"github.com/dubbersthehoser/httpserver/internal/database"
	
)

const ChirpSize int = 140
const servFiles string = "./servfiles"

type apiConfig struct {
	fileserverHits atomic.Int32
	DBQ *database.Queries
	Platform string
	JWTSecret string
	PolkaKey string
}

func main() {
	
	file, err := os.OpenFile("chirp.log", os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0o664)
	if err != nil {
		log.Fatal(err)
	}
	log.SetOutput(file)
	
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	jwtSecret := os.Getenv("JWT_SECRET_KEY")
	polkaKey := os.Getenv("POLKA_KEY")

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal(err)
	}

	dbQueries := database.New(db)

	conf := apiConfig{
		DBQ: dbQueries,
		Platform: os.Getenv("PLATFORM"),
		JWTSecret: jwtSecret,
		PolkaKey: polkaKey,
	}

	appHandler := http.StripPrefix("/app/", http.FileServer(http.Dir(servFiles + "/app")))
	appAssetsHandler := http.StripPrefix("/app/assets/", http.FileServer(http.Dir(servFiles + "/app/assets")))
	readinessHandler := http.HandlerFunc(ReadinessHandler)
	createChirpHandler := http.HandlerFunc(conf.CreateChirpHandler)
	getAllChirpHandler := http.HandlerFunc(conf.GetAllChirpsHandler)
	getAChirpHandler := http.HandlerFunc(conf.GetAChirpHandler)
	loginUserHandler := http.HandlerFunc(conf.LoginUserHandler)

	sMux := http.NewServeMux()

	sMux.Handle("/app/",                    conf.middlewareMetricsInc(appHandler))
	sMux.Handle("/app/assets/",             conf.middlewareMetricsInc(appAssetsHandler))
	sMux.Handle("GET /api/healthz",         conf.middlewareMetricsInc(readinessHandler))
	sMux.HandleFunc("POST /api/users",      conf.addUserHandler)
	sMux.HandleFunc("PUT /api/users",       conf.UpdateUserHandler)
	sMux.Handle("POST /api/chirps",         createChirpHandler)
	sMux.Handle("GET /api/chirps",          getAllChirpHandler)
	sMux.Handle("GET /api/chirps/{ChirpID}", getAChirpHandler)
	sMux.HandleFunc("DELETE /api/chirps/{ChirpID}", conf.RemoveChirpHandler)
	sMux.Handle("POST /api/login",          loginUserHandler)
	sMux.HandleFunc("POST /api/refresh",    conf.RefreshToken)
	sMux.HandleFunc("POST /api/revoke",     conf.RevokeToken)
	sMux.HandleFunc("POST /api/polka/webhooks", conf.PolkaHandler)
	sMux.HandleFunc("GET /admin/metrics",   conf.adminHandler)
	sMux.HandleFunc("POST /admin/reset",    conf.adminResetHandler)

	s := &http.Server{
		Addr: ":8080",
		Handler: sMux,
	}

	log.Fatal(s.ListenAndServe())
}
