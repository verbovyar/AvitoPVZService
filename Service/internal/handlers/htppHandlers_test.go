package handlers

import (
	"AvitoPVZService/Service/internal/domain"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type fakeRepoHTTP struct{}

func (f *fakeRepoHTTP) Login(email string) (error, *domain.User) {
	//TODO implement me
	panic("implement me")
}

func (f *fakeRepoHTTP) CreatePVZ(city, id string, regTime time.Time) error {
	//TODO implement me
	panic("implement me")
}

func (f *fakeRepoHTTP) CreateReception(PVZID string, id string, dateTime time.Time) (error, *json.RawMessage) {
	//TODO implement me
	panic("implement me")
}

func (f *fakeRepoHTTP) AddProduct(id string, dateTime time.Time, prodType, PVZID string) (error, *json.RawMessage) {
	//TODO implement me
	panic("implement me")
}

func (f *fakeRepoHTTP) DeleteLastProduct(PVZID string) error {
	//TODO implement me
	panic("implement me")
}

func (f *fakeRepoHTTP) CloseLastReception(PVZID string, closedAt time.Time) error {
	//TODO implement me
	panic("implement me")
}

func (f *fakeRepoHTTP) ListPVZ(endStr, startStr string, limit, offset int) (error, *[]domain.PVZ) {
	//TODO implement me
	panic("implement me")
}

func (f *fakeRepoHTTP) GrpcListPVz(endPeriod, startPeriod string, limit, offset int) (error, []*ProtoPVZ) {
	//TODO implement me
	panic("implement me")
}

func (f *fakeRepoHTTP) Register(email, password, role string) error {
	if email == "" || role != "employee" && role != "moderator" {
		return fmt.Errorf("bad data")
	}
	return nil
}

func TestDummyLogin_BadRequest(t *testing.T) {
	handler := NewHttpHandlers(&fakeRepoHTTP{})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/dummyLogin", bytes.NewBufferString(`{"role":"bad"}`))
	handler.DummyLogin(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestDummyLogin_Success(t *testing.T) {
	handler := NewHttpHandlers(&fakeRepoHTTP{})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/dummyLogin", bytes.NewBufferString(`{"role":"employee"}`))
	handler.DummyLogin(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var token string
	if err := json.NewDecoder(rec.Body).Decode(&token); err != nil {
		t.Fatal("response is not a token string")
	}
	if token == "" {
		t.Error("empty token")
	}
}
