package handlers

import (
	"AvitoPVZService/Service/internal/domain"
	"context"
	"encoding/json"
	"google.golang.org/protobuf/types/known/timestamppb"
	"testing"
	"time"
)

type fakeRepo struct{}

func (f *fakeRepo) Register(email, password, role string) error        { return nil }
func (f *fakeRepo) Login(email string) (error, *domain.User)           { return nil, nil }
func (f *fakeRepo) CreatePVZ(city, id string, regTime time.Time) error { return nil }
func (f *fakeRepo) CreateReception(PVZID, id string, dateTime time.Time) (error, *json.RawMessage) {
	return nil, nil
}
func (f *fakeRepo) AddProduct(id string, dateTime time.Time, prodType, PVZID string) (error, *json.RawMessage) {
	return nil, nil
}
func (f *fakeRepo) DeleteLastProduct(PVZID string) error                      { return nil }
func (f *fakeRepo) CloseLastReception(PVZID string, closedAt time.Time) error { return nil }
func (f *fakeRepo) ListPVZ(endStr, startStr string, limit, offset int) (error, *[]domain.PVZ) {
	return nil, nil
}
func (f *fakeRepo) GrpcListPVz(endPeriod, startPeriod string, limit, offset int) (error, []*ProtoPVZ) {
	return nil, []*ProtoPVZ{
		{Id: "1", City: "TestCity", RegistrationDate: timestamppb.Now()},
	}
}

func TestGetPVZList(t *testing.T) {
	repo := &fakeRepo{}
	g := NewGrpcHandlers(repo)
	resp, err := g.GetPVZList(context.Background(), &GetPVZListRequest{})
	if err != nil {
		t.Fatalf("GetPVZList error: %v", err)
	}
	if len(resp.PVZs) != 1 {
		t.Errorf("expected 1 PVZ, got %d", len(resp.PVZs))
	}
	if resp.PVZs[0].City != "TestCity" {
		t.Errorf("unexpected city: %s", resp.PVZs[0].City)
	}
}
