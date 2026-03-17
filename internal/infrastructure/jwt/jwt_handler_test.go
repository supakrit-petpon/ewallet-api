package jwt

import (
	"testing"

	"github.com/golang-jwt/jwt"
	"github.com/stretchr/testify/assert"
)

func TestJWTHandler_GenerateToken(t *testing.T) {
	

	t.Run("should generate valid token string", func(t *testing.T) {
		secret := "test-secret-key"
		handler := &JWTHandler{SecretKey: secret}
		userId := uint(123)

        tokenString, err := handler.GenerateToken(userId)

        assert.NoError(t, err)
        assert.NotEmpty(t, tokenString)

        // ตรวจสอบไส้ในของ Token
        parsedToken, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
            return []byte(secret), nil
        })

        assert.NoError(t, err)
        assert.True(t, parsedToken.Valid)

        claims := parsedToken.Claims.(jwt.MapClaims)
        // หมายเหตุ: jwt-go จะมองตัวเลขเป็น float64 เมื่อ parse กลับมา
        assert.Equal(t, float64(userId), claims["user_id"])
    })
	t.Run("secret key is missing", func(t *testing.T) {
		secret := ""
		handler := &JWTHandler{SecretKey: secret}
		userId := uint(123)

        tokenString, err := handler.GenerateToken(userId)

        assert.Error(t, err)
        assert.Equal(t, "", tokenString)
    	assert.Contains(t, err.Error(), "secret key is missing")
    })

}

func TestNewTokenProvider(t *testing.T) {
	t.Run("should return a valid Tth secret keokenProvider wiy", func(t *testing.T) {
        
        expectedSecret := "super-secret-key"
        provider := NewTokenProvider(expectedSecret)

        assert.NotNil(t, provider)

        handler, ok := provider.(*JWTHandler)
        assert.True(t, ok, "Provider should be of type *JWTHandler")

        assert.Equal(t, expectedSecret, handler.SecretKey)
    })
}