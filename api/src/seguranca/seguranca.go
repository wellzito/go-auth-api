package seguranca

import (
	"crypto/rand"
	"fmt"
	"log"

	"golang.org/x/crypto/bcrypt"
)

// Hash recebe uma string e coloca um hash nela
func Hash(senha string) ([]byte, error) {
	log.Println("Iniciando o processo de hash para a senha.")

	// Gerar o hash da senha
	hash, err := bcrypt.GenerateFromPassword([]byte(senha), bcrypt.DefaultCost)
	if err != nil {
		log.Println("Erro ao gerar hash da senha:", err)
		return nil, err
	}

	log.Println("Hash gerado com sucesso.")
	return hash, nil
}

// VerificarSenha compara uma senha e um hash e retorna se elas são iguais
func VerificarSenha(senhaComHash, senhaString string) error {
	log.Println("Iniciando a verificação da senha.")

	// Comparar o hash da senha com o valor passado
	err := bcrypt.CompareHashAndPassword([]byte(senhaComHash), []byte(senhaString))
	if err != nil {
		log.Println("Erro ao comparar a senha com o hash:", err)
		return err
	}

	log.Println("Senha verificada com sucesso.")
	return nil
}

// GerarCodigo gera um código de 4 dígitos
func GerarCodigo() (string, error) {
	bytes := make([]byte, 2) // 2 bytes são suficientes para gerar um número de 4 dígitos
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}

	codigo := fmt.Sprintf("%04d", (int(bytes[0])<<8+int(bytes[1]))%10000)
	return codigo, nil
}

// HashCodigo criptografa o código de recuperação
func HashCodigo(codigo string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(codigo), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// VerificarCodigo compara um código digitado com o hash salvo e retorna se são iguais
func VerificarCodigo(hash, codigo string) (bool, error) {
	// Tenta comparar o código com o hash salvo
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(codigo))
	if err != nil {
		// Se houver erro na comparação, retorna falso e o erro
		return false, fmt.Errorf("código de recuperação inválido")
	}
	// Retorna verdadeiro caso o código seja válido
	return true, nil
}
