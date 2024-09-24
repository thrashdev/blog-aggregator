package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"

	"runtime/debug"
	"strings"
	"time"

	"github.com/google/uuid"
	// "github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/thrashdev/blog-aggregator/internal/config"
	"github.com/thrashdev/blog-aggregator/internal/database"
	"github.com/thrashdev/blog-aggregator/internal/rss"
)

type state struct {
	cfg *config.Config
}

type command struct {
	name      string
	arguments []string
}

type commands struct {
	handlers map[string]func(*state, command) error
}

func (c *commands) register(name string, f func(*state, command) error) {
	c.handlers[name] = f
}

func (c *commands) run(s *state, cmd command) error {
	fmt.Println(c.handlers)
	fmt.Println(cmd.name)
	handler, ok := c.handlers[cmd.name]
	if !ok {
		return errors.New(fmt.Sprintf("Command %v does not exist", cmd.name))
	}
	err := handler(s, cmd)
	if err != nil {
		return err
	}
	return nil
}

func handlerLogin(s *state, cmd command) error {
	if len(cmd.arguments) == 0 {
		return errors.New("Login expects a username")
	}
	username := cmd.arguments[0]
	err := s.cfg.SetUser(username)
	if err != nil {
		log.Println(err)
		return err
	}
	fmt.Printf("User %v has logged in\n", username)
	return nil
}

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

type feedFollowDTO struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	FeedID    uuid.UUID `json:"feed_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
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
	if err != nil {
		printError(err)
		respondWithError(w, 500, err.Error())
	}
	feedDto := feedDTO{ID: feed.ID, CreatedAt: feed.CreatedAt, UpdatedAt: feed.UpdatedAt, Name: feed.Name.String, Url: feed.Url.String, UserID: feed.UserID}
	newFeedFollow := database.CreateFeedFollowParams{ID: uuid.New(), UserID: user.ID, FeedID: feed.ID, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	feedFollow, err := config.DB.CreateFeedFollow(ctx, newFeedFollow)
	if err != nil {
		printError(err)
		respondWithError(w, 500, err.Error())
	}
	feedFollowDto := feedFollowDTO{ID: feedFollow.ID, UserID: feedFollow.UserID, FeedID: feedFollow.FeedID, CreatedAt: feedFollow.CreatedAt, UpdatedAt: feedFollow.UpdatedAt}
	type jsonResponse struct {
		Feed       feedDTO       `json:"feed"`
		FeedFollow feedFollowDTO `json:"feed_follow"`
	}
	resp := jsonResponse{Feed: feedDto, FeedFollow: feedFollowDto}
	respondWithJSON(w, 200, resp)

}

func (config *apiConfig) getFeedsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	feeds, err := config.DB.GetFeeds(ctx)
	printError(err)
	respondWithJSON(w, 200, feeds)

}

func (config *apiConfig) createFeedFollowHandler(w http.ResponseWriter, r *http.Request, user database.User) {
	type parameters struct {
		FeedID uuid.UUID `json:"feed_id"`
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
	newFeedFollow := database.CreateFeedFollowParams{ID: uuid.New(), UserID: user.ID, FeedID: params.FeedID, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	feedFollow, err := config.DB.CreateFeedFollow(ctx, newFeedFollow)
	printError(err)
	if err != nil {
		respondWithError(w, 500, err.Error())
		return
	}
	respondWithJSON(w, 200, feedFollow)

}

func (config *apiConfig) deleteFeedFollowHandler(w http.ResponseWriter, r *http.Request, user database.User) {
	stringFeedFollowID := r.PathValue("feedFollowID")
	feedFollowID, err := uuid.Parse(stringFeedFollowID)
	printError(err)
	if err != nil {
		respondWithError(w, 500, err.Error())
		return
	}
	ctx := r.Context()
	feedFollows, err := config.DB.GetFeedFollowsByUserID(ctx, user.ID)
	feedFollowToDelete := database.FeedFollow{}
	for _, feedFollow := range feedFollows {
		if feedFollow.ID == feedFollowID && feedFollow.UserID == user.ID {
			feedFollowToDelete = feedFollow
		}
	}
	if feedFollowToDelete == (database.FeedFollow{}) {
		respondWithError(w, 400, "Feed follow not found")
		return
	}
	err = config.DB.DeleteFeedFollowsByID(ctx, feedFollowID)
	if err != nil {
		printError(err)
		respondWithError(w, 500, err.Error())
		return
	}
	w.WriteHeader(204)

}

func (config *apiConfig) getFeedFollowsHandler(w http.ResponseWriter, r *http.Request, user database.User) {
	ctx := r.Context()
	feedFollows, err := config.DB.GetFeedFollowsByUserID(ctx, user.ID)
	printError(err)
	if err != nil {
		respondWithError(w, 500, err.Error())
		return
	}
	respondWithJSON(w, 200, feedFollows)
}

func (config *apiConfig) workerFeedsFetch(feedstoFetch int) {
	ticker := time.NewTicker(60 * time.Second)
	fmt.Println("Ticker started")
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			fmt.Println("Worker is running...")
			var wg sync.WaitGroup
			ctx := context.Background()

			feeds, err := config.DB.GetNextFeedsToFetch(ctx, int32(feedstoFetch))
			if err != nil {
				log.Fatalf(err.Error())
			}

			for _, nextFeed := range feeds {
				wg.Add(1)
				go func(feedUrl string) {
					feed, err := rss.FetchFeed(feedUrl)
					if err != nil {
						log.Println("Error encountered while processing: ", feedUrl)
						log.Panic(err.Error())
					}
					for _, item := range feed.Channel.Items {
						fmt.Println(item.Title)
					}
					wg.Done()
				}(nextFeed.Url.String)
			}
			wg.Wait()
		}
	}
}

// func (config *apiConfig) processFeed(feedUrl string) {
// 	feed, err := rss.FetchFeed(feedUrl)
// 	if err != nil {
// 		log.Fatalf(err.Error())
// 	}
// 	for _, item := range feed.Channel.Items {
// 		fmt.Println(item.Title)
// 	}
// }

func main() {
	// const feedsToFetch = 10
	// err := godotenv.Load()
	// if err != nil {
	// 	log.Fatal(err)
	// }
	cfg, err := config.Read()
	if err != nil {
		log.Fatalln(err)
	}
	st := state{cfg: &cfg}
	handlers := map[string]func(*state, command) error{
		"login": handlerLogin,
	}
	cmds := commands{handlers: handlers}
	args := os.Args
	if len(args) < 2 {
		log.Fatal("No arguments provided")
	}
	cmd := command{name: args[1], arguments: args[2:]}
	err = cmds.run(&st, cmd)
	if err != nil {
		log.Fatalln(err)
	}
	// url := "https://blog.boot.dev/index.xml"
	// xml, err := rss.FetchFeed(url)
	// if err != nil {
	// 	log.Fatalf(err.Error())
	// }
	// fmt.Println(xml)
	// port := os.Getenv("PORT")
	// dbURL := os.Getenv("connection_string")
	// db, err := sql.Open("postgres", dbURL)
	// dbQueries := database.New(db)
	// config := apiConfig{DB: dbQueries}
	// config.workerFeedsFetch(feedsToFetch)
	// serveMux := http.NewServeMux()
	// serveMux.HandleFunc("GET /v1/healthz", handlerReadiness)
	// serveMux.HandleFunc("GET /v1/err", handlerErrResp)
	// serveMux.HandleFunc("POST /v1/users", createUserHandler(config))
	// serveMux.HandleFunc("GET /v1/users", config.middlewareAuth(config.getUserByApiKeyHandler))
	// serveMux.HandleFunc("POST /v1/feeds", config.middlewareAuth(config.createFeedHandler))
	// serveMux.HandleFunc("GET /v1/feeds", config.getFeedsHandler)
	// serveMux.HandleFunc("GET /v1/feed_follows", config.middlewareAuth(config.getFeedFollowsHandler))
	// serveMux.HandleFunc("POST /v1/feed_follows", config.middlewareAuth(config.createFeedFollowHandler))
	// serveMux.HandleFunc("DELETE /v1/feed_follows/{feedFollowID}", config.middlewareAuth(config.deleteFeedFollowHandler))
	// server := http.Server{Handler: serveMux, Addr: ":" + port}
	// server.ListenAndServe()
}
