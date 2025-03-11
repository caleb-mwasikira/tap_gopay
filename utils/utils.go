package utils

import (
	"log"
	"path/filepath"
	"runtime"

	"github.com/joho/godotenv"
)

func LoadEnvVariables() {
	_, fname, _, ok := runtime.Caller(0)
	if !ok {
		log.Fatalln("error acquiring program file path")
	}

	projectDir := filepath.Dir(filepath.Dir(fname))

	envFile := filepath.Join(projectDir, ".env")
	err := godotenv.Load(envFile)
	if err != nil {
		log.Fatalf("error loading environment variables; %v\n", err)
	}
}
