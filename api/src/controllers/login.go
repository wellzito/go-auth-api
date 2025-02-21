package controllers

import (
	"api/src/autenticacao"
	"api/src/banco"
	"api/src/config"
	"api/src/modelos"
	"api/src/repositorios"
	"api/src/respostas"
	"api/src/seguranca"
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"
)

var ctx = context.Background()

const (
	maxLoginAttempts = 5                // Número máximo de tentativas permitidas
	timeWindow       = 15 * time.Minute // Janela de tempo para resetar tentativas
	blockTime        = 1 * time.Minute  // Tempo de bloqueio após atingir o limite
)

// Login autentica um usuário com controle de tentativas e bloqueio
func Login(w http.ResponseWriter, r *http.Request) {
	corpoRequisicao, erro := ioutil.ReadAll(r.Body)
	if erro != nil {
		respostas.Erro(w, http.StatusUnprocessableEntity, erro)
		return
	}

	var usuario modelos.Usuario
	if erro = json.Unmarshal(corpoRequisicao, &usuario); erro != nil {
		respostas.Erro(w, http.StatusBadRequest, erro)
		return
	}

	// Conectar ao Redis
	rdb := config.RedisClient

	// Definir as chaves para tentativas de login, bloqueio e token
	loginKey := "login_attempts:" + usuario.Email
	blockKey := "login_blocked:" + usuario.Email
	tokenKey := "auth_token:" + usuario.Email
	userDataKey := "user_data:" + usuario.Email // Nova chave para armazenar os dados do usuário (id, nome, etc.)

	// Verificar se o token já está armazenado no Redis
	tokenExistente, err := rdb.Get(ctx, tokenKey).Result()
	if err == nil && tokenExistente != "" {
		// Se o token já existir no Redis, podemos retornar o token sem precisar consultar o banco de dados.
		// Agora, também precisamos verificar se o token está relacionado ao ID correto armazenado no Redis
		userData, err := rdb.Get(ctx, userDataKey).Result()
		if err != nil {
			respostas.Erro(w, http.StatusInternalServerError, errors.New("erro ao recuperar dados do usuário do Redis"))
			return
		}

		// Aqui, extraímos o ID e nome dos dados armazenados no Redis
		var usuarioRedis modelos.Usuario
		erro = json.Unmarshal([]byte(userData), &usuarioRedis)
		if erro != nil {
			respostas.Erro(w, http.StatusInternalServerError, erro)
			return
		}

		// Log de debug indicando que o login foi feito com Redis
		log.Println("Login realizado com sucesso usando Redis.")

		// Retornar o token e ID diretamente se o usuário estiver no Redis
		respostas.JSON(w, http.StatusOK, modelos.DadosAutenticacao{
			ID:    strconv.FormatUint(usuarioRedis.ID, 10), // Retorna o ID do usuário
			Token: tokenExistente,
		})
		return
	}

	// Se o token não existir ou houve erro ao obter do Redis, verifica se o usuário está bloqueado
	blocked, _ := rdb.Get(ctx, blockKey).Result()
	if blocked == "1" {
		respostas.Erro(w, http.StatusTooManyRequests, errors.New("muitas tentativas, tente novamente mais tarde"))
		return
	}

	// Verificar o número de tentativas de login no Redis
	attempts, _ := rdb.Get(ctx, loginKey).Int()
	if attempts >= maxLoginAttempts {
		// Bloquear o usuário por 1 hora no Redis
		rdb.Set(ctx, blockKey, "1", blockTime)
		rdb.Del(ctx, loginKey) // Resetar as tentativas
		respostas.Erro(w, http.StatusTooManyRequests, errors.New("muitas tentativas. Conta bloqueada por 1 minuto"))
		return
	}

	// Conectar ao banco de dados PostgreSQL
	db, erro := banco.Conectar()
	if erro != nil {
		respostas.Erro(w, http.StatusInternalServerError, erro)
		return
	}
	defer db.Close()

	// Buscar o usuário no PostgreSQL, se o Redis não fornecer as informações
	repositorio := repositorios.NovoRepositorioDeUsuarios(db)
	usuarioSalvoNoBanco, erro := repositorio.BuscarPorEmail(usuario.Email)
	if erro != nil {
		respostas.Erro(w, http.StatusInternalServerError, erro)
		return
	}

	// Verificar as credenciais do usuário
	if erro = seguranca.VerificarSenha(usuarioSalvoNoBanco.Senha, usuario.Senha); erro != nil {
		// Incrementar as tentativas de login no Redis
		rdb.Incr(ctx, loginKey)
		rdb.Expire(ctx, loginKey, timeWindow)
		respostas.Erro(w, http.StatusUnauthorized, errors.New("credenciais inválidas"))
		return
	}

	// Log de debug indicando que o login foi feito com PostgreSQL
	log.Println("Login realizado com sucesso usando PostgreSQL.")

	// Gerar o token de autenticação
	token, erro := autenticacao.CriarToken(usuarioSalvoNoBanco.ID)
	if erro != nil {
		respostas.Erro(w, http.StatusInternalServerError, erro)
		return
	}

	// Armazenar o token no Redis (expira após algum tempo, por exemplo, 1 hora)
	rdb.Set(ctx, tokenKey, token, timeWindow)

	// Armazenar os dados do usuário no Redis (ID, nome, etc.)
	usuarioRedis := modelos.Usuario{
		ID:    usuarioSalvoNoBanco.ID,
		Nome:  usuarioSalvoNoBanco.Nome,
		Email: usuarioSalvoNoBanco.Email,
	}

	usuarioRedisData, err := json.Marshal(usuarioRedis)
	if err != nil {
		respostas.Erro(w, http.StatusInternalServerError, err)
		return
	}

	rdb.Set(ctx, userDataKey, usuarioRedisData, timeWindow)

	// Se o login for bem-sucedido, resetar tentativas e bloqueio no Redis
	rdb.Del(ctx, loginKey, blockKey)

	// Converter o ID do usuário para string e retornar o token
	usuarioID := strconv.FormatUint(usuarioSalvoNoBanco.ID, 10)
	respostas.JSON(w, http.StatusOK, modelos.DadosAutenticacao{ID: usuarioID, Token: token})
}

// LoginAnonimo gera um token para um usuário anônimo
func LoginAnonimo(w http.ResponseWriter, r *http.Request) {
	// Conectar ao Redis
	rdb := config.RedisClient

	// Chave única para o login anônimo, pode ser um ID único gerado para o usuário anônimo (UUID, por exemplo)
	anonimoKey := "anonimo_token:" + r.RemoteAddr // Usando o IP ou algum identificador único

	// Tenta buscar o token no Redis para o usuário anônimo
	tokenExistente, err := rdb.Get(ctx, anonimoKey).Result()
	if err == nil && tokenExistente != "" {
		// Se o token já existir no Redis, retorna ele diretamente
		respostas.JSON(w, http.StatusOK, map[string]string{"token": tokenExistente})
		return
	}

	// Caso não exista, cria um novo token anônimo
	token, erro := autenticacao.CriarTokenAnonimo()
	if erro != nil {
		respostas.Erro(w, http.StatusInternalServerError, erro)
		return
	}

	// Armazenar o token no Redis, com tempo de expiração, por exemplo, 15 minutos
	rdb.Set(ctx, anonimoKey, token, timeWindow)

	// Retorna o token gerado
	respostas.JSON(w, http.StatusOK, map[string]string{"token": token})
}
