package dotenv

import (
	stderrors "errors"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

func init() { //nolint:gochecknoinits
	initDotEnv()
}

func initDotEnv() {
	var envFiles []string
	if files := os.Getenv("DOT_ENV_FILES"); files != "" {
		envFiles = strings.Split(files, ";")
	}

	if err := godotenv.Load(envFiles...); err != nil {
		if len(envFiles) > 0 || !stderrors.Is(err, os.ErrNotExist) {
			panic(err)
		}
		// ignore errors if the default .env file does not exist
	}
}
