package config

import (
	"os"
	"path/filepath"
	"runtime"

	"github.com/joho/godotenv"
)

var (
	_, b, _, _ = runtime.Caller(0)
	basepath   = filepath.Dir(b)
)

// InitEnv loads the environment variables from the .env file if in development mode
func InitEnv() string {

	if os.Getenv("ENV") == "" {
		err := godotenv.Load(basepath + "/.env")
		if err != nil {
			return err.Error() + basepath
		}
	}

	return ""
}
