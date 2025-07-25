package server

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

type Server struct {
	DB *sql.DB
}

func NewServer(db *sql.DB) *Server {
	return &Server{DB: db}
}

func (s *Server) GetUserByPhone(w http.ResponseWriter, r *http.Request) {
	phone := mux.Vars(r)["phone"]
	row := s.DB.QueryRow("SELECT telegram_id, phone, email FROM users WHERE phone = $1 LIMIT 1", phone)
	var user struct {
		TelegramID int64  `json:"telegram_id"`
		Phone      string `json:"phone"`
		Email      string `json:"email"`
	}
	if err := row.Scan(&user.TelegramID, &user.Phone, &user.Email); err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(user)
}

func (s *Server) GetUserByEmail(w http.ResponseWriter, r *http.Request) {
	email := mux.Vars(r)["email"]
	row := s.DB.QueryRow("SELECT telegram_id, phone, email FROM users WHERE email = $1 LIMIT 1", email)
	var user struct {
		TelegramID int64  `json:"telegram_id"`
		Phone      string `json:"phone"`
		Email      string `json:"email"`
	}
	if err := row.Scan(&user.TelegramID, &user.Phone, &user.Email); err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(user)
}

func (s *Server) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req struct {
		TelegramID int64  `json:"telegram_id"`
		Phone      string `json:"phone"`
		Email      string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	row := s.DB.QueryRow("INSERT INTO users (telegram_id, phone, email) VALUES ($1, $2, $3) RETURNING telegram_id, phone, email", req.TelegramID, req.Phone, req.Email)
	var user struct {
		TelegramID int64  `json:"telegram_id"`
		Phone      string `json:"phone"`
		Email      string `json:"email"`
	}
	if err := row.Scan(&user.TelegramID, &user.Phone, &user.Email); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(user)
}

func (s *Server) UpsertUserByContact(w http.ResponseWriter, r *http.Request) {
	var req struct {
		TelegramID int64  `json:"telegram_id"`
		Phone      string `json:"phone"`
		Email      string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	row := s.DB.QueryRow(`INSERT INTO users (telegram_id, phone, email) VALUES ($1, $2, $3)
		ON CONFLICT (phone, email) DO UPDATE
		SET telegram_id = COALESCE(EXCLUDED.telegram_id, users.telegram_id),
		    phone = COALESCE(EXCLUDED.phone, users.phone),
		    email = COALESCE(EXCLUDED.email, users.email)
		RETURNING telegram_id, phone, email`, req.TelegramID, req.Phone, req.Email)
	var user struct {
		TelegramID int64  `json:"telegram_id"`
		Phone      string `json:"phone"`
		Email      string `json:"email"`
	}
	if err := row.Scan(&user.TelegramID, &user.Phone, &user.Email); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(user)
}

func (s *Server) Router() http.Handler {
	r := mux.NewRouter()
	r.HandleFunc("/user/phone/{phone}", s.GetUserByPhone).Methods("GET")
	r.HandleFunc("/user/email/{email}", s.GetUserByEmail).Methods("GET")
	r.HandleFunc("/user", s.CreateUser).Methods("POST")
	r.HandleFunc("/user/upsert", s.UpsertUserByContact).Methods("POST")
	return r
}
