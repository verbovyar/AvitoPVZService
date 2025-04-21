package db

import (
	"AvitoPVZService/Service/internal/domain"
	"AvitoPVZService/Service/internal/handlers"
	"AvitoPVZService/Service/internal/repositories/interfaces"
	"context"
	"encoding/json"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
	"log"
	"sync"
	"time"
)

const poolSize = 5

type PostgresRepository struct {
	Pool interfaces.PgxPoolIface

	mu          sync.RWMutex
	poolChannel chan struct{}
}

func New(pool interfaces.PgxPoolIface) *PostgresRepository {
	return &PostgresRepository{
		Pool:        pool,
		mu:          sync.RWMutex{},
		poolChannel: make(chan struct{}, poolSize),
	}
}

func (r *PostgresRepository) Register(email, password, role string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	id := uuid.NewString()
	const query = `INSERT INTO avito_schema.users(id,email,password_hash,role,registration_date) VALUES($1,$2,$3,$4,$5)`
	_, err := r.Pool.Exec(context.Background(), query, id, email, password, role, time.Now())

	return err
}

func (r *PostgresRepository) Login(email string) (error, *domain.User) {
	r.mu.Lock()
	defer r.mu.Unlock()

	var user domain.User
	const query = `SELECT id,password_hash,role FROM avito_schema.users WHERE email=$1`
	row := r.Pool.QueryRow(context.Background(), query, email)
	err := row.Scan(&user.ID, &user.PasswordHash, &user.Role)

	return err, &user
}

func (r *PostgresRepository) CreatePVZ(city, id string, regTime time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	const query = `INSERT INTO avito_schema.pvz(id, city, registration_date, is_reception_open, receptions) VALUES($1,$2,$3,$4,$5)`
	_, err := r.Pool.Exec(context.Background(), query, id, city, regTime, false, "[]")

	return err
}

type receptionJson struct {
	ID       string        `json:"id"`
	PvzID    string        `json:"pvz_id"`
	OpenAt   time.Time     `json:"open_at"`
	ClosedAt string        `json:"closed_at"`
	Products []interface{} `json:"products"`
}

func (r *PostgresRepository) CreateReception(PVZID string, id string, dateTime time.Time) (error, *json.RawMessage) {
	r.mu.Lock()
	defer r.mu.Unlock()

	rec := receptionJson{
		ID:       id,
		PvzID:    PVZID,
		OpenAt:   dateTime,
		ClosedAt: "",
		Products: []interface{}{},
	}
	recJSON, err := json.Marshal(rec)
	if err != nil {
		log.Fatal(err)
	}

	const query = `
	UPDATE avito_schema.pvz
	SET is_reception_open = $1,
		receptions = receptions || $2::jsonb
	WHERE id = $3
	  AND is_reception_open = false
	RETURNING receptions->-1;
	`

	var lastElem json.RawMessage
	err = r.Pool.QueryRow(context.Background(), query, true, recJSON, PVZID).Scan(&lastElem)
	if err != nil {
		log.Fatal("Update failed or no rows affected:", err)
	}

	return err, &lastElem
}

type productJson struct {
	ID          string    `json:"id"`
	DateTime    time.Time `json:"dateTime"`
	Type        string    `json:"type"`
	ReceptionID string    `json:"receptionId"`
}

func (r *PostgresRepository) AddProduct(id string, dateTime time.Time, prodType, PVZID string) (error, *json.RawMessage) {
	r.mu.Lock()
	defer r.mu.Unlock()

	prod := productJson{
		ID:          id,
		DateTime:    dateTime,
		Type:        prodType,
		ReceptionID: "",
	}
	prodJSON, err := json.Marshal(prod)
	if err != nil {
		log.Fatal(err)
	}

	const query = `
	UPDATE avito_schema.pvz
	SET receptions = jsonb_insert(
		receptions,
		'{-1,products,-1}',
		$1::jsonb,
		true
	)
	WHERE id = $2
	  AND is_reception_open = true
	RETURNING
	  receptions->-1->>'id'         AS reception_id,
	  receptions->-1->'products'->-1 AS last_product;
	`

	var lastProduct json.RawMessage
	var receptionID string
	err = r.Pool.QueryRow(context.Background(), query, prodJSON, PVZID).Scan(&receptionID, &lastProduct)
	if err != nil {
		log.Fatalf("update failed or no rows returned: %v", err)
	}

	var p productJson
	if err := json.Unmarshal(lastProduct, &p); err != nil {
		return err, nil
	}
	p.ReceptionID = receptionID

	updatedJSON, err := json.Marshal(p)
	if err != nil {
		return err, nil
	}
	updatedRaw := json.RawMessage(updatedJSON)

	return err, &updatedRaw
}

func (r *PostgresRepository) DeleteLastProduct(PVZID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	const query = `
	UPDATE avito_schema.pvz
	SET receptions = jsonb_set(
		receptions,
		'{-1,products}',
		(receptions->-1->'products') - (-1),
		false
	)
	WHERE id = $1
	  AND is_reception_open = $2
	  AND jsonb_array_length(receptions->-1->'products') != 0
	RETURNING id;
	`

	var updatedID string
	err := r.Pool.QueryRow(context.Background(), query, PVZID, true).Scan(&updatedID)

	return err
}

func (r *PostgresRepository) CloseLastReception(PVZID string, closedAt time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	const query = `
	UPDATE avito_schema.pvz
	SET
		is_reception_open = $1,
		receptions = jsonb_set(
			receptions,
			'{-1,closed_at}',
			to_jsonb($2::text),
			false
		)
	WHERE id = $3
	  AND is_reception_open = $4
	RETURNING id;
`

	closedAtStr := closedAt.Format("2006-01-02 15:04:05.999999")

	var updatedID string
	err := r.Pool.QueryRow(context.Background(), query, false, closedAtStr, PVZID, true).Scan(&updatedID)

	return err
}

func (r *PostgresRepository) ListPVZ(endStr, startStr string, limit, offset int) (error, *[]domain.PVZ) {
	r.mu.Lock()
	defer r.mu.Unlock()

	const sqlQuery = `
	SELECT id, city, registration_date
	  FROM avito_schema.pvz
	 WHERE EXISTS (
		 SELECT 1
		   FROM jsonb_array_elements(receptions) AS reception
		  WHERE (reception->>'open_at') <= $1
			AND (reception->>'closed_at') >= $2
	 )
	 LIMIT $3 OFFSET $4;
`
	rows, err := r.Pool.Query(context.Background(), sqlQuery, endStr, startStr, limit, offset)

	defer rows.Close()

	var result []domain.PVZ
	for rows.Next() {
		var p domain.PVZ
		err = rows.Scan(&p.ID, &p.City, &p.RegistrationDate)
		if err != nil {
			continue
		}
		result = append(result, p)
	}

	return err, &result
}

func (r *PostgresRepository) GrpcListPVz(endPeriod, startPeriod string, limit, offset int) (error, []*handlers.ProtoPVZ) {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	const sqlQuery = `
	SELECT id, registration_date, city
	  FROM avito_schema.pvz
	 WHERE EXISTS (
		 SELECT 1
		   FROM jsonb_array_elements(receptions) AS reception
		  WHERE (reception->>'open_at') <= $1
			AND (reception->>'closed_at') >= $2
	 )
	 LIMIT $3 OFFSET $4;
`

	rows, err := r.Pool.Query(context.Background(), sqlQuery, endPeriod, startPeriod, limit, offset)

	defer rows.Close()

	var resp []*handlers.ProtoPVZ
	for rows.Next() {
		var id, city string
		var dt time.Time
		if err := rows.Scan(&id, &dt, &city); err != nil {
			continue
		}
		resp = append(resp, &handlers.ProtoPVZ{
			Id:               id,
			RegistrationDate: timestamppb.New(dt),
			City:             city,
		})
	}

	return err, resp
}
