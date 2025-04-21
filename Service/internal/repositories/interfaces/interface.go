package interfaces

import (
	"AvitoPVZService/Service/internal/domain"
	"encoding/json"
	"time"
)

type Repository interface {
	Register(email, password, role string) error
	Login(email string) (error, *domain.User)
	CreatePVZ(city, id string, regTime time.Time) error
	CreateReception(PVZID string, id string, dateTime time.Time) (error, *json.RawMessage)
	AddProduct(id string, dateTime time.Time, prodType, PVZID string) (error, *json.RawMessage)
	DeleteLastProduct(PVZID string) error
	CloseLastReception(PVZID string, closedAt time.Time) error
	ListPVZ(endStr, startStr string, limit, offset int) (error, *[]domain.PVZ)
}
