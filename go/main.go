package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hashicorp/go-uuid"
)

type todoitem struct {
	Title     string `json:"title"`
	Completed bool   `json:"completed"`
	Url       string `json:"url"`
	Uuid      string `json:"uuid"`
	Order     int    `json:"order"`
}

type server struct {
	router *http.ServeMux
	todos  []todoitem
}

func (s *server) todoIndividualHandle() http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, PATCH, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		if r.Method == "OPTIONS" {
			return
		}
		if r.Method == "GET" {
			for i := range s.todos {
				if s.todos[i].Uuid == r.RequestURI[1:] {
					w.Header().Set("Content-Type", "Application/JSON")
					json.NewEncoder(w).Encode(s.todos[i])
					return
				}
			}
		}

		if r.Method == "PATCH" {
			for i := range s.todos {
				if s.todos[i].Uuid == r.RequestURI[1:] {
					var input todoitem
					err := json.NewDecoder(r.Body).Decode(&input)
					if err != nil {
						http.Error(w, err.Error(), http.StatusBadRequest)
						return
					}
					s.todos[i].Title = input.Title
					s.todos[i].Completed = input.Completed
					s.todos[i].Order = input.Order
					json.NewEncoder(w).Encode(s.todos[i])
				}
			}
		}

		if r.Method == "DELETE" {
			for i := range s.todos {
				if s.todos[i].Uuid == r.RequestURI[1:] {
					s.todos = append(s.todos[:i], s.todos[i+1:]...)
					fmt.Fprint(w, "{}")
				}
			}
		}
	}

	return http.HandlerFunc(fn)
}

func (s *server) todoHandle() http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, PATCH, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		if r.Method == "OPTIONS" {
			return // Preflight sets headers and we're done
		}

		if r.Method == "POST" {
			var input todoitem
			err := json.NewDecoder(r.Body).Decode(&input)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			url, err := uuid.GenerateUUID()

			input.Uuid = url

			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			input.Url = fmt.Sprintf("http://localhost:4000/%s", url)

			s.todos = append(s.todos, input)

			json.NewEncoder(w).Encode(input)

			s.router.Handle("/"+url, s.todoIndividualHandle())
		}

		if r.Method == "DELETE" {
			s.todos = make([]todoitem, 0)
			fmt.Fprint(w, "{}")
		}

		if r.Method == "GET" {
			w.Header().Set("Content-Type", "Application/JSON")
			json.NewEncoder(w).Encode(s.todos)
		}

	}

	return http.HandlerFunc(fn)
}

func main() {
	srv := server{
		router: http.NewServeMux(),
		todos:  make([]todoitem, 0),
	}

	srv.router.Handle("/", srv.todoHandle())
	http.ListenAndServe(":4000", srv.router)
}
