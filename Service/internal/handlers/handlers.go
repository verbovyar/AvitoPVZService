package handlers

import (
	"AvitoPVZService/Service/internal/domain"
	"AvitoPVZService/Service/internal/repositories/interfaces"
	"AvitoPVZService/Service/internal/tokens"
	"encoding/json"
	"github.com/google/uuid"
	"net/http"
	"time"
)

type Handlers struct {
	Data interfaces.Repository
}

func New(repo interfaces.Repository) *Handlers {
	return &Handlers{Data: repo}
}

type dummyReq struct {
	Role string `json:"role"`
}

func (h *Handlers) DummyLogin(w http.ResponseWriter, r *http.Request) {
	var body dummyReq
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || (body.Role != "employee" && body.Role != "moderator") {
		http.Error(w, `{"message": "Invalid request"}`, http.StatusBadRequest)
		return
	}
	userID := uuid.New().String()
	token, err := tokens.CreateToken(userID, body.Role)
	if err != nil {
		http.Error(w, `{"message": "Failed to create token"}`, http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(token)
}

// -----------------
type registerReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

func (h *Handlers) Register(w http.ResponseWriter, r *http.Request) {
	var body registerReq
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Email == "" || body.Password == "" || (body.Role != "employee" && body.Role != "moderator") {
		http.Error(w, `{"message": "Invalid request"}`, http.StatusBadRequest)
		return
	}

	// TODO Вызов к базе, поиск уже существующего юзера

	userID := uuid.New().String()
	user := &domain.User{
		Id:               userID,
		Email:            body.Email,
		PasswordHash:     body.Password,
		Role:             body.Role,
		RegistrationDate: time.Now(),
	}

	// TODO Вставить нашего нового юзера в БД

	w.WriteHeader(http.StatusCreated)
	resp := map[string]interface{}{
		"id":    user.Id,
		"email": user.Email,
		"role":  user.Role,
	}
	json.NewEncoder(w).Encode(resp)
}

//-----------------
