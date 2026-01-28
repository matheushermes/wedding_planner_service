package auth

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/matheushermes/wedding_planner_service/configs"
)

// Constantes de configuração para tokens
const (
	// TokenExpirationTime define o tempo de vida do token (24h é um bom balanço entre segurança e UX)
	TokenExpirationTime = 24 * time.Hour

	// RefreshTokenExpirationTime define o tempo de vida do refresh token (7 dias)
	RefreshTokenExpirationTime = 7 * 24 * time.Hour

	// TokenType é o tipo do token (Bearer é o padrão OAuth 2.0)
	TokenType = "Bearer"
)

// Erros customizados para melhor tratamento
var (
	ErrTokenMissing         = errors.New("authorization token is missing")
	ErrTokenInvalid         = errors.New("token is invalid or malformed")
	ErrTokenExpired         = errors.New("token has expired")
	ErrTokenNotValidYet     = errors.New("token is not valid yet")
	ErrInvalidSigningMethod = errors.New("invalid token signing method")
	ErrJWTSecretNotSet      = errors.New("JWT_SECRET environment variable not set")
)

// Claims representa as informações estruturadas contidas no token JWT
type Claims struct {
	UserID uint   `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

// CreateToken cria um novo token JWT para o usuário com claims estruturadas
// Retorna o token assinado ou erro caso falhe
func CreateToken(userID uint, email string) (string, error) {
	now := time.Now()
	expirationTime := now.Add(TokenExpirationTime)

	claims := &Claims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime), // Tempo de expiração
			IssuedAt:  jwt.NewNumericDate(now),            // Data de emissão (importante para auditoria)
			NotBefore: jwt.NewNumericDate(now),            // Token não pode ser usado antes desta data
			Issuer:    "wedding_planner_service",          // Identifica o emissor (importante em microserviços)
			Subject:   fmt.Sprintf("%d", userID),          // Subject identifica o usuário
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(configs.JWT_SECRET))
}

// ExtractToken extrai o token JWT do header Authorization
func ExtractToken(c *gin.Context) string {
	authHeader := c.GetHeader("Authorization")

	if len(authHeader) > 7 && strings.HasPrefix(authHeader, "Bearer ") {
		return strings.TrimSpace(authHeader[7:]) // Remove "Bearer " e espaços extras
	}

	// SEGURANÇA ADICIONAL: Também verifica query param (útil para WebSockets)
	// Comentado por padrão, descomentar se necessário:
	// if token := c.Query("token"); token != "" {
	// 	return token
	// }

	return ""
}

// returnVerificationKey retorna a chave para verificação do token
func returnVerificationKey(token *jwt.Token) (interface{}, error) {
	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
		return nil, fmt.Errorf("%w: %v", ErrInvalidSigningMethod, token.Header["alg"])
	}

	return configs.JWT_SECRET, nil
}

// VerifyToken verifica se o token JWT é válido
// Valida assinatura, expiração e estrutura do token
func VerifyToken(c *gin.Context) error {
	tokenString := ExtractToken(c)
	if tokenString == "" {
		return ErrTokenMissing
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, returnVerificationKey)
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return ErrTokenExpired
		}
		if errors.Is(err, jwt.ErrTokenNotValidYet) {
			return ErrTokenNotValidYet
		}
		return fmt.Errorf("%w: %v", ErrTokenInvalid, err)
	}

	// Verifica se o token está válido (assinatura correta, não expirado, etc)
	if !token.Valid {
		return ErrTokenInvalid
	}

	return nil
}

// ExtractUserID extrai o ID do usuário do token de forma type-safe
func ExtractUserID(c *gin.Context) (uint, error) {
	tokenString := ExtractToken(c)
	if tokenString == "" {
		return 0, ErrTokenMissing
	}

	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, returnVerificationKey)
	if err != nil || !token.Valid {
		return 0, ErrTokenInvalid
	}

	return claims.UserID, nil
}

// ExtractTokenMetadata extrai todas as informações (metadata) do token
func ExtractTokenMetadata(c *gin.Context) (*Claims, error) {
	tokenString := ExtractToken(c)
	if tokenString == "" {
		return nil, ErrTokenMissing
	}

	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, returnVerificationKey)
	if err != nil || !token.Valid {
		return nil, ErrTokenInvalid
	}

	return claims, nil
}

// AuthMiddleware é um middleware para proteger rotas que requerem autenticação
// USO: router.Use(auth.AuthMiddleware())
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Verifica o token
		if err := VerifyToken(c); err != nil {
			c.JSON(401, gin.H{
				"error": err.Error(),
			})
			c.Abort() // Impede que handlers subsequentes sejam executados
			return
		}

		// PERFORMANCE: Extrai e armazena o userID no contexto para evitar re-parsing
		// Handlers podem pegar com: userID := c.GetUint("user_id")
		userID, err := ExtractUserID(c)
		if err != nil {
			c.JSON(401, gin.H{
				"error": "failed to extract user information",
			})
			c.Abort()
			return
		}

		// Armazena o user_id no contexto para uso nos handlers
		c.Set("user_id", userID)

		// Continua para o próximo handler
		c.Next()
	}
}
