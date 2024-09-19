package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime/debug"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/thrashdev/blog-aggregator/internal/database"
)

type authedHandler func(http.ResponseWriter, *http.Request, database.User)

type apiConfig struct {
	DB *database.Queries
}

type userDTO struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Name      string    `json:"name"`
	Apikey    string    `json:"api_key"`
}

type feedDTO struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Name      string    `json:"name"`
	Url       string    `json:"url"`
	UserID    uuid.UUID `json:"user_id"`
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) error {
	resp, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshalling JSON: %v ", err)
		w.WriteHeader(500)
		return errors.New(fmt.Sprintf("Couldn't marshall json %v", payload))
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(resp)
	return nil
}

func respondWithError(w http.ResponseWriter, code int, msg string) error {
	type errResponse struct {
		Error string `json:"error"`
	}
	resp := errResponse{msg}
	err := respondWithJSON(w, code, resp)
	if err != nil {
		return err
	}
	return nil
}

func handlerReadiness(w http.ResponseWriter, r *http.Request) {
	type responseJSON struct {
		Status string `json:"status"`
	}
	resp := responseJSON{"OK"}
	err := respondWithJSON(w, 200, resp)
	if err != nil {
		debug.PrintStack()
		log.Println(err)
	}
}

func handlerErrResp(w http.ResponseWriter, r *http.Request) {
	err := respondWithError(w, 500, "Internal Server Error")
	printError(err)
}

func printError(err error) {
	if err != nil {
		log.Println(err)
		debug.PrintStack()
	}
}

func createUserHandler(config apiConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		type parameters struct {
			Name string `json:"name"`
		}
		decoder := json.NewDecoder(r.Body)
		params := parameters{}
		err := decoder.Decode(&params)
		if err != nil {
			err_msg := fmt.Sprintf("Error decoding parameters %v ", err)
			log.Printf(err_msg)
			respondWithError(w, 500, err_msg)
			return
		}
		ctx := r.Context()
		uid := uuid.New()
		name := sql.NullString{String: params.Name, Valid: true}
		arg := database.CreateUserParams{ID: uid, CreatedAt: time.Now(), UpdatedAt: time.Now(), Name: name}
		user, err := config.DB.CreateUser(ctx, arg)
		printError(err)
		respondWithJSON(w, 200, user)
		// config.DB.CreateUser()
	}
}

func (config *apiConfig) getUserByApiKeyHandler(w http.ResponseWriter, r *http.Request, user database.User) {
	userDto := userDTO{ID: user.ID, CreatedAt: user.CreatedAt, UpdatedAt: user.UpdatedAt, Name: user.Name.String, Apikey: user.Apikey}
	respondWithJSON(w, 200, userDto)
}

func (config *apiConfig) middlewareAuth(handler authedHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		apiKey := strings.Replace(authHeader, "ApiKey ", "", 1)
		ctx := r.Context()
		user, err := config.DB.GetUserByApiKey(ctx, apiKey)
		printError(err)
		if user == (database.User{}) {
			respondWithError(w, 400, "User not found")
			return
		}

		handler(w, r, user)
	}
}

func (config *apiConfig) createFeedHandler(w http.ResponseWriter, r *http.Request, user database.User) {
	type parameters struct {
		Name string `json:"name"`
		Url  string `json:"url"`
	}
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		err_msg := fmt.Sprintf("Error decoding parameters %v ", err)
		log.Printf(err_msg)
		respondWithError(w, 500, err_msg)
		return
	}
	name := sql.NullString{String: params.Name, Valid: true}
	url := sql.NullString{String: params.Url, Valid: true}
	newFeed := database.CreateFeedParams{ID: uuid.New(), CreatedAt: time.Now(), UpdatedAt: time.Now(), Name: name, Url: url, UserID: user.ID}
	ctx := r.Context()
	feed, err := config.DB.CreateFeed(ctx, newFeed)
	printError(err)
	feedDto := feedDTO{ID: feed.ID, CreatedAt: feed.CreatedAt, UpdatedAt: feed.UpdatedAt, Name: feed.Name.String, Url: feed.Name.String, UserID: feed.UserID}
	respondWithJSON(w, 200, feedDto)

}

func (config *apiConfig) getFeedsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	feeds, err := config.DB.GetFeeds(ctx)
	printError(err)
	respondWithJSON(w, 200, feeds)

}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}
	port := os.Getenv("PORT")
	dbURL := os.Getenv("connection_string")
	db, err := sql.Open("postgres", dbURL)
	dbQueries := database.New(db)
	config := apiConfig{DB: dbQueries}
	serveMux := http.NewServeMux()
	serveMux.HandleFunc("GET /v1/healthz", handlerReadiness)
	serveMux.HandleFunc("GET /v1/err", handlerErrResp)
	serveMux.HandleFunc("POST /v1/users", createUserHandler(config))
	serveMux.HandleFunc("GET /v1/users", config.middlewareAuth(config.getUserByApiKeyHandler))
	serveMux.HandleFunc("POST /v1/feeds", config.middlewareAuth(config.createFeedHandler))
	serveMux.HandleFunc("GET /v1/feeds", config.getFeedsHandler)
	server := http.Server{Handler: serveMux, Addr: ":" + port}
	server.ListenAndServe()
}
