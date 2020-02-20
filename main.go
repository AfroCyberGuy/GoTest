package main

import (
	"github.com/gorilla/mux"
	"net/http"
	"fmt"
	"log"
	"time"
	
)

type server struct {
	router *mux.Router
	LastSeen map[string]time.Time
	AccountsCreated map[string]int32
	FailedTransactions map[string]int32
}

func (s *server) routes() {
	var h http.Handler
	h = s.handleExample()
	h = lastSeenMiddleware(s, h)
	h = failedTransactionsMiddleware(s, h)
	h = accountsCreatedMiddleware(s,h)
	s.router.Handle("/api/{username}", h)
}

func (s *server) handleExample() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("in request handler")
		w.Write([]byte("Hello world"))
	}
}

func GetIP(r *http.Request) string {
	forwarded := r.Header.Get("X-FORWARDED-FOR")
	if forwarded != "" {
		return forwarded
	}
	return r.RemoteAddr
}

func lastSeenMiddleware(s *server, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	  now := time.Now()
	  params := mux.Vars(r)
	  username := params["username"]
	  lastSeen, ok := s.LastSeen[username]
	  s.LastSeen[username] = now
	  if ok {
		  if now.Sub(lastSeen) < time.Second {
			fmt.Println("Too many requests in one second")
			return
		  }
	  }
	  next.ServeHTTP(w, r)
	})
}

func accountsCreatedMiddleware(s *server, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	  ipAddress := GetIP(r)
	  if s.AccountsCreated[ipAddress] >= 20 {
		  fmt.Println("exceeded accounts created")
		  return
	  }
	  s.AccountsCreated[ipAddress] += 1
	  next.ServeHTTP(w, r)
	})
}

func failedTransactionsMiddleware(s *server, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	  params := mux.Vars(r)
	  username := params["username"]
	  if s.FailedTransactions[username] >= 20 {
		return
	  }
	  s.FailedTransactions[username] += 1
	  next.ServeHTTP(w, r)
	})
}

func main() {
	
	router := mux.NewRouter()

	srv := server{
		router,
		make(map[string]time.Time),
		make(map[string]int32),
		make(map[string]int32),
	}
	srv.routes()

	srv2 := &http.Server{
		Handler: srv.router,
		Addr:    "127.0.0.1:8000",
	}

	log.Fatal(srv2.ListenAndServe())
}
