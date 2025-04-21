package handlers

import (
	"AvitoPVZService/Service/internal/domain"
	"AvitoPVZService/Service/internal/repositories/interfaces"
	"AvitoPVZService/Service/internal/tokens"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type HttpHandlers struct {
	Data interfaces.Repository
}

func NewHttpHandlers(repo interfaces.Repository) *HttpHandlers {
	return &HttpHandlers{Data: repo}
}

func (h *HttpHandlers) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	switch {
	case path == "/dummyLogin" && r.Method == http.MethodPost:
		h.DummyLogin(w, r)

	case path == "/register" && r.Method == http.MethodPost:
		h.Register(w, r)

	case path == "/login" && r.Method == http.MethodPost:
		h.Login(w, r)

	case path == "/pvz":
		if r.Method == http.MethodPost {
			tokens.AuthMiddleware(h.CreatePVZ, "moderator")(w, r)
		} else if r.Method == http.MethodGet {
			tokens.AuthMiddleware(h.ListPVZ, "moderator", "employee")(w, r)
		} else {
			http.Error(w, `{"message":"method not allowed"}`, http.StatusMethodNotAllowed)
		}

	case path == "/receptions" && r.Method == http.MethodPost:
		tokens.AuthMiddleware(h.CreateReception, "employee")(w, r)

	case path == "/products" && r.Method == http.MethodPost:
		tokens.AuthMiddleware(h.AddProduct, "employee")(w, r)

	case strings.HasPrefix(path, "/pvz/"):
		if strings.HasSuffix(path, "/delete_last_product") {
			tokens.AuthMiddleware(h.DeleteLastProduct, "employee")(w, r)
		} else if strings.HasSuffix(path, "/close_last_reception") {
			tokens.AuthMiddleware(h.CloseLastReception, "employee")(w, r)
		} else {
			http.NotFound(w, r)
		}

	default:
		http.NotFound(w, r)
	}
}

// ----------

type dummyLoginReq struct {
	Role string `json:"role"`
}

func (h *HttpHandlers) DummyLogin(w http.ResponseWriter, r *http.Request) {
	var req dummyLoginReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || (req.Role != "employee" && req.Role != "moderator") {
		http.Error(w, `{"message":"bad request"}`, http.StatusBadRequest)
		return
	}
	token, err := tokens.CreateToken(uuid.NewString(), req.Role)
	if err != nil {
		http.Error(w, `{"message":"token error"}`, http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(token)
}

// ----------

type registerReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

func (h *HttpHandlers) Register(w http.ResponseWriter, r *http.Request) {
	var req registerReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || (req.Role != "employee" && req.Role != "moderator") {
		http.Error(w, `{"message":"bad request"}`, http.StatusBadRequest)
		return
	}

	err := h.Data.Register(req.Email, req.Password, req.Role)
	if err != nil {
		http.Error(w, `{"message":"cannot create user"}`, http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"email": req.Email, "role": req.Role})
}

// ----------

type loginReq struct {
	Email    string
	Password string
}

func (h *HttpHandlers) Login(w http.ResponseWriter, r *http.Request) {
	var req loginReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"message":"bad request"}`, http.StatusBadRequest)
		return
	}

	err, user := h.Data.Login(req.Email)
	if err != nil {
		http.Error(w, `{"message":"invalid credentials"}`, http.StatusUnauthorized)
		return
	}
	if user.PasswordHash != req.Password {
		http.Error(w, `{"message":"invalid credentials"}`, http.StatusUnauthorized)
		return
	}
	token, _ := tokens.CreateToken(user.ID, user.Role)
	json.NewEncoder(w).Encode(token)
}

// ----------

type createPVZReq struct {
	City string `json:"city"`
}

func (h *HttpHandlers) CreatePVZ(w http.ResponseWriter, r *http.Request) {
	var req createPVZReq
	json.NewDecoder(r.Body).Decode(&req)
	var allowedCities = map[string]bool{"Москва": true, "Санкт-Петербург": true, "Казань": true}
	if !allowedCities[req.City] {
		http.Error(w, `{"message":"bad request"}`, http.StatusBadRequest)
		return
	}

	id := uuid.NewString()
	regTime := time.Now()
	err := h.Data.CreatePVZ(req.City, id, regTime)

	if err != nil {
		http.Error(w, `{"message":"cannot create pvz"}`, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(domain.PVZ{ID: id, City: req.City, RegistrationDate: regTime})
}

// ----------

type RecWithProds struct {
	Reception domain.Reception `json:"reception"`
	Products  []domain.Product `json:"products"`
}
type PVZWithRecs struct {
	PVZ        domain.PVZ     `json:"pvz"`
	Receptions []RecWithProds `json:"receptions"`
}

func (h *HttpHandlers) ListPVZ(w http.ResponseWriter, r *http.Request) {
	startStr := r.URL.Query().Get("startDate")
	endStr := r.URL.Query().Get("endDate")
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	if limitStr == "" {
		limitStr = "10"
	}
	if offsetStr == "" {
		offsetStr = "0"
	}
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 10
	}
	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}
	if startStr == "" {
		startStr = time.Now().Add(-24 * time.Hour).Format(time.RFC3339)
	}
	if endStr == "" {
		endStr = time.Now().Format(time.RFC3339)
	}

	err, result := h.Data.ListPVZ(endStr, startStr, limit, offset)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"message":"db error: %v"}`, err), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(result)
}

// ----------

type createReceptionReq struct {
	PVZID string `json:"pvzId"`
}

func (h *HttpHandlers) CreateReception(w http.ResponseWriter, r *http.Request) {
	var req createReceptionReq
	json.NewDecoder(r.Body).Decode(&req)

	id := uuid.NewString()
	dateTime := time.Now()
	err, lastElem := h.Data.CreateReception(req.PVZID, id, dateTime)
	if err != nil {
		http.Error(w, `{"message":"cannot create reception"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(&lastElem)
}

// ----------

type addProductReq struct {
	Type  string `json:"type"`
	PVZID string `json:"pvzId"`
}

func (h *HttpHandlers) AddProduct(w http.ResponseWriter, r *http.Request) {
	var req addProductReq
	json.NewDecoder(r.Body).Decode(&req)
	var allowedTypes = map[string]bool{"электроника": true, "одежда": true, "обувь": true}
	if !allowedTypes[req.Type] {
		http.Error(w, `{"message":"bad type"}`, http.StatusBadRequest)
		return
	}

	id := uuid.NewString()
	dateTime := time.Now()
	err, lastElem := h.Data.AddProduct(id, dateTime, req.Type, req.PVZID)
	if err != nil {
		http.Error(w, `{"message":"cannot add product"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(&lastElem)
}

// ----------

func (h *HttpHandlers) DeleteLastProduct(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	pvzId := parts[2]

	err := h.Data.DeleteLastProduct(pvzId)

	if err != nil {
		http.Error(w, `{"message":"cannot delete"}`, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// ----------

func (h *HttpHandlers) CloseLastReception(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	pvzId := parts[2]

	err := h.Data.CloseLastReception(pvzId, time.Now())
	if err != nil {
		http.Error(w, `{"message":"cannot delete"}`, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
