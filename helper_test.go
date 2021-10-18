package exasol

import (
	"database/sql/driver"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNamedValuesToValues(t *testing.T) {
	namedValues := []driver.NamedValue{{Name: ""}, {Name: ""}}
	values, err := namedValuesToValues(namedValues)
	assert.Equal(t, []driver.Value{driver.Value(nil), driver.Value(nil)}, values)
	assert.NoError(t, err)
}

func TestNamedValuesToValuesInvalidName(t *testing.T) {
	namedValues := []driver.NamedValue{{Name: "some name"}}
	values, err := namedValuesToValues(namedValues)
	assert.Nil(t, values)
	assert.EqualError(t, err, "E-EGOD-7: named parameters not supported")
}
