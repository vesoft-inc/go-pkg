package dotenv

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultEnvFile(t *testing.T) {
	cwd, err := os.Getwd()
	assert.NoError(t, err)
	defer func() {
		assert.NoError(t, os.Chdir(cwd))
	}()

	assert.NoError(t, os.Chdir("testdata"))

	initDotEnv()

	assert.Equal(t, os.Getenv("TEST_ENV_NOT_EXISTS"), "")
	assert.Equal(t, os.Getenv("TEST_ENV1"), "1")
	assert.Equal(t, os.Getenv("TEST_ENV2"), "2")
}

func TestCustomEnvFile(t *testing.T) {
	assert.NoError(t, os.Setenv("DOT_ENV_FILES", "./testdata/.env;./testdata/custom.env"))

	initDotEnv()

	assert.Equal(t, os.Getenv("TEST_ENV_NOT_EXISTS"), "")
	assert.Equal(t, os.Getenv("TEST_ENV1"), "1")
	assert.Equal(t, os.Getenv("TEST_ENV2"), "2")
	assert.Equal(t, os.Getenv("TEST_CUSTOM_ENV1"), "1")
	assert.Equal(t, os.Getenv("TEST_CUSTOM_ENV2"), "2")
}

func TestCustomEnvFileNotFound(t *testing.T) {
	assert.NoError(t, os.Setenv("DOT_ENV_FILES", "./testdata/.env;./testdata/not-exists.env"))
	assert.Panics(t, initDotEnv)
}
