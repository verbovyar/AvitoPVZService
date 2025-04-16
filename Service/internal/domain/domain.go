package domain

import "time"

type User struct {
	Id               string    `json:"id"`
	Email            string    `json:"email"`
	PasswordHash     string    `json:"-"`
	Role             string    `json:"role"`
	RegistrationDate time.Time `json:"registrationDate"`
}

type PVZ struct {
	Id               string    `json:"id"`
	RegistrationDate time.Time `json:"registrationDate"`
	City             string    `json:"city"`
}

type Reception struct {
	Id       string     `json:"id"`
	DateTime time.Time  `json:"dateTime"`
	PVZId    string     `json:"pvzId"`
	Status   string     `json:"status"`
	Products []*Product `json:"-"`
}

type Product struct {
	Id          string    `json:"id"`
	DateTime    time.Time `json:"dateTime"`
	Type        string    `json:"type"`
	ReceptionId string    `json:"receptionId"`
}
