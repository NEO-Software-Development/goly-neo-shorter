package model

import (
	"database/sql/driver"
	"goly-app/database"
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

// Mock for gorm.DB
// This is a manual mock for the gorm.DB object. It is used to simulate the database in the tests.
// It only implements the methods that are used in the code under test.
type mockGormDB struct {
	*gorm.DB
	err      error
	expected interface{}
}

func (m *mockGormDB) Find(dest interface{}, conds ...interface{}) *gorm.DB {
	if m.err != nil {
		return &gorm.DB{Error: m.err}
	}
	// This is a simplified mock. It just returns the expected data.
	// A more complex mock would inspect the `dest` and `conds` arguments.
	switch d := dest.(type) {
	case *[]Goly:
		*d = m.expected.([]Goly)
	}
	return &gorm.DB{Error: nil}
}

func (m *mockGormDB) Where(query interface{}, args ...interface{}) *gorm.DB {
	// This mock doesn't implement Where fully. It just returns itself.
	return &gorm.DB{Error: m.err}
}

func (m *mockGormDB) First(dest interface{}, conds ...interface{}) *gorm.DB {
	if m.err != nil {
		return &gorm.DB{Error: m.err}
	}
	switch d := dest.(type) {
	case *Goly:
		*d = m.expected.(Goly)
	}
	return &gorm.DB{Error: nil}
}

func (m *mockGormDB) Create(value interface{}) *gorm.DB {
	if m.err != nil {
		return &gorm.DB{Error: m.err}
	}
	return &gorm.DB{Error: nil}
}

func (m *mockGormDB) Save(value interface{}) *gorm.DB {
	if m.err != nil {
		return &gorm.DB{Error: m.err}
	}
	return &gorm.DB{Error: nil}
}

func (m *mockGormDB) Unscoped() *gorm.DB {
	return &gorm.DB{Error: m.err}
}

func (m *mockGormDB) Delete(value interface{}, conds ...interface{}) *gorm.DB {
	if m.err != nil {
		return &gorm.DB{Error: m.err}
	}
	return &gorm.DB{Error: nil}
}

// TestGetAllGolies is a table-driven test for the GetAllGolies function.
func TestGetAllGolies(t *testing.T) {
	// --- Test Cases ---
	testCases := []struct {
		name          string
		mockDB        *mockGormDB
		expected      []Goly
		expectedError bool
	}{
		{
			name: "should return all golies successfully",
			mockDB: &mockGormDB{
				expected: []Goly{
					{ID: 1, Goly: "abc", Redirect: "https://example.com"},
					{ID: 2, Goly: "def", Redirect: "https://google.com"},
				},
			},
			expected: []Goly{
				{ID: 1, Goly: "abc", Redirect: "https://example.com"},
				{ID: 2, Goly: "def", Redirect: "https://google.com"},
			},
			expectedError: false,
		},
		{
			name: "should return an error if the database call fails",
			mockDB: &mockGormDB{
				err: gorm.ErrRecordNotFound,
			},
			expected:      []Goly{},
			expectedError: true,
		},
	}

	// --- Run Tests ---
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Replace the real DB with the mock DB
			originalDB := database.DB
			database.DB = tc.mockDB
			defer func() { database.DB = originalDB }()

			// Call the function under test
			golies, err := GetAllGolies()

			// --- Assertions ---
			assert.Equal(t, tc.expected, golies)
			if tc.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetGoly(t *testing.T) {
	testCases := []struct {
		name          string
		mockDB        *mockGormDB
		inputID       uint64
		expected      Goly
		expectedError bool
	}{
		{
			name: "should return a goly successfully",
			mockDB: &mockGormDB{
				expected: Goly{ID: 1, Goly: "abc", Redirect: "https://example.com"},
			},
			inputID:       1,
			expected:      Goly{ID: 1, Goly: "abc", Redirect: "https://example.com"},
			expectedError: false,
		},
		{
			name: "should return an error if the goly is not found",
			mockDB: &mockGormDB{
				err: gorm.ErrRecordNotFound,
			},
			inputID:       1,
			expected:      Goly{},
			expectedError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			originalDB := database.DB
			database.DB = tc.mockDB
			defer func() { database.DB = originalDB }()

			goly, err := GetGoly(tc.inputID)

			assert.Equal(t, tc.expected, goly)
			if tc.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCreateGoly(t *testing.T) {
	testCases := []struct {
		name          string
		mockDB        *mockGormDB
		inputGoly     Goly
		expectedError bool
	}{
		{
			name: "should create a goly successfully",
			mockDB: &mockGormDB{},
			inputGoly:     Goly{Goly: "abc", Redirect: "https://example.com"},
			expectedError: false,
		},
		{
			name: "should return an error if the database call fails",
			mockDB: &mockGormDB{
				err: gorm.ErrInvalidDB,
			},
			inputGoly:     Goly{Goly: "abc", Redirect: "https://example.com"},
			expectedError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			originalDB := database.DB
			database.DB = tc.mockDB
			defer func() { database.DB = originalDB }()

			err := CreateGoly(tc.inputGoly)

			if tc.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUpdateGoly(t *testing.T) {
	testCases := []struct {
		name          string
		mockDB        *mockGormDB
		inputGoly     Goly
		expectedError bool
	}{
		{
			name: "should update a goly successfully",
			mockDB: &mockGormDB{},
			inputGoly:     Goly{ID: 1, Goly: "abc", Redirect: "https://example.com"},
			expectedError: false,
		},
		{
			name: "should return an error if the database call fails",
			mockDB: &mockGormDB{
				err: gorm.ErrInvalidDB,
			},
			inputGoly:     Goly{ID: 1, Goly: "abc", Redirect: "https://example.com"},
			expectedError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			originalDB := database.DB
			database.DB = tc.mockDB
			defer func() { database.DB = originalDB }()

			err := UpdateGoly(tc.inputGoly)

			if tc.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDeleteGoly(t *testing.T) {
	testCases := []struct {
		name          string
		mockDB        *mockGormDB
		inputID       uint64
		expectedError bool
	}{
		{
			name: "should delete a goly successfully",
			mockDB: &mockGormDB{},
			inputID:       1,
			expectedError: false,
		},
		{
			name: "should return an error if the database call fails",
			mockDB: &mockGormDB{
				err: gorm.ErrInvalidDB,
			},
			inputID:       1,
			expectedError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			originalDB := database.DB
			database.DB = tc.mockDB
			defer func() { database.DB = originalDB }()

			err := DeleteGoly(tc.inputID)

			if tc.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestFindByGolyUrl(t *testing.T) {
	testCases := []struct {
		name          string
		mockDB        *mockGormDB
		inputUrl      string
		expected      Goly
		expectedError bool
	}{
		{
			name: "should find a goly by url successfully",
			mockDB: &mockGormDB{
				expected: Goly{ID: 1, Goly: "abc", Redirect: "https://example.com"},
			},
			inputUrl:      "abc",
			expected:      Goly{ID: 1, Goly: "abc", Redirect: "https://example.com"},
			expectedError: false,
		},
		{
			name: "should return an error if the goly is not found",
			mockDB: &mockGormDB{
				err: gorm.ErrRecordNotFound,
			},
			inputUrl:      "abc",
			expected:      Goly{},
			expectedError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			originalDB := database.DB
			database.DB = tc.mockDB
			defer func() { database.DB = originalDB }()

			goly, err := FindByGolyUrl(tc.inputUrl)

			assert.Equal(t, tc.expected, goly)
			if tc.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// To run these tests, you would use the following command:
// go test -v ./goly/model

// Note: As I was unable to run the application or the tests in the provided environment,
// these tests have been written "blind" and may contain errors. They should be run and
// verified in a stable environment.

// Mock for sql.driver.Value
type anyValue struct{}

func (a anyValue) Value() (driver.Value, error) {
	return "any", nil
}
