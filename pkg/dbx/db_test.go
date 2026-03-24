package dbx

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jwbonnell/go-libs/pkg/logx"
	"github.com/stretchr/testify/suite"

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
	logger := logx.NewCILogger("integration-tests")
	connCfg, cleanup, err := setupPostgres(logger)
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	d, err := New(ctx, connCfg, logger)
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

	err := Exec(s.T().Context(), s.db.Pool(), `
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

func (s *DBTestSuite) TestInsertReturning_Integration() {
	uid := uuid.New()
	nu := User{
		UUID:  uid,
		Name:  "TestInsertReturning_Integration",
		Email: "TestInsertReturning_Integration",
		Address: Address{
			Street: "1st street",
			Zip:    "12345",
			City:   "Detroit",
			State:  "Michigan",
		},
		Properties: Properties{
			Married: true,
			Preferences: Preferences{
				DarkMode: true,
				Language: "en",
				TimeZone: "America/Detroit",
			},
		},
	}

	var dest User
	err := ExecReturn[User](s.T().Context(), s.db.Pool(), `
		INSERT INTO users (uuid, name, email, address, properties) 
			VALUES (@uuid, @name, @email, @address, @properties)
		RETURNING *
	`, &dest, nu)
	s.Require().NoError(err)
	s.Require().Equal("TestInsertReturning_Integration", dest.Name)
	s.Require().Equal("Detroit", dest.Address.City)
	s.Require().Equal("America/Detroit", dest.Properties.Preferences.TimeZone)

	var u User
	err = QueryOne[User](s.T().Context(), s.db.Pool(), "SELECT * FROM users WHERE uuid=@uuid", &u, pgx.NamedArgs{"uuid": uid})
	s.Require().NoError(err)
	s.Require().Equal("TestInsertReturning_Integration", u.Name)
	s.Require().Equal("Detroit", u.Address.City)
	s.Require().Equal("America/Detroit", u.Properties.Preferences.TimeZone)
}

func (s *DBTestSuite) TestInsertReturningMultiple_Integration() {
	uid := uuid.New()
	uid2 := uuid.New()
	nus := []User{
		{
			UUID:  uid,
			Name:  "TestInsertReturningMultiple_Integration",
			Email: "TestInsertReturningMultiple_Integration",
			Address: Address{
				Street: "2nd street",
				Zip:    "12345",
				City:   "Livonia",
				State:  "Michigan",
			},
			Properties: Properties{
				Married: true,
				Preferences: Preferences{
					DarkMode: true,
					Language: "en",
					TimeZone: "America/Detroit",
				},
			},
		},
		{
			UUID:  uid2,
			Name:  "TestInsertReturningMultiple_Integration2",
			Email: "TestInsertReturningMultiple_Integration2",
			Address: Address{
				Street: "3rd street",
				Zip:    "12345",
				City:   "Troy",
				State:  "Michigan",
			},
			Properties: Properties{
				Married: true,
				Preferences: Preferences{
					DarkMode: true,
					Language: "en",
					TimeZone: "America/New York",
				},
			},
		},
	}

	var dest []User
	err := ExecReturnMany[User](s.T().Context(), s.db.Pool(), `
		INSERT INTO users (uuid, name, email, address, properties)
			VALUES (@uuid1, @name1, @email1, @address1, @properties1),
				   (@uuid2, @name2, @email2, @address2, @properties2)
		RETURNING *
	`, &dest, nus)
	s.Require().NoError(err)
	s.Require().NotNil(dest)
	s.Require().Equal("TestInsertReturningMultiple_Integration", dest[0].Name)
	s.Require().Equal("Livonia", dest[0].Address.City)
	s.Require().Equal("America/Detroit", dest[0].Properties.Preferences.TimeZone)
	s.Require().Equal("TestInsertReturningMultiple_Integration2", dest[1].Name)
	s.Require().Equal("Troy", dest[1].Address.City)
	s.Require().Equal("America/New York", dest[1].Properties.Preferences.TimeZone)

	var u User
	err = QueryOne[User](s.T().Context(), s.db.Pool(), "SELECT * FROM users WHERE uuid=@uuid", &u, pgx.NamedArgs{"uuid": uid})
	s.Require().NoError(err)
	s.Require().Equal("TestInsertReturningMultiple_Integration", u.Name)
	s.Require().Equal("Livonia", u.Address.City)
	s.Require().Equal("America/Detroit", u.Properties.Preferences.TimeZone)
}

func (s *DBTestSuite) TestExecReturnMany_LargeDataSet() {
	// Test with many rows
	users := make([]User, 10)
	for i := 0; i < 10; i++ {
		users[i] = User{
			UUID:  uuid.New(),
			Name:  fmt.Sprintf("Bulk User %d", i+1),
			Email: fmt.Sprintf("bulk%d@example.com", i+1),
			Address: Address{
				Street: fmt.Sprintf("%d Main St", i+1),
				Zip:    fmt.Sprintf("%05d", 10000+i),
				City:   fmt.Sprintf("City%d", i+1),
				State:  "TestState",
			},
			Properties: Properties{
				Married: true,
				Preferences: Preferences{
					DarkMode: true,
					Language: "en",
					TimeZone: "America/Detroit",
				},
			},
		}
	}

	var dest []User
	// Build dynamic SQL for 10 rows
	sqlParts := make([]string, 10)
	for i := 0; i < 10; i++ {
		sqlParts[i] = fmt.Sprintf("(@uuid%d, @name%d, @email%d, @address%d, @properties%d)", i+1, i+1, i+1, i+1, i+1)
	}
	sql := fmt.Sprintf(`
		INSERT INTO users (uuid, name, email, address, properties)
		VALUES %s
		RETURNING *
	`, strings.Join(sqlParts, ","))

	err := ExecReturnMany[User](s.T().Context(), s.db.Pool(), sql, &dest, users)
	s.Require().NoError(err)
	s.Require().Len(dest, 10)

	// Verify all rows were inserted
	for i, u := range dest {
		s.Require().Equal(fmt.Sprintf("Bulk User %d", i+1), u.Name)
	}
}

func (s *DBTestSuite) TestExecReturnMany_NoRowsReturned() {
	// Insert a user first
	uid := uuid.New()
	user := []User{
		{
			UUID:  uid,
			Name:  "NoRows Test",
			Email: "norows@example.com",
			Address: Address{
				Street: "Test Street",
				Zip:    "12345",
				City:   "TestCity",
				State:  "TestState",
			},
			Properties: Properties{
				Married: true,
				Preferences: Preferences{
					DarkMode: true,
					Language: "en",
					TimeZone: "America/Detroit",
				},
			},
		},
	}

	var inserted []User
	err := ExecReturnMany[User](s.T().Context(), s.db.Pool(), `
		INSERT INTO users (uuid, name, email, address, properties)
		VALUES (@uuid1, @name1, @email1, @address1, @properties1)
		RETURNING *
	`, &inserted, user)
	s.Require().NoError(err)

	// Try to update with WHERE clause that matches nothing
	updateData := []User{
		{
			UUID:  uuid.New(), // Different UUID
			Name:  "Should Not Update",
			Email: "shouldnotupdate@example.com",
			Address: Address{
				Street: "Update Street",
				Zip:    "54321",
				City:   "UpdateCity",
				State:  "UpdateState",
			},
			Properties: Properties{
				Married: true,
				Preferences: Preferences{
					DarkMode: true,
					Language: "en",
					TimeZone: "America/Detroit",
				},
			},
		},
	}

	var updated []User
	err = ExecReturnMany[User](s.T().Context(), s.db.Pool(), `
		UPDATE users SET name = @name1, email = @email1
		WHERE uuid = @uuid1
		RETURNING *
	`, &updated, updateData)

	// Should return ErrNoRows when no rows match
	s.Require().Error(err)
	s.Require().Equal(pgx.ErrNoRows, err)
}

func (s *DBTestSuite) TestExecReturnMany_WithContext_Cancellation() {
	uid := uuid.New()
	user := User{
		UUID:  uid,
		Name:  "Context Cancel Test",
		Email: "contextcancel@example.com",
	}

	ctx, cancel := context.WithCancel(s.T().Context())
	cancel() // Cancel immediately

	var dest []User
	err := ExecReturnMany[User](ctx, s.db.Pool(), `
		INSERT INTO users (uuid, name, email)
		VALUES (@uuid1, @name1, @email1)
		RETURNING *
	`, &dest, []User{user})

	s.Require().Error(err)
	// Should contain context cancellation error
	s.Require().True(errors.Is(err, context.Canceled) || strings.Contains(err.Error(), "context canceled"))
}

func (s *DBTestSuite) TestExecReturnMany_InsufficientNamedArgs() {
	// Test that missing named args in the struct causes an error
	uid := uuid.New()
	users := []User{
		{
			UUID:  uid,
			Name:  "Missing Args Test",
			Email: "missingargs@example.com",
		},
	}

	var dest []User
	err := ExecReturnMany[User](s.T().Context(), s.db.Pool(), `
		INSERT INTO users (uuid, name, email, address)
		VALUES (@uuid1, @name1, @email1, @address1)
		RETURNING *
	`, &dest, users)

	// Should fail because @address1 is not provided
	s.Require().Error(err)
}

func (s *DBTestSuite) TestExecReturnMany_RollbackOnError() {
	// Test that if one row fails, the entire transaction is rolled back
	uid1 := uuid.New()
	uid2 := uuid.New()

	addr := Address{
		Street: "Update Street",
		Zip:    "54321",
		City:   "UpdateCity",
		State:  "UpdateState",
	}
	props := Properties{
		Married: true,
		Preferences: Preferences{
			DarkMode: true,
			Language: "en",
			TimeZone: "America/Detroit",
		},
	}

	// Create a duplicate email to force a constraint violation
	existingUser := User{
		UUID:       uuid.New(),
		Name:       "Existing User",
		Email:      "duplicate@example.com",
		Address:    addr,
		Properties: props,
	}

	var inserted []User
	err := ExecReturnMany[User](s.T().Context(), s.db.Pool(), `
		INSERT INTO users (uuid, name, email, address, properties)
		VALUES (@uuid1, @name1, @email1, @address1, @properties1)
		RETURNING *
	`, &inserted, []User{existingUser})
	s.Require().NoError(err)

	// Try to insert two users where the second has a duplicate email
	users := []User{
		{
			UUID:       uid1,
			Name:       "New User 1",
			Email:      "newuser1@example.com",
			Address:    addr,
			Properties: props,
		},
		{
			UUID:       uid2,
			Name:       "Duplicate User",
			Email:      "duplicate@example.com", // This should cause a constraint violation
			Address:    addr,
			Properties: props,
		},
	}

	var dest []User
	err = ExecReturnMany[User](s.T().Context(), s.db.Pool(), `
		INSERT INTO users (uuid, name, email, address, properties)
			VALUES (@uuid1, @name1, @email1, @address1, @properties1),
				   (@uuid2, @name2, @email2, @address2, @properties2)
		RETURNING *
	`, &dest, users)

	// Should fail due to duplicate email
	s.Require().Error(err)

	// Verify that the first user was NOT inserted (rollback occurred)
	var count int
	err = s.db.Pool().QueryRow(s.T().Context(),
		"SELECT COUNT(*) FROM users WHERE name = 'New User 1'").Scan(&count)
	s.Require().NoError(err)
	s.Require().Equal(0, count, "User should not have been inserted due to rollback")
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

	err := Exec(s.T().Context(), s.db.Pool(), `
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
	err = Exec(s.T().Context(), s.db.Pool(), `
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

	err = Exec(s.T().Context(), tx, `
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

	err = Exec(s.T().Context(), tx, `
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

func (s *DBTestSuite) TestStructToNamedArgs() {
	uid := uuid.New()
	u := User{
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

	na, err := StructToNamedArgs(u)
	s.Require().NoError(err)
	s.Require().Equal(na["name"], "Bob Transaction Rollback")

	e, ok := na["uuid"].(uuid.UUID)
	s.Require().True(ok)
	s.Require().Equal(e.String(), uid.String())

	e2, ok := na["address"].(Address)
	s.Require().True(ok)
	s.Require().Equal(e2.Street, "main street")

	e3, ok := na["properties"].(Properties)
	s.Require().True(ok)
	s.Require().Equal(e3.Preferences.TimeZone, "Europe/London")
}

func (s *DBTestSuite) TestNamedArgsToStruct() {
	u := uuid.New()
	namedArgs := pgx.NamedArgs{
		"uuid":   u,
		"name":   "TestNamedArgsToStruct",
		"number": 8,
	}

	args, err := StructToNamedArgs(namedArgs)
	s.Require().NoError(err)
	s.Require().Equal(args["uuid"], u)
	s.Require().Equal(args["name"], namedArgs["name"])
	s.Require().Equal(args["number"], namedArgs["number"])
}

func (s *DBTestSuite) TestSliceToNamedArgs() {
	uid1 := uuid.New()
	uid2 := uuid.New()
	users := []User{
		{
			UUID:  uid1,
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
		},
		{
			UUID:  uid2,
			Name:  "Alice Test User",
			Email: "alice@gmail.com",
			Address: Address{
				Street: "oak avenue",
				Zip:    "97214",
				City:   "Seattle",
				State:  "Washington",
			},
			Properties: Properties{
				Married: false,
				Preferences: Preferences{
					DarkMode: false,
					Language: "es",
					TimeZone: "America/Los_Angeles",
				},
			},
		},
	}

	na, err := SliceToNamedArgs(users)
	s.Require().NoError(err)

	// Test first user's fields
	s.Require().Equal(na["name1"], "Bob Transaction Rollback")
	s.Require().Equal(na["email1"], "bob@gmail.com")

	e1, ok := na["uuid1"].(uuid.UUID)
	s.Require().True(ok)
	s.Require().Equal(e1.String(), uid1.String())

	addr1, ok := na["address1"].(Address)
	s.Require().True(ok)
	s.Require().Equal(addr1.Street, "main street")
	s.Require().Equal(addr1.Zip, "56456")

	props1, ok := na["properties1"].(Properties)
	s.Require().True(ok)
	s.Require().True(props1.Married)
	s.Require().Equal(props1.Preferences.TimeZone, "Europe/London")

	// Test second user's fields
	s.Require().Equal(na["name2"], "Alice Test User")
	s.Require().Equal(na["email2"], "alice@gmail.com")

	e2, ok := na["uuid2"].(uuid.UUID)
	s.Require().True(ok)
	s.Require().Equal(e2.String(), uid2.String())

	addr2, ok := na["address2"].(Address)
	s.Require().True(ok)
	s.Require().Equal(addr2.Street, "oak avenue")
	s.Require().Equal(addr2.City, "Seattle")

	props2, ok := na["properties2"].(Properties)
	s.Require().True(ok)
	s.Require().False(props2.Married)
	s.Require().Equal(props2.Preferences.Language, "es")
}

func (s *DBTestSuite) TestSliceToNamedArgs_SingleElement() {
	uid := uuid.New()
	users := []User{
		{
			UUID:  uid,
			Name:  "Single User",
			Email: "single@gmail.com",
			Address: Address{
				Street: "test street",
				Zip:    "12345",
				City:   "TestCity",
				State:  "TestState",
			},
		},
	}

	na, err := SliceToNamedArgs(users)
	s.Require().NoError(err)

	s.Require().Equal(na["name1"], "Single User")
	s.Require().Equal(na["email1"], "single@gmail.com")

	e, ok := na["uuid1"].(uuid.UUID)
	s.Require().True(ok)
	s.Require().Equal(e.String(), uid.String())
}

func (s *DBTestSuite) TestSliceToNamedArgs_EmptySlice() {
	users := []User{}

	na, err := SliceToNamedArgs(users)
	s.Require().NoError(err)
	s.Require().Empty(na)
}

func (s *DBTestSuite) TestSliceToNamedArgs_PointerToSlice() {
	uid := uuid.New()
	users := []User{
		{
			UUID:  uid,
			Name:  "Pointer Test",
			Email: "pointer@gmail.com",
		},
	}

	na, err := SliceToNamedArgs(&users)
	s.Require().NoError(err)
	s.Require().Equal(na["name1"], "Pointer Test")
}

func (s *DBTestSuite) TestSliceToNamedArgs_AlreadyNamedArgs() {
	existingArgs := pgx.NamedArgs{
		"custom1": "value1",
		"custom2": "value2",
	}

	na, err := SliceToNamedArgs(existingArgs)
	s.Require().NoError(err)
	s.Require().Equal(na, existingArgs)
}

func (s *DBTestSuite) TestSliceToNamedArgs_InvalidInput_NotSlice() {
	invalidInput := "not a slice"

	_, err := SliceToNamedArgs(invalidInput)
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "input must be a slice of structs")
}

func (s *DBTestSuite) TestSliceToNamedArgs_InvalidInput_NonStructElements() {
	invalidInput := []string{"not", "structs"}

	_, err := SliceToNamedArgs(invalidInput)
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "slice elements must be structs")
}

func (s *DBTestSuite) TestSliceToNamedArgs_PointerSliceElements() {
	uid1 := uuid.New()
	uid2 := uuid.New()
	users := []*User{
		{
			UUID:  uid1,
			Name:  "Pointer Element 1",
			Email: "ptr1@gmail.com",
		},
		{
			UUID:  uid2,
			Name:  "Pointer Element 2",
			Email: "ptr2@gmail.com",
		},
	}

	na, err := SliceToNamedArgs(users)
	s.Require().NoError(err)

	s.Require().Equal(na["name1"], "Pointer Element 1")
	s.Require().Equal(na["name2"], "Pointer Element 2")
	s.Require().Equal(na["email1"], "ptr1@gmail.com")
	s.Require().Equal(na["email2"], "ptr2@gmail.com")
}

func (s *DBTestSuite) TestSliceToNamedArgs_CustomDBTags() {
	type CustomTagUser struct {
		ID    string `db:"user_id"`
		Name  string `db:"full_name"`
		Email string
	}

	users := []CustomTagUser{
		{
			ID:    "123",
			Name:  "John Doe",
			Email: "john@example.com",
		},
		{
			ID:    "456",
			Name:  "Jane Doe",
			Email: "jane@example.com",
		},
	}

	na, err := SliceToNamedArgs(users)
	s.Require().NoError(err)

	s.Require().Equal(na["user_id1"], "123")
	s.Require().Equal(na["full_name1"], "John Doe")
	s.Require().Equal(na["Email1"], "john@example.com")

	s.Require().Equal(na["user_id2"], "456")
	s.Require().Equal(na["full_name2"], "Jane Doe")
	s.Require().Equal(na["Email2"], "jane@example.com")
}

/*
 * Setup
	   ____    __
	  / __/__ / /___ _____
	 _\ \/ -_) __/ // / _ \
	/___/\__/\__/\_,_/ .__/
					/_/
*/

func setupPostgres(log *logx.Logger) (connCfg ConnectionConfig, cleanup func(), err error) {
	testPool, err := dockertest.NewPool("")
	if err != nil {
		return ConnectionConfig{}, func() {}, err
	}

	resource, err := testPool.Run("postgres", "15-alpine", []string{
		"POSTGRES_USER=postgres",
		"POSTGRES_PASSWORD=secret",
		"POSTGRES_DB=testdb",
	})

	if err != nil {
		return ConnectionConfig{}, func() {}, err
	}

	// Expire container after 10 minutes to avoid leaks in CI
	err = resource.Expire(600)
	if err != nil {
		return ConnectionConfig{}, func() {}, err
	}

	connCfg = ConnectionConfig{
		User:         "postgres",
		Password:     "secret",
		Host:         fmt.Sprintf("localhost:%s", resource.GetPort("5432/tcp")),
		Name:         "testdb",
		MaxIdleConns: 0,
		MaxOpenConns: 0,
		DisableTLS:   true,
	}

	q := make(url.Values)
	q.Set("sslmode", "disable")
	q.Set("timezone", "utc")

	u := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(connCfg.User, connCfg.Password),
		Host:     connCfg.Host,
		Path:     connCfg.Name,
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
			return ConnectionConfig{}, func() {}, err
		}
	}

	cleanup = func() {
		if err := testPool.Purge(resource); err != nil {
			log.Error(context.Background(), fmt.Sprintf("could not purge resource: %v", err))
		}
	}

	return connCfg, cleanup, nil
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
