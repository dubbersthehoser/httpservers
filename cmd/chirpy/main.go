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
	
	file, err := os.OpenFile("chirpy.log", os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0o664)
	if err != nil {
		log.Fatal(err)
	}
	log.SetOutput(file)
	
	godotenv.Load()

	dbURL := os.Getenv("DB_URL")
	jwtSecret := os.Getenv("JWT_SECRET_KEY")
	polkaKey := os.Getenv("POLKA_KEY")
	platform := os.Getenv("PLATFORM")

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal(err)
	}

	dbQueries := database.New(db)

	conf := apiConfig{
		DBQ: dbQueries,
		Platform: platform,
		JWTSecret: jwtSecret,
		PolkaKey: polkaKey,
	}


	sMux := http.NewServeMux()

	// Main Page
	appHandler := http.StripPrefix("/app/", http.FileServer(http.Dir(servFiles + "/app")))
	sMux.Handle("/app/", conf.middlewareMetricsInc(appHandler))

	// assets
	appAssetsHandler := http.StripPrefix("/app/assets/", http.FileServer(http.Dir(servFiles + "/app/assets")))
	sMux.Handle("/app/assets/", conf.middlewareMetricsInc(appAssetsHandler))

	// server status
	readinessHandler := http.HandlerFunc(ReadinessHandler)
	sMux.Handle("GET /api/healthz", conf.middlewareMetricsInc(readinessHandler))

	// users
	addUserHandler := http.HandlerFunc(conf.AddUserHandler)
	updateUserHandler := http.HandlerFunc(conf.UpdateUserHandler)

	sMux.Handle("POST /api/users", conf.middlewareMetricsInc(addUserHandler))
	sMux.Handle("PUT /api/users", conf.middlewareMetricsInc(updateUserHandler))

	// auth
	refreshToken := http.HandlerFunc(conf.RefreshToken)
	revokeToken := http.HandlerFunc(conf.RevokeToken)
	loginUserHandler := http.HandlerFunc(conf.LoginUserHandler)

	sMux.Handle("POST /api/refresh", conf.middlewareMetricsInc(refreshToken))
	sMux.Handle("POST /api/revoke", conf.middlewareMetricsInc(revokeToken))
	sMux.Handle("POST /api/login", conf.middlewareMetricsInc(loginUserHandler))

	// chirps / users posts
	createChirpHandler := http.HandlerFunc(conf.CreateChirpHandler)
	getAllChirpHandler := http.HandlerFunc(conf.GetAllChirpsHandler)
	getAChirpHandler := http.HandlerFunc(conf.GetAChirpHandler)
	removeAChirpHandler := http.HandlerFunc(conf.RemoveChirpHandler)

	sMux.Handle("POST /api/chirps", conf.middlewareMetricsInc(createChirpHandler))
	sMux.Handle("GET /api/chirps", conf.middlewareMetricsInc(getAllChirpHandler))
	sMux.Handle("GET /api/chirps/{ChirpID}", conf.middlewareMetricsInc(getAChirpHandler))
	sMux.Handle("DELETE /api/chirps/{ChirpID}", conf.middlewareMetricsInc(removeAChirpHandler))

	// admin
	sMux.HandleFunc("GET /admin/metrics", conf.AdminHandler)
	sMux.HandleFunc("POST /admin/reset", conf.AdminResetHandler)

	// chirpy red
	polkaHandler := http.HandlerFunc(conf.PolkaHandler)
	sMux.Handle("POST /api/polka/webhooks", conf.middlewareMetricsInc(polkaHandler))

	s := &http.Server{
		Addr: ":8080",
		Handler: sMux,
	}

	log.Fatal(s.ListenAndServe())
}
