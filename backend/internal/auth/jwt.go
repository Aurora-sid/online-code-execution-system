package auth

import (
	"errors"
	"log"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// SecretKey JWT签名密钥，必须在应用启动时通过 Init() 初始化
var SecretKey []byte

// Init 初始化JWT密钥，从配置或环境变量加载
// 必须在应用启动时调用此函数
func Init(secret string) {
	if secret == "" {
		// 尝试从环境变量获取
		secret = os.Getenv("JWT_SECRET")
	}
	if secret == "" {
		log.Println("[警告] JWT_SECRET 未设置，使用随机生成的临时密钥（不适用于生产环境）")
		// 生产环境应该强制要求设置密钥
		secret = "temporary-dev-secret-change-in-production"
	}
	SecretKey = []byte(secret)
	log.Println("[Auth] JWT密钥已初始化")
}

// Claims JWT令牌声明结构
type Claims struct {
	UserID uint `json:"user_id"`
	jwt.RegisteredClaims
}

// GenerateToken 为指定用户ID生成JWT令牌
// 令牌有效期为24小时
func GenerateToken(userID uint) (string, error) {
	if len(SecretKey) == 0 {
		return "", errors.New("JWT密钥未初始化，请先调用 auth.Init()")
	}
	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(SecretKey)
}

// ValidateToken 验证JWT令牌并返回声明信息
func ValidateToken(tokenString string) (*Claims, error) {
	if len(SecretKey) == 0 {
		return nil, errors.New("JWT密钥未初始化，请先调用 auth.Init()")
	}
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return SecretKey, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}
	return nil, errors.New("无效的令牌")
}
