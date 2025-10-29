package db

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"log"
	"net/url"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/ory/dockertest/v3"
)

type User struct {
	ID         int64
	UUID       uuid.UUID  `db:"uuid"`
	Name       string     `db:"name"`
	Email      string     `db:"email"`
	Address    Address    `db:"address"`
	Properties Properties `db:"properties"`
}

type Address struct {
	Street string `db:"street"`
	Zip    string `db:"zip"`
	City   string `db:"city"`
	State  string `db:"state"`
}

type Properties struct {
	Married     bool        `db:"married"`
	Preferences Preferences `db:"preferences"`
}

type Preferences struct {
	DarkMode bool   `db:"dark_mode"`
	Language string `db:"language"`
	TimeZone string `db:"timezone"`
}

func TestDBTestSuite(t *testing.T) {
	suite.Run(t, new(DBTestSuite))
}

type DBTestSuite struct {
	suite.Suite
	db      *DB
	connStr string
	cleanup func()
}

func (s *DBTestSuite) SetupSuite() {
	connStr, cleanup, err := setupPostgres()
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	d, err := New(ctx, ConnectionConfig{})
	if err != nil {
		panic(err)
	}

	s.cleanup = cleanup
	s.db = d
}

func (s *DBTestSuite) TearDownSuite() {
	s.cleanup()
	if s.db != nil {
		s.db.Close()
	}

}

func (s *DBTestSuite) SetupTest() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := s.db.Pool().Exec(ctx, `
		DROP TABLE IF EXISTS users;
		CREATE TABLE users (
		  	id bigserial PRIMARY KEY,
			uuid UUID NOT NULL UNIQUE,
			name text NOT NULL,
			email text NOT NULL UNIQUE,
			address JSONB NOT NULL,
			properties JSONB NOT NULL
	);`)
	if err != nil {
		s.T().Fatal(err)
	}

	err = s.seed()
	if err != nil {
		s.T().Fatal(err)
	}
}

func (s *DBTestSuite) TestQueries_Integration() {
	var usr1 []User
	err := Query[User](s.T().Context(), s.db.Pool(), "SELECT uuid, id, name, email, address, properties FROM users WHERE name = @name AND email = @email", &usr1, pgx.NamedArgs{"name": "Alice", "email": "alice@example.com"})
	s.Require().NoError(err)
	s.Require().NotEmpty(usr1[0].UUID.String())
	s.Require().Equal("Alice", usr1[0].Name)
	s.Require().Equal(true, usr1[0].Properties.Married)
	s.Require().Equal(true, usr1[0].Properties.Preferences.DarkMode)
	s.Require().Equal("en", usr1[0].Properties.Preferences.Language)

	// Test QueryOne
	var usr2 User
	err = QueryOne[User](s.T().Context(), s.db.Pool(), "SELECT uuid, id, name, email, address, properties FROM users WHERE name = @name AND email = @email", &usr2, pgx.NamedArgs{"name": "Alice", "email": "alice@example.com"})
	s.Require().NoError(err)
	s.Require().NotEmpty(usr2.UUID.String())
	s.Require().Equal("Alice", usr2.Name)
	s.Require().Equal("Anchorage", usr2.Address.City)
	s.Require().Equal(true, usr2.Properties.Married)
	s.Require().Equal(true, usr2.Properties.Preferences.DarkMode)
	s.Require().Equal("en", usr2.Properties.Preferences.Language)

	// Test QueryOne - no rows returned
	var usr3 User
	err = QueryOne[User](s.T().Context(), s.db.Pool(), "SELECT uuid, id, name, email, address, properties FROM users WHERE name = @name AND email = @email", &usr3, pgx.NamedArgs{"name": "Unknown", "email": "unknown@example.com"})
	s.Require().ErrorIs(err, pgx.ErrNoRows)
}

func (s *DBTestSuite) TestInsert_Integration() {
	uid := uuid.New()
	nu := User{
		UUID:  uid,
		Name:  "Bob",
		Email: "bob@gmail.com",
		Address: Address{
			Street: "main street",
			Zip:    "56456",
			City:   "Portland",
			State:  "Oregon",
		},
		Properties: Properties{
			Married: true,
			Preferences: Preferences{
				DarkMode: true,
				Language: "en",
				TimeZone: "Europe/London",
			},
		},
	}

	err := Exec[User](s.T().Context(), s.db.Pool(), `
		INSERT INTO users (uuid, name, email, address, properties) 
			VALUES (@uuid, @name, @email, @address, @properties)
	`, nu)
	s.Require().NoError(err)

	var u User
	err = QueryOne[User](s.T().Context(), s.db.Pool(), "SELECT * FROM users WHERE uuid=@uuid", &u, pgx.NamedArgs{"uuid": uid})
	s.Require().NoError(err)
	s.Require().Equal("Bob", u.Name)
	s.Require().Equal("Portland", u.Address.City)
	s.Require().Equal("Europe/London", u.Properties.Preferences.TimeZone)
}

func (s *DBTestSuite) TestUpdate_Integration() {
	uid := uuid.New()
	nu := User{
		UUID:  uid,
		Name:  "Bob Update Test Before",
		Email: "bob@gmail.com",
		Address: Address{
			Street: "main street",
			Zip:    "56456",
			City:   "Portland",
			State:  "Oregon",
		},
		Properties: Properties{
			Married: true,
			Preferences: Preferences{
				DarkMode: true,
				Language: "en",
				TimeZone: "Europe/London",
			},
		},
	}

	err := Exec[User](s.T().Context(), s.db.Pool(), `
		INSERT INTO users (uuid, name, email, address, properties) 
			VALUES (@uuid, @name, @email, @address, @properties)
	`, nu)
	s.Require().NoError(err)

	var u User
	err = QueryOne[User](s.T().Context(), s.db.Pool(), "SELECT * FROM users WHERE uuid=@uuid", &u, pgx.NamedArgs{"uuid": uid})
	s.Require().NoError(err)
	s.Require().Equal("Bob Update Test Before", u.Name)
	s.Require().Equal("Portland", u.Address.City)
	s.Require().Equal("Europe/London", u.Properties.Preferences.TimeZone)

	nu.Name = "Bob Update Test After"
	err = Exec[User](s.T().Context(), s.db.Pool(), `
		UPDATE users SET name=@name 
			WHERE uuid=@uuid
	`, nu)
	s.Require().NoError(err)

	err = QueryOne[User](s.T().Context(), s.db.Pool(), "SELECT * FROM users WHERE uuid=@uuid", &u, pgx.NamedArgs{"uuid": uid})
	s.Require().NoError(err)
	s.Require().Equal("Bob Update Test After", u.Name)
	s.Require().Equal("Portland", u.Address.City)
	s.Require().Equal("Europe/London", u.Properties.Preferences.TimeZone)
}

func (s *DBTestSuite) TestTransactionCommit_Integration() {
	uid := uuid.New()
	nu := User{
		UUID:  uid,
		Name:  "Bob Transaction Commit",
		Email: "bob@gmail.com",
		Address: Address{
			Street: "main street",
			Zip:    "56456",
			City:   "Portland",
			State:  "Oregon",
		},
		Properties: Properties{
			Married: true,
			Preferences: Preferences{
				DarkMode: true,
				Language: "en",
				TimeZone: "Europe/London",
			},
		},
	}
	tx, err := s.db.Pool().Begin(s.T().Context())
	s.Require().NoError(err)

	err = Exec[User](s.T().Context(), tx, `
		INSERT INTO users (uuid, name, email, address, properties) 
			VALUES (@uuid, @name, @email, @address, @properties)
	`, nu)
	s.Require().NoError(err)

	var txu User
	err = QueryOne[User](s.T().Context(), tx, "SELECT * FROM users WHERE uuid=@uuid", &txu, pgx.NamedArgs{"uuid": uid})
	s.Require().NoError(err)
	s.Require().Equal("Bob Transaction Commit", txu.Name)
	s.Require().Equal("Portland", txu.Address.City)
	s.Require().Equal("Europe/London", txu.Properties.Preferences.TimeZone)

	err = tx.Commit(s.T().Context())
	s.Require().NoError(err)

	var u User
	err = QueryOne[User](s.T().Context(), s.db.Pool(), "SELECT * FROM users WHERE uuid=@uuid", &u, pgx.NamedArgs{"uuid": uid})
	s.Require().NoError(err)
	s.Require().Equal("Bob Transaction Commit", u.Name)
	s.Require().Equal("Portland", u.Address.City)
	s.Require().Equal("Europe/London", u.Properties.Preferences.TimeZone)
}

func (s *DBTestSuite) TestTransactionRollback_Integration() {
	uid := uuid.New()
	nu := User{
		UUID:  uid,
		Name:  "Bob Transaction Rollback",
		Email: "bob@gmail.com",
		Address: Address{
			Street: "main street",
			Zip:    "56456",
			City:   "Portland",
			State:  "Oregon",
		},
		Properties: Properties{
			Married: true,
			Preferences: Preferences{
				DarkMode: true,
				Language: "en",
				TimeZone: "Europe/London",
			},
		},
	}
	tx, err := s.db.Pool().Begin(s.T().Context())
	s.Require().NoError(err)

	err = Exec[User](s.T().Context(), tx, `
		INSERT INTO users (uuid, name, email, address, properties) 
			VALUES (@uuid, @name, @email, @address, @properties)
	`, nu)
	s.Require().NoError(err)

	var txu User
	err = QueryOne[User](s.T().Context(), tx, "SELECT * FROM users WHERE uuid=@uuid", &txu, pgx.NamedArgs{"uuid": uid})
	s.Require().NoError(err)
	s.Require().Equal("Bob Transaction Rollback", txu.Name)
	s.Require().Equal("Portland", txu.Address.City)
	s.Require().Equal("Europe/London", txu.Properties.Preferences.TimeZone)

	err = tx.Rollback(s.T().Context())
	s.Require().NoError(err)

	var u User
	err = QueryOne[User](s.T().Context(), s.db.Pool(), "SELECT * FROM users WHERE uuid=@uuid", &u, pgx.NamedArgs{"uuid": uid})
	s.Require().ErrorIs(err, pgx.ErrNoRows)
}

/*
 * Setup
	   ____    __
	  / __/__ / /___ _____
	 _\ \/ -_) __/ // / _ \
	/___/\__/\__/\_,_/ .__/
					/_/
*/

func setupPostgres() (connStr string, cleanup func(), err error) {
	testPool, err := dockertest.NewPool("")
	if err != nil {
		return "", func() {}, err
	}

	resource, err := testPool.Run("postgres", "15-alpine", []string{
		"POSTGRES_USER=postgres",
		"POSTGRES_PASSWORD=secret",
		"POSTGRES_DB=testdb",
	})

	if err != nil {
		return "", func() {}, err
	}

	// Expire container after 10 minutes to avoid leaks in CI
	err = resource.Expire(600)
	if err != nil {
		return "", func() {}, err
	}

	//connStr = fmt.Sprintf("postgres://postgres:secret@localhost:%s/testdb?sslmode=disable", resource.GetPort("5432/tcp"))

	cfg := ConnectionConfig{}

	q := make(url.Values)
	q.Set("sslmode", "disable")
	q.Set("timezone", "utc")

	u := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword("postgres", "secret"),
		Host:     fmt.Sprintf("localhost:%s", resource.GetPort("5432/tcp")),
		Path:     "testdb",
		RawQuery: q.Encode(),
	}

	connStr := u.String()

	// Exponential backoff to wait for Postgres readiness
	testPool.MaxWait = 2 * time.Minute
	err = testPool.Retry(func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		conn, err := pgx.Connect(ctx, connStr)
		if err != nil {
			return err
		}
		defer conn.Close(ctx)
		return conn.Ping(ctx)
	})
	if err != nil {
		err = testPool.Purge(resource)
		if err != nil {
			return "", func() {}, err
		}
	}

	cleanup = func() {
		if err := testPool.Purge(resource); err != nil {
			log.Printf("could not purge resource: %v", err)
		}
	}

	return connStr, cleanup, nil
}

func (s *DBTestSuite) seed() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	for _, u := range getSeedData() {
		var id int64
		err := s.db.Pool().QueryRow(ctx,
			`INSERT INTO users (uuid, name, email, address, properties) VALUES ($1, $2, $3, $4, $5) RETURNING id`, u.UUID, u.Name, u.Email, u.Address, u.Properties).Scan(&id)
		if err != nil {
			return err
		}
	}
	return nil
}

func getSeedData() []User {
	return []User{
		{
			UUID:  uuid.New(),
			Name:  "Alice",
			Email: "alice@example.com",
			Address: Address{
				Street: "123 street",
				Zip:    "1234",
				City:   "Anchorage",
				State:  "Alaska",
			},
			Properties: Properties{
				Married: true,
				Preferences: Preferences{
					DarkMode: true,
					Language: "en",
					TimeZone: "Asia/Tokyo",
				},
			},
		},
		{
			UUID:  uuid.New(),
			Name:  "Bob",
			Email: "bob@example.com",
			Address: Address{
				Street: "main street",
				Zip:    "145645644",
				City:   "San Francisco",
				State:  "California",
			},
			Properties: Properties{
				Married: false,
				Preferences: Preferences{
					DarkMode: false,
					Language: "sp",
					TimeZone: "America/Los_Angeles",
				},
			},
		},
	}
}
