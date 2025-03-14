package handlers

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	db "github.com/caleb-mwasikira/tap_gopay/database"
	"github.com/caleb-mwasikira/tap_gopay/handlers/api"
	"github.com/caleb-mwasikira/tap_gopay/utils"
	"github.com/caleb-mwasikira/tap_gopay/validators"
	"github.com/golang-jwt/jwt"
)

var (
	secretKey string

	ErrInvalidPasswordFormat error = errors.New("invalid password format stored in database")
)

func init() {
	utils.LoadEnvVariables()
	secretKey = os.Getenv("SECRET_KEY")
}

func HandleSignUp(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")

	user := validators.RegisterForm{}
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		api.Error(
			w,
			"Invalid JSON data provided as input",
			err,
			http.StatusBadRequest,
		)
		return
	}

	// validate register credentials
	errs := validators.ValidateStruct(user)
	if len(errs) != 0 {
		api.SendResponse(
			w,
			"Validation errors",
			nil, errs,
			http.StatusBadRequest,
		)
		return
	}

	// hash user password
	user.Password = hashPassword(user.Password, nil)

	// check if account already exists
	dbUser, err := db.GetUser(user.Email)
	if err != nil && err != sql.ErrNoRows {
		api.Error(
			w,
			"Unexpected error registering user",
			err,
			http.StatusBadRequest,
		)
		return
	}

	if dbUser != nil {
		api.Error(
			w,
			"User account already exists",
			nil,
			http.StatusConflict,
		)
		return
	}

	// create new user account
	err = db.CreateUser(user)
	if err != nil {
		api.Error(
			w,
			"Unexpected error registering user",
			err,
			http.StatusBadRequest,
		)
		return
	}

	api.SendResponse(
		w,
		"Registration successful",
		user, nil,
		http.StatusCreated,
	)
}

func HandleLogin(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")

	user := validators.LoginForm{}
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		api.Error(
			w,
			"Invalid JSON data provided as input",
			err,
			http.StatusBadRequest,
		)
		return
	}

	errs := validators.ValidateStruct(user)
	if len(errs) != 0 {
		api.SendResponse(
			w,
			"Validation errors",
			nil, errs,
			http.StatusBadRequest,
		)
		return
	}

	// fetch database user
	dbUser, err := db.GetUser(user.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			api.SendResponse(
				w,
				"User account does not exist",
				nil, nil,
				http.StatusUnauthorized,
			)
			return
		}

		api.Error(
			w,
			"Unexpected error loggin in user",
			err,
			http.StatusInternalServerError,
		)
		return
	}

	passwordMatch := verifyPassword(user.Password, dbUser.Password)
	if !passwordMatch {
		api.Error(
			w,
			"Invalid username or password",
			err,
			http.StatusForbidden,
		)
		return
	}

	signedToken, err := createToken(*dbUser)
	if err != nil {
		api.Error(
			w,
			"Unexpected error loggin in user",
			err,
			http.StatusInternalServerError,
		)
		return
	}

	api.SendResponse(
		w,
		"Login successful",
		signedToken, nil,
		http.StatusOK,
	)
}

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")

		if authHeader == "" {
			api.Error(
				w,
				"Authorization required",
				nil,
				http.StatusUnauthorized,
			)
			return
		}

		fields := strings.Split(authHeader, " ")
		if len(fields) != 2 || fields[0] != "Bearer" {
			api.Error(
				w,
				"Invalid Authorization headers. Must in the format Bearer API_KEY",
				nil,
				http.StatusUnauthorized,
			)
			return
		}

		token := fields[1]
		claims, err := verifyToken(token)
		if err != nil {
			api.Error(
				w,
				"Invalid Authorization token",
				err,
				http.StatusUnauthorized,
			)
			return
		}

		user, err := extractFromClaims[db.User](claims)
		if err != nil {
			api.Error(
				w,
				"Invalid Authorization token",
				fmt.Errorf("error extracting user from JWT token; %v", err),
				http.StatusInternalServerError,
			)
			return
		}

		// set user object in request context
		ctx := context.WithValue(r.Context(), "user", user)
		new_req := r.WithContext(ctx)

		next.ServeHTTP(w, new_req)
	})
}

func createToken(user db.User) (string, error) {
	data, err := json.Marshal(user)
	if err != nil {
		return "", err
	}

	aud := base64.RawStdEncoding.EncodeToString(data)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": user.Email,                            // Subject (user identifier)
		"iss": "tap_gopay",                           // Issuer
		"aud": aud,                                   // Audience (user data)
		"exp": time.Now().Add(24 * time.Hour).Unix(), // Expiration time
		"iat": time.Now().Unix(),                     // Issued at
	})

	signedToken, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", fmt.Errorf("error signing JWT; %v", err)
	}

	return signedToken, nil
}

func verifyToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(*jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	} else {
		return nil, fmt.Errorf("invalid JWT token")
	}
}

func extractFromClaims[T any](claims jwt.MapClaims) (*T, error) {
	audVal, ok := claims["aud"]
	if !ok {
		return nil, fmt.Errorf("missing 'aud' key in JWT claims")
	}

	b64EncodedData, ok := audVal.(string)
	if !ok {
		return nil, fmt.Errorf("invalid type on 'aud' key in JWT claims")
	}

	data, err := base64.RawStdEncoding.DecodeString(b64EncodedData)
	if err != nil {
		return nil, err
	}

	var result T
	err = json.Unmarshal(data, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func getLoggedInUser(ctx context.Context) *db.User {
	user, _ := ctx.Value("user").(*db.User)
	return user
}

/*
hashes a password using HMAC + SHA256 alg and returns hashed result
in the format $id$salt$hashed.

$id is the algoirthm prefix used on Linux as follows:

	$1$ is MD5
	$2a$ is Blowfish
	$2y$ is Blowfish
	$5$ is SHA-256
	$6$ is SHA-512
	$y$ is yescrypt
*/
func hashPassword(password string, salt []byte) string {
	if len(salt) == 0 {
		salt = generateSalt(32)
	}

	saltAndPassword := fmt.Sprintf("%s.%s", hex.EncodeToString(salt), password)
	h := hmac.New(sha256.New, []byte(secretKey))

	_, err := h.Write([]byte(saltAndPassword))
	if err != nil {
		log.Fatalf("error hashing password: %v\n", err)
	}

	return fmt.Sprintf("$5$%x$%x", salt, h.Sum(nil))
}

func verifyPassword(password, dbPassword string) bool {
	// extract salt and HMAC from dbPassword
	dbPassword = strings.Trim(dbPassword, "$")
	fields := strings.Split(dbPassword, "$")
	if len(fields) != 3 {
		log.Println("Invalid password format")
		return false
	}

	expectedHmac, err := hex.DecodeString(fields[2]) // decode stored HMAC from hex
	if err != nil {
		log.Println("Invalid HMAC encoding")
		return false
	}

	// compute HMAC with the same salt
	// use original hex-encoded salt string
	saltAndPassword := fmt.Sprintf("%s.%s", fields[1], password)
	h := hmac.New(sha256.New, []byte(secretKey))

	_, err = h.Write([]byte(saltAndPassword))
	if err != nil {
		log.Fatalf("error hashing password: %v\n", err)
	}

	actualHmac := h.Sum(nil)
	return hmac.Equal(expectedHmac, actualHmac)
}

func generateSalt(length int) []byte {
	salt := make([]byte, length)
	_, err := rand.Read(salt)
	if err != nil {
		log.Fatalf("error reading random value into bytes; %v\n", err)
	}
	return salt
}
