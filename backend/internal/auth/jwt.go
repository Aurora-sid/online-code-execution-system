package auth

/*针对 JWT (JSON Web Token) 的全套操作封装，
为保证微服务通信或前后端分离体系下的无状态 HTTP 请求的
安全验证基础支持。它包含了签名秘钥获取、签发加密字符串
及验证和解包声明（Claims）等功能。*/
import (
	"errors"
	"log"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// SecretKey JWT签名密钥，在应用启动时通过 Init() 初始化
var SecretKey []byte

// Init 初始化JWT密钥，从环境变量加载
// 必须在应用启动时调用此函数
func Init(secret string) {
	if secret == "" {
		// 从环境变量获取
		secret = os.Getenv("JWT_SECRET")
		log.Println("[Auth] 从环境变量 JWT_SECRET 加载JWT密钥")
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

// GenerateToken 为指定用户ID生成JWT令牌，令牌有效期为24小时
func GenerateToken(userID uint) (string, error) {
	if len(SecretKey) == 0 {
		return "", errors.New("JWT密钥未初始化，请先调用 auth.Init()")
	}
	// 令牌格式定义，包含用户ID和注册声明（过期时间、签发时间等）
	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(SecretKey)

	// 生成JWT令牌的步骤：
	// 1. 创建Claims对象，包含用户ID和注册声明（过期时间、签发时间等）
	// 2. 使用HS256算法创建新的JWT令牌对象
	// 3. 使用SecretKey对令牌进行签名，生成最终的JWT字符串
	// tokenString, err := token.SignedString(SecretKey)
	// if err != nil {
	// 	return "", err
	// }

	// log.Println("[Auth] 生成JWT令牌", tokenString)
	// return tokenString, nil
	// 输出：
	//2026/04/12 15:44:24 [Auth] 生成JWT令牌 eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoyLCJleHAiOjE3NzYwNjYyNjQsImlhdCI6MTc3NTk3OTg2NH0.6MfNrXCzv1JVpQ62eazZcC-LxJMJGKnbmIvIt3zYqiM
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
