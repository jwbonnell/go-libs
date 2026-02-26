package validate

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestStruct is a sample struct for validation
type TestStruct struct {
	Name    string `json:"name" validate:"required"`
	Email   string `json:"email" validate:"required,email"`
	Country string `validate:"required"`
}

// TestCheck_Valid tests the Check function with valid input
func TestCheck_Valid(t *testing.T) {
	v := TestStruct{Name: "John Doe", Email: "john.doe@example.com", Country: "US"}
	err := Check(v)
	assert.NoError(t, err)
}

// TestCheck_Invalid tests the Check function with invalid input
func TestCheck_Invalid(t *testing.T) {
	v := TestStruct{Name: "", Email: "invalid-email", Country: ""}
	err := Check(v)

	assert.Error(t, err)
	vErrors, ok := err.(VErrors)
	assert.True(t, ok)
	assert.Len(t, vErrors, 3)

	assert.Equal(t, "name", vErrors[0].Field)
	assert.Equal(t, "name is a required field", vErrors[0].Err)

	assert.Equal(t, "email", vErrors[1].Field)
	assert.Equal(t, "email must be a valid email address", vErrors[1].Err)

	assert.Equal(t, "Country", vErrors[2].Field)
	assert.Equal(t, "Country is a required field", vErrors[2].Err)
}

// TestCheck_Nil tests the Check function with a nil value
func TestCheck_Nil(t *testing.T) {
	err := Check(nil)
	assert.Error(t, err)
}
