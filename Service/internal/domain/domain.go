package domain

import "time"

type User struct {
	ID               string    `json:"id"`
	Email            string    `json:"email"`
	PasswordHash     string    `json:"-"`
	Role             string    `json:"role"`
	RegistrationDate time.Time `json:"registrationDate"`
}

type PVZ struct {
	ID               string    `json:"id"`
	City             string    `json:"city"`
	RegistrationDate time.Time `json:"registrationDate"`
}

type Reception struct {
	ID       string    `json:"id"`
	PVZID    string    `json:"pvzId"`
	DateTime time.Time `json:"dateTime"`
	Status   string    `json:"status"`
}

type Product struct {
	ID          string    `json:"id"`
	ReceptionID string    `json:"receptionId"`
	Type        string    `json:"type"`
	DateTime    time.Time `json:"dateTime"`
}
