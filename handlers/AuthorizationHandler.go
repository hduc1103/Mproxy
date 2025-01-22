package handlers

import (
	"database/sql"
	"errors"
	"project/models"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

var jwtSecret = []byte("secretKeyForJWT@123456")

func GenerateToken(device models.Device, expiryDuration time.Duration) (string, error) {
	if device.DeviceID == "" {
		return "", errors.New("device ID is required")
	}

	claims := jwt.MapClaims{
		"device_id": device.DeviceID,
		"exp":       time.Now().Add(expiryDuration).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", err
	}

	return signedToken, nil
}

func VerifyToken(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return jwtSecret, nil
	})

	if err != nil {
		return "", err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if deviceID, exists := claims["device_id"].(string); exists {
			return deviceID, nil
		}
		return "", errors.New("device_id not found in token")
	}

	return "", jwt.ErrSignatureInvalid
}

func AuthenticateAndGenerateToken(db *sql.DB, deviceID string, password string, expiryDuration time.Duration) (string, error) {
	query := `SELECT COUNT(*) FROM devices WHERE device_id = ? AND password = ?`
	var count int
	err := db.QueryRow(query, deviceID, password).Scan(&count)
	if err != nil {
		return "", errors.New("error querying the database")
	}
	if count == 0 {
		return "", errors.New("invalid device ID or password")
	}

	device := models.Device{
		DeviceID: deviceID,
	}

	return GenerateToken(device, expiryDuration)
}
