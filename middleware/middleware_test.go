package middleware

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultSkipper(t *testing.T) {
	assert.False(t, DefaultSkipper(nil))
}
