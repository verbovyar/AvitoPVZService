package handlers

import (
	"AvitoPVZService/Service/internal/domain"
	"AvitoPVZService/Service/internal/tokens"
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type memoryRepo struct {
	pvzs []domain.PVZ
}

func (m *memoryRepo) Register(email, password, role string) error {
	return nil
}
func (m *memoryRepo) Login(email string) (error, *domain.User) {
	return nil, &domain.User{ID: "test-user", PasswordHash: "pass", Role: "moderator"}
}
func (m *memoryRepo) CreatePVZ(city, id string, regTime time.Time) error {
	m.pvzs = append(m.pvzs, domain.PVZ{ID: id, City: city, RegistrationDate: regTime})
	return nil
}
func (m *memoryRepo) CreateReception(PVZID, id string, dateTime time.Time) (error, *json.RawMessage) {
	return errors.New("not implemented"), nil
}
func (m *memoryRepo) AddProduct(id string, dateTime time.Time, prodType, PVZID string) (error, *json.RawMessage) {
	return errors.New("not implemented"), nil
}
func (m *memoryRepo) DeleteLastProduct(PVZID string) error {
	return errors.New("not implemented")
}
func (m *memoryRepo) CloseLastReception(PVZID string, closedAt time.Time) error {
	return errors.New("not implemented")
}
func (m *memoryRepo) ListPVZ(endStr, startStr string, limit, offset int) (error, *[]domain.PVZ) {
	return nil, &m.pvzs
}
func (m *memoryRepo) GrpcListPVz(endPeriod, startPeriod string, limit, offset int) (error, []*ProtoPVZ) {
	return errors.New("not implemented"), nil
}

func TestCreateAndListPVZ(t *testing.T) {
	repo := &memoryRepo{}
	h := NewHttpHandlers(repo)
	server := httptest.NewServer(h)
	defer server.Close()

	token, err := tokens.CreateToken("user1", "moderator")
	if err != nil {
		t.Fatalf("failed to create token: %v", err)
	}

	createReq := map[string]string{"city": "Москва"}
	body, _ := json.Marshal(createReq)
	req, _ := http.NewRequest(http.MethodPost, server.URL+"/pvz", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("failed request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected status 201 Created, got %d", resp.StatusCode)
	}

	var created domain.PVZ
	if err := json.NewDecoder(resp.Body).Decode(&created); err != nil {
		t.Fatalf("failed decoding response: %v", err)
	}
	if created.City != "Москва" {
		t.Errorf("expected city 'Москва', got '%s'", created.City)
	}

	listURL := server.URL + "/pvz?limit=1&offset=0"
	req, _ = http.NewRequest(http.MethodGet, listURL, nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("failed request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200 OK, got %d", resp.StatusCode)
	}

	var list []domain.PVZ
	if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
		t.Fatalf("failed decoding list: %v", err)
	}
	if len(list) != 1 {
		t.Errorf("expected 1 pvz, got %d", len(list))
	}
}
