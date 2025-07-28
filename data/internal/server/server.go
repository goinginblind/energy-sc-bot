package server

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"strconv"

	"github.com/goinginblind/energy-sc-bot/data/internal/db"
	"github.com/gorilla/mux"
)

type Server struct {
	DB *db.Queries
}

func NewServer(database *sql.DB) *Server {
	return &Server{DB: db.New(database)}
}

func (s *Server) GetBillsByUserID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	bills, err := s.DB.GetBillsByUserID(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if bills == nil {
		bills = []db.Bill{}
	}

	json.NewEncoder(w).Encode(bills)
}

func (s *Server) GetUserByPhone(w http.ResponseWriter, r *http.Request) {
	phone := mux.Vars(r)["phone"]
	user, err := s.DB.GetUserByPhone(r.Context(), sql.NullString{String: phone, Valid: true})
	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(user)
}

func (s *Server) GetUserByEmail(w http.ResponseWriter, r *http.Request) {
	email := mux.Vars(r)["email"]
	user, err := s.DB.GetUserByEmail(r.Context(), sql.NullString{String: email, Valid: true})
	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(user)
}

func (s *Server) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req db.CreateUserParams
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	user, err := s.DB.CreateUser(r.Context(), req)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(user)
}

func (s *Server) UpsertUserByContact(w http.ResponseWriter, r *http.Request) {
	var req db.UpsertUserByContactParams
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	user, err := s.DB.UpsertUserByContact(r.Context(), req)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(user)
}

func (s *Server) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (s *Server) Router() http.Handler {
	r := mux.NewRouter()
	r.HandleFunc("/health", s.HealthCheck).Methods("GET")
	r.HandleFunc("/users/{id}/bills", s.GetBillsByUserID).Methods("GET")
	r.HandleFunc("/user/phone/{phone}", s.GetUserByPhone).Methods("GET")
	r.HandleFunc("/user/email/{email}", s.GetUserByEmail).Methods("GET")
	r.HandleFunc("/user", s.CreateUser).Methods("POST")
	r.HandleFunc("/user/upsert", s.UpsertUserByContact).Methods("POST")
	return r
}