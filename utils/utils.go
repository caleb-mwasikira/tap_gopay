package utils

import (
	"log"
	"path/filepath"
	"runtime"

	"github.com/joho/godotenv"
)

var (
	ProjectDir    string
	EmailViewsDir string
)

func init() {
	_, fname, _, ok := runtime.Caller(0)
	if !ok {
		log.Fatalln("error acquiring program file path")
	}

	ProjectDir = filepath.Dir(filepath.Dir(fname))
	EmailViewsDir = filepath.Join(ProjectDir, "views/emails")
}

func LoadEnvVariables() {
	envFile := filepath.Join(ProjectDir, ".env")
	err := godotenv.Load(envFile)
	if err != nil {
		log.Fatalf("error loading environment variables; %v\n", err)
	}
}

func StringToRuneSlice(s string) []string {
	runes := []rune(s)
	chars := make([]string, len(runes))
	for i, r := range runes {
		chars[i] = string(r)
	}
	return chars
}
