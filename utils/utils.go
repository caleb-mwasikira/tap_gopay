package utils

import (
	"crypto/rand"
	"fmt"
	"log"
	"math/big"
	"path/filepath"
	"runtime"
	"strings"

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

func RandNumbers(len int) string {
	nums := []string{}

	for i := 0; i < len; i++ {
		bigNum, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			log.Printf("error generating random number; %v\n", err)
			return ""
		}
		nums = append(nums, fmt.Sprintf("%v", bigNum.Int64()))
	}

	return strings.Join(nums, "")
}

func StringToRuneSlice(s string) []string {
	runes := []rune(s)
	chars := make([]string, len(runes))
	for i, r := range runes {
		chars[i] = string(r)
	}
	return chars
}
