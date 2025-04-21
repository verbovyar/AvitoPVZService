package db_test

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	pgconn "github.com/jackc/pgconn"
	pgproto32 "github.com/jackc/pgproto3/v2"
	"github.com/jackc/pgx/v4"

	"AvitoPVZService/Service/internal/domain"
	db "AvitoPVZService/Service/internal/repositories/db"
)

type mockPool struct {
	ExecFunc     func(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error)
	QueryRowFunc func(ctx context.Context, sql string, args ...interface{}) pgx.Row
	QueryFunc    func(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
}

func (m *mockPool) Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error) {
	return m.ExecFunc(ctx, sql, args...)
}
func (m *mockPool) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	return m.QueryRowFunc(ctx, sql, args...)
}
func (m *mockPool) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	return m.QueryFunc(ctx, sql, args...)
}

type mockRow struct {
	scanFunc func(dest ...interface{}) error
}

func (r *mockRow) Scan(dest ...interface{}) error {
	return r.scanFunc(dest...)
}

type mockRows struct {
	rows [][]interface{}
	idx  int
}

func (m *mockRows) CommandTag() pgconn.CommandTag {
	//TODO implement me
	panic("implement me")
}

func (m *mockRows) Values() ([]interface{}, error) {
	//TODO implement me
	panic("implement me")
}

func (m *mockRows) RawValues() [][]byte {
	//TODO implement me
	panic("implement me")
}

func (m *mockRows) Next() bool {
	m.idx++
	return m.idx <= len(m.rows)
}
func (m *mockRows) Scan(dest ...interface{}) error {
	row := m.rows[m.idx-1]
	for i, d := range dest {
		switch dd := d.(type) {
		case *string:
			*dd = row[i].(string)
		case *time.Time:
			*dd = row[i].(time.Time)
		default:
			return fmt.Errorf("unsupported scan type %T", dd)
		}
	}
	return nil
}
func (m *mockRows) Err() error                                      { return nil }
func (m *mockRows) Close()                                          {}
func (m *mockRows) FieldDescriptions() []pgproto32.FieldDescription { return nil }

func TestRegister_Success(t *testing.T) {
	called := false
	mock := &mockPool{
		ExecFunc: func(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error) {
			called = true
			if !strings.Contains(sql, "INSERT INTO avito_schema.users") {
				t.Errorf("unexpected SQL: %s", sql)
			}
			if len(args) < 5 {
				t.Errorf("expected at least 5 args, got %d", len(args))
			}
			return pgconn.CommandTag("INSERT"), nil
		},
	}
	repo := db.New(mock)
	if err := repo.Register("email@example.com", "pass", "role"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !called {
		t.Error("Exec was not called")
	}
}

func TestLogin_Success(t *testing.T) {
	expID, expHash, expRole := "uid", "hash", "role"
	mock := &mockPool{
		QueryRowFunc: func(ctx context.Context, sql string, args ...interface{}) pgx.Row {
			return &mockRow{scanFunc: func(dest ...interface{}) error {
				*(dest[0].(*string)) = expID
				*(dest[1].(*string)) = expHash
				*(dest[2].(*string)) = expRole
				return nil
			}}
		},
	}
	repo := db.New(mock)
	err, user := repo.Login("email@example.com")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if user.ID != expID || user.PasswordHash != expHash || user.Role != expRole {
		t.Errorf("unexpected user: %+v", user)
	}
}

func TestCreatePVZ_Success(t *testing.T) {
	called := false
	mock := &mockPool{
		ExecFunc: func(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error) {
			called = true
			if !strings.Contains(sql, "INSERT INTO avito_schema.pvz") {
				t.Errorf("unexpected SQL: %s", sql)
			}
			return pgconn.CommandTag("INSERT"), nil
		},
	}
	repo := db.New(mock)
	if err := repo.CreatePVZ("City", "pvz-id", time.Now()); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !called {
		t.Error("Exec was not called")
	}
}

func TestDeleteLastProduct_Success(t *testing.T) {
	mock := &mockPool{
		QueryRowFunc: func(ctx context.Context, sql string, args ...interface{}) pgx.Row {
			return &mockRow{scanFunc: func(dest ...interface{}) error {
				*(dest[0].(*string)) = "deleted-id"
				return nil
			}}
		},
	}
	repo := db.New(mock)
	if err := repo.DeleteLastProduct("pvz1"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestCloseLastReception_Success(t *testing.T) {
	mock := &mockPool{
		QueryRowFunc: func(ctx context.Context, sql string, args ...interface{}) pgx.Row {
			return &mockRow{scanFunc: func(dest ...interface{}) error {
				*(dest[0].(*string)) = "pvz1"
				return nil
			}}
		},
	}
	repo := db.New(mock)
	if err := repo.CloseLastReception("pvz1", time.Now()); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestListPVZ_Success(t *testing.T) {
	t1 := time.Now().Add(-time.Hour)
	t2 := time.Now()
	mockRows := &mockRows{rows: [][]interface{}{{"id1", "city1", t1}, {"id2", "city2", t2}}}
	mock := &mockPool{
		QueryFunc: func(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
			return mockRows, nil
		},
	}
	repo := db.New(mock)
	err, list := repo.ListPVZ("end", "start", 2, 0)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	expected := []domain.PVZ{{ID: "id1", City: "city1", RegistrationDate: t1}, {ID: "id2", City: "city2", RegistrationDate: t2}}
	if !reflect.DeepEqual(*list, expected) {
		t.Errorf("expected %+v, got %+v", expected, *list)
	}
}

func TestGrpcListPVz_Success(t *testing.T) {
	t1 := time.Now().Add(-time.Hour)
	mockRows := &mockRows{rows: [][]interface{}{{"id1", t1, "city1"}}}
	mock := &mockPool{
		QueryFunc: func(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
			return mockRows, nil
		},
	}
	repo := db.New(mock)
	err, resp := repo.GrpcListPVz("end", "start", 1, 0)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(resp) != 1 {
		t.Fatalf("expected 1 result, got %d", len(resp))
	}
	pvz := resp[0]
	if pvz.Id != "id1" || pvz.City != "city1" || !pvz.RegistrationDate.AsTime().Equal(t1) {
		t.Errorf("unexpected ProtoPVZ: %+v", pvz)
	}
}

func TestRegister_Error(t *testing.T) {
	want := errors.New("fail exec")
	mock := &mockPool{
		ExecFunc: func(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error) {
			return pgconn.CommandTag(""), want
		},
	}
	repo := db.New(mock)
	if err := repo.Register("em", "pw", "rl"); err != want {
		t.Errorf("expected %v, got %v", want, err)
	}
}

func TestLogin_NotFound(t *testing.T) {
	want := errors.New("no user")
	mock := &mockPool{
		QueryRowFunc: func(ctx context.Context, sql string, args ...interface{}) pgx.Row {
			return &mockRow{scanFunc: func(dest ...interface{}) error {
				return want
			}}
		},
	}
	repo := db.New(mock)
	err, _ := repo.Login("em")
	if err != want {
		t.Errorf("expected %v, got %v", want, err)
	}
}
