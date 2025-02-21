package autenticacao

import (
	"api/src/config"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
)

// CriarToken retorna um token assinado com as permissões do usuário
func CriarToken(usuarioID uint64) (string, error) {
	permissoes := jwt.MapClaims{}
	permissoes["authorized"] = true
	permissoes["exp"] = time.Now().Add(time.Hour * 6).Unix()
	permissoes["usuarioId"] = usuarioID
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, permissoes)
	return token.SignedString([]byte(config.SecretKey))
}

// ValidarToken verifica se o token passado na requisição é válido e retorna se é anônimo ou não
func ValidarToken(r *http.Request) (bool, error) {
	tokenString := extrairToken(r)
	token, erro := jwt.Parse(tokenString, retornarChaveDeVerificacao)
	if erro != nil {
		return false, erro
	}

	permissoes, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return false, errors.New("token inválido")
	}

	// Se o token tem "anonimo" = true, ele é um login anônimo
	if _, anonimo := permissoes["anonimo"].(bool); anonimo {
		return true, nil
	}

	return false, nil
}

// ExtrairUsuarioID retorna o usuarioId ou 0 se for anônimo
func ExtrairUsuarioID(r *http.Request) (uint64, error) {
	tokenString := extrairToken(r)
	log.Printf("Token ok") // Log do token JWT

	token, erro := jwt.Parse(tokenString, retornarChaveDeVerificacao)
	if erro != nil {
		log.Printf("Erro ao parsear token: %v", erro)
		return 0, erro
	}

	permissoes, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		log.Printf("Token inválido")
		return 0, errors.New("token inválido")
	}

	log.Printf("Permissões extraídas: %+v", permissoes) // Log completo das permissões

	// Verifica se o token é anônimo
	if _, anonimo := permissoes["anonimo"].(bool); anonimo {
		log.Printf("Usuário anônimo")
		return 0, nil
	}

	// Verifica se usuarioId está presente no token
	valor, ok := permissoes["usuarioId"].(float64)
	if !ok {
		log.Printf("Erro: usuarioId não encontrado ou em formato inválido")
		return 0, errors.New("usuarioId ausente ou inválido no token")
	}

	usuarioID := uint64(valor)
	log.Printf("usuarioId extraído: %d", usuarioID) // Log do usuarioID extraído

	return usuarioID, nil
}

// ExtrairUsuarioIDComTokenString retorna o usuarioId que está salvo no token
func ExtrairUsuarioIDComTokenString(tokenString string) (uint64, error) {
	token, erro := jwt.Parse(tokenString, retornarChaveDeVerificacao)
	if erro != nil {
		return 0, erro
	}

	if permissoes, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		usuarioID, erro := strconv.ParseUint(fmt.Sprintf("%.0f", permissoes["usuarioId"]), 10, 64)
		if erro != nil {
			return 0, erro
		}

		return usuarioID, nil
	}

	return 0, errors.New("token inválido")
}

func extrairToken(r *http.Request) string {
	token := r.Header.Get("Authorization")
	log.Printf("Token extraído com sucesso") // Log do token extraído
	if len(strings.Split(token, " ")) == 2 {
		return strings.Split(token, " ")[1]
	}

	return ""
}

func retornarChaveDeVerificacao(token *jwt.Token) (interface{}, error) {
	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
		return nil, fmt.Errorf("método de assinatura inesperado! %v", token.Header["alg"])
	}

	return config.SecretKey, nil
}

// ValidarTokenComTokenString valida o token JWT passado diretamente como string
func ValidarTokenComTokenString(tokenString string) error {
	token, erro := jwt.Parse(tokenString, retornarChaveDeVerificacao)
	if erro != nil {
		return erro
	}

	if _, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return nil
	}

	return errors.New("token inválido")
}

// CriarTokenAnonimo gera um token para usuários anônimos
func CriarTokenAnonimo() (string, error) {
	permissoes := jwt.MapClaims{}
	permissoes["exp"] = time.Now().Add(time.Hour * 24).Unix() // Expira em 24h
	permissoes["anonimo"] = true                              // Define como usuário anônimo

	// Criar token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, permissoes)
	return token.SignedString([]byte(config.SecretKey))
}
