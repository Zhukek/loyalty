package utils

import (
	"net/http"
	"strconv"
	"time"

	"github.com/Zhukek/loyalty/internal/errs"
	"github.com/Zhukek/loyalty/internal/middlewares"
	"github.com/Zhukek/loyalty/internal/models"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

/// Хеширование пароля и сравнение с кешем

func HashPass(pass string) (string, error) {
	hashedPass, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(hashedPass), nil
}

func CheckPass(hash string, pass string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(pass))
}

/// Создание JWT и вытягивание из него информации о юзере

type claims struct {
	jwt.RegisteredClaims
	UserID   int
	Username string
}

const (
	tokenExp  = time.Hour
	SecretStr = "asdfygausdf"
)

func GenerateJWT(user *models.UserPublic) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims{
		UserID:   user.Id,
		Username: user.Log,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(tokenExp)),
		},
	})

	return token.SignedString([]byte(SecretStr))
}

func GetTokenData(tokenStr string) (*models.UserPublic, error) {
	claims := &claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errs.ErrSigningMethod
		}
		return []byte(SecretStr), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errs.ErrNotValidToken
	}

	user := models.UserPublic{
		Id:  claims.UserID,
		Log: claims.Username,
	}

	return &user, nil
}

/// Работа с Cookie

func GenerateJWTCookie(jwtString string) *http.Cookie {
	return &http.Cookie{
		Name:     "jwt",
		Value:    jwtString,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		MaxAge:   60 * 60 * 10,
	}
}

/// Проверить и вытащить user в handler

func GetUserFromReq(w http.ResponseWriter, req *http.Request) (*models.UserPublic, bool) {
	userValue := req.Context().Value(middlewares.UserKey)

	if userValue == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return nil, false
	}

	user, ok := userValue.(*models.UserPublic)
	return user, ok
}

/// Проверить алгоритмом Луна

func CheckLuhn(order string) bool {
	if len(order) < 2 {
		return false
	}

	sum := 0
	isSecond := false

	for i := len(order) - 1; i >= 0; i-- {
		digit, err := strconv.Atoi(string(order[i]))
		if err != nil {
			return false
		}

		if isSecond {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}

		sum += digit
		isSecond = !isSecond
	}

	return sum%10 == 0
}
