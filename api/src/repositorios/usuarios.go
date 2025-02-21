package repositorios

import (
	"api/src/modelos"
	"database/sql"
	"fmt"
)

// Usuarios representa um repositório de usuarios
type Usuarios struct {
	db *sql.DB
}

// NovoRepositorioDeUsuarios cria um repositório de usuários
func NovoRepositorioDeUsuarios(db *sql.DB) *Usuarios {
	return &Usuarios{db}
}

// Criar insere um usuário no banco de dados
func (repositorio Usuarios) Criar(usuario modelos.Usuario) (uint64, error) {
	statement, erro := repositorio.db.Prepare(
		"insert into usuarios (nome, nick, email, senha) values($1, $2, $3, $4) returning id", // Alterado para PostgreSQL com RETURNING
	)
	if erro != nil {
		return 0, erro
	}
	defer statement.Close()

	// A consulta agora retorna o ID gerado, que é capturado na variável resultado
	var id uint64
	err := statement.QueryRow(usuario.Nome, usuario.Nick, usuario.Email, usuario.Senha).Scan(&id)
	if err != nil {
		return 0, err
	}

	// Retorna o ID gerado pelo INSERT
	return id, nil
}

func (repositorio Usuarios) Buscar(nomeOuNick string) ([]modelos.Usuario, error) {
	// Limitar o tamanho do input para evitar DoS com consultas excessivas
	if len(nomeOuNick) > 100 {
		return nil, fmt.Errorf("termo de busca muito longo")
	}

	// Sanitizar e preparar o filtro de busca
	// Evita que caracteres como '%' ou '_' causem problemas no LIKE
	nomeOuNick = "%" + nomeOuNick + "%"

	// Query preparada, evitando SQL Injection
	linhas, erro := repositorio.db.Query(
		"select id, nome, nick, email, criadoEm from usuarios where nome ILIKE $1 or nick ILIKE $1", // ILIKE para busca sem case-sensitive
		nomeOuNick,
	)

	if erro != nil {
		return nil, erro
	}
	defer linhas.Close()

	var usuarios []modelos.Usuario

	// Iterar sobre os resultados e preencher a lista de usuários
	for linhas.Next() {
		var usuario modelos.Usuario

		// Scan para mapear os resultados da consulta
		if erro = linhas.Scan(
			&usuario.ID,
			&usuario.Nome,
			&usuario.Nick,
			&usuario.Email,
			&usuario.CriadoEm,
		); erro != nil {
			return nil, erro
		}

		usuarios = append(usuarios, usuario)
	}

	// Retornar os usuários encontrados
	return usuarios, nil
}

// BuscarPorID traz um usuário do banco de dados
func (repositorio Usuarios) BuscarPorID(ID uint64) (modelos.Usuario, error) {
	var usuario modelos.Usuario

	// Usar QueryRow para buscar um único usuário
	linha := repositorio.db.QueryRow(
		"select id, nome, nick, email, criadoEm from usuarios where id = $1", // Alterado para PostgreSQL
		ID,
	)

	// Scan para mapear os resultados
	err := linha.Scan(
		&usuario.ID,
		&usuario.Nome,
		&usuario.Nick,
		&usuario.Email,
		&usuario.CriadoEm,
	)

	// Se não encontrar o usuário (não existe linha), retornar um erro mais claro
	if err == sql.ErrNoRows {
		return modelos.Usuario{}, fmt.Errorf("usuário com ID %d não encontrado", ID)
	}

	// Se ocorrer outro erro, retorna ele
	if err != nil {
		return modelos.Usuario{}, err
	}

	// Retorna o usuário encontrado
	return usuario, nil
}

// BuscarPorTermo busca usuários cujo nome ou nick correspondem ao termo de busca
func (repositorio Usuarios) BuscarPorTermo(termoBusca string, usuarioID uint64) ([]modelos.Usuario, error) {
	// Limitar o tamanho do input para evitar DoS com consultas excessivas
	if len(termoBusca) > 100 {
		return nil, fmt.Errorf("termo de busca muito longo")
	}

	// Sanitizar e preparar o filtro de busca
	termoBusca = "%" + termoBusca + "%" // Adiciona % para buscar qualquer parte da string

	// Query com ILIKE para fazer a busca insensível a maiúsculas e minúsculas
	linhas, erro := repositorio.db.Query(
		"select id, nome, nick, email, criadoEm from usuarios where nome ILIKE $1 or nick ILIKE $1",
		termoBusca,
	)
	if erro != nil {
		return nil, erro
	}
	defer linhas.Close()

	var usuarios []modelos.Usuario

	// Itera sobre os resultados e preenche a lista de usuários
	for linhas.Next() {
		var usuario modelos.Usuario

		// Scan para mapear os resultados da consulta
		if erro = linhas.Scan(
			&usuario.ID,
			&usuario.Nome,
			&usuario.Nick,
			&usuario.Email,
			&usuario.CriadoEm,
		); erro != nil {
			return nil, erro
		}

		// Verifica se o usuário encontrado é o mesmo que está buscando
		if usuario.ID == usuarioID {
			continue // Ignora o usuário que está fazendo a busca
		}

		// Verifica a conexão entre os usuários
		var segueVoce, voceSegue bool
		repoErro := repositorio.db.QueryRow(
			`SELECT EXISTS(SELECT 1 FROM seguidores WHERE seguidor_id = $1 AND usuario_id = $2)`, usuario.ID, usuarioID,
		).Scan(&segueVoce)
		if repoErro != nil {
			return nil, repoErro
		}

		repoErro = repositorio.db.QueryRow(
			`SELECT EXISTS(SELECT 1 FROM seguidores WHERE seguidor_id = $1 AND usuario_id = $2)`, usuarioID, usuario.ID,
		).Scan(&voceSegue)
		if repoErro != nil {
			return nil, repoErro
		}

		// Define a conexão
		if segueVoce && voceSegue {
			usuario.Conexao = "se seguem"
		} else if segueVoce {
			usuario.Conexao = "segue você"
		} else if voceSegue {
			usuario.Conexao = "seguindo"
		} else {
			usuario.Conexao = "nenhuma"
		}

		usuarios = append(usuarios, usuario)
	}

	// Retornar a lista de usuários encontrados (ou lista vazia se não encontrar)
	return usuarios, nil
}

// Atualizar altera as informações de um usuário no banco de dados
func (repositorio Usuarios) Atualizar(ID uint64, usuario modelos.Usuario) error {
	statement, erro := repositorio.db.Prepare(
		"update usuarios set nome = $1, nick = $2, email = $3 where id = $4",
	)
	if erro != nil {
		return erro
	}
	defer statement.Close()

	if _, erro = statement.Exec(usuario.Nome, usuario.Nick, usuario.Email, ID); erro != nil {
		return erro
	}

	return nil
}

// Deletar exclui as informações de um usuário no banco de dados
func (repositorio Usuarios) Deletar(ID uint64) error {
	statement, erro := repositorio.db.Prepare("delete from usuarios where id = $1")
	if erro != nil {
		return erro
	}
	defer statement.Close()

	if _, erro = statement.Exec(ID); erro != nil {
		return erro
	}

	return nil
}

// BuscarPorEmail busca um usuário por email e retorna o seu id e senha com hash
func (repositorio Usuarios) BuscarPorEmail(email string) (modelos.Usuario, error) {
	var usuario modelos.Usuario

	// Usar QueryRow para otimizar e buscar apenas um resultado
	linha := repositorio.db.QueryRow("select id, senha from usuarios where email = $1", email)

	// Verifica se houve erro durante o Scan ou se não foi encontrado nenhum usuário
	if err := linha.Scan(&usuario.ID, &usuario.Senha); err != nil {
		if err == sql.ErrNoRows {
			// Se não encontrar o usuário, retornar um erro específico
			return modelos.Usuario{}, fmt.Errorf("usuário com email %s não encontrado", email)
		}
		// Retorna outros erros que possam ocorrer durante o Scan
		return modelos.Usuario{}, err
	}

	// Retorna o usuário encontrado
	return usuario, nil
}

// Seguir permite que um usuário siga outro e retorna o usuário seguido com a conexão atualizada
func (repositorio Usuarios) Seguir(usuarioID, seguidorID uint64) ([]modelos.Usuario, error) {
	// Validar se o usuário não está tentando seguir a si mesmo
	if usuarioID == seguidorID {
		return nil, fmt.Errorf("um usuário não pode seguir a si mesmo")
	}

	// Iniciar uma transação
	tx, erro := repositorio.db.Begin()
	if erro != nil {
		return nil, fmt.Errorf("erro ao iniciar transação: %v", erro)
	}
	defer tx.Rollback()

	// Verificar se o seguidor já segue o usuário
	var count int
	erro = tx.QueryRow("SELECT COUNT(*) FROM seguindo WHERE usuario_id = $1 AND seguindo_id = $2", usuarioID, seguidorID).Scan(&count)
	if erro != nil {
		return nil, fmt.Errorf("erro ao verificar se já segue: %v", erro)
	}

	if count > 0 {
		return nil, fmt.Errorf("o usuário %d já segue o usuário %d", seguidorID, usuarioID)
	}

	// Inserir na tabela "seguindo"
	_, erro = tx.Exec(
		"INSERT INTO seguindo (usuario_id, seguindo_id) VALUES ($1, $2)",
		usuarioID, seguidorID,
	)
	if erro != nil {
		return nil, fmt.Errorf("erro ao inserir na tabela 'seguindo': %v", erro)
	}

	// Inserir na tabela "seguidores"
	_, erro = tx.Exec(
		"INSERT INTO seguidores (usuario_id, seguidor_id) VALUES ($1, $2)",
		usuarioID, seguidorID,
	)
	if erro != nil {
		return nil, fmt.Errorf("erro ao inserir na tabela 'seguidores': %v", erro)
	}

	// Confirmar a transação
	if erro = tx.Commit(); erro != nil {
		return nil, fmt.Errorf("erro ao confirmar transação: %v", erro)
	}

	// Buscar os dados do usuário seguido
	linha := repositorio.db.QueryRow(`SELECT id, nome, nick, email, criadoEm FROM usuarios WHERE id = $1`, usuarioID)

	var usuario modelos.Usuario
	if erro = linha.Scan(&usuario.ID, &usuario.Nome, &usuario.Nick, &usuario.Email, &usuario.CriadoEm); erro != nil {
		return nil, fmt.Errorf("erro ao buscar dados do usuário seguido: %v", erro)
	}

	// **Verifica a conexão entre os usuários APÓS a transação**
	var segueVoce, voceSegue bool
	repoErro := repositorio.db.QueryRow(
		`SELECT EXISTS(SELECT 1 FROM seguidores WHERE seguidor_id = $1 AND usuario_id = $2)`, usuario.ID, seguidorID,
	).Scan(&segueVoce)
	if repoErro != nil {
		return nil, repoErro
	}

	repoErro = repositorio.db.QueryRow(
		`SELECT EXISTS(SELECT 1 FROM seguidores WHERE seguidor_id = $1 AND usuario_id = $2)`, seguidorID, usuario.ID,
	).Scan(&voceSegue)
	if repoErro != nil {
		return nil, repoErro
	}

	// Atualiza a conexão do usuário
	if segueVoce && voceSegue {
		usuario.Conexao = "se seguem"
	} else if segueVoce {
		usuario.Conexao = "segue você"
	} else if voceSegue {
		usuario.Conexao = "seguindo"
	} else {
		usuario.Conexao = "nenhuma"
	}

	// Retorna o usuário seguido com a conexão atualizada
	return []modelos.Usuario{usuario}, nil
}

// PararDeSeguir permite que um usuário pare de seguir outro
func (repositorio Usuarios) PararDeSeguir(usuarioID, seguidorID uint64) ([]modelos.Usuario, error) {
	// Validar se o seguidor não está tentando parar de seguir a si mesmo
	if usuarioID == seguidorID {
		return nil, fmt.Errorf("um usuário não pode parar de seguir a si mesmo")
	}

	// Iniciar uma transação
	tx, erro := repositorio.db.Begin()
	if erro != nil {
		return nil, fmt.Errorf("erro ao iniciar transação: %v", erro)
	}
	defer tx.Rollback()

	// Verificar se a relação já existe antes de tentar excluir
	var count int
	erro = tx.QueryRow("SELECT COUNT(*) FROM seguindo WHERE usuario_id = $1 AND seguindo_id = $2", usuarioID, seguidorID).Scan(&count)
	if erro != nil {
		return nil, fmt.Errorf("erro ao verificar se segue: %v", erro)
	}

	if count == 0 {
		// Se o seguidor não segue o usuário, retornar uma mensagem informando
		return nil, fmt.Errorf("o usuário %d não segue o usuário %d", seguidorID, usuarioID)
	}

	// Remover da tabela "seguindo" (indica que seguidorID não segue mais usuarioID)
	_, erro = tx.Exec(
		"DELETE FROM seguindo WHERE usuario_id = $1 AND seguindo_id = $2",
		usuarioID, seguidorID,
	)
	if erro != nil {
		return nil, fmt.Errorf("erro ao remover da tabela 'seguindo': %v", erro)
	}

	// Remover da tabela "seguidores" (indica que usuarioID não tem mais seguidorID como seguidor)
	_, erro = tx.Exec(
		"DELETE FROM seguidores WHERE usuario_id = $1 AND seguidor_id = $2",
		usuarioID, seguidorID,
	)
	if erro != nil {
		return nil, fmt.Errorf("erro ao remover da tabela 'seguidores': %v", erro)
	}

	// Confirmar a transação
	if erro = tx.Commit(); erro != nil {
		return nil, fmt.Errorf("erro ao confirmar transação: %v", erro)
	}

	// Buscar os dados do usuário seguido
	linha := repositorio.db.QueryRow(`SELECT id, nome, nick, email, criadoEm FROM usuarios WHERE id = $1`, usuarioID)

	var usuario modelos.Usuario
	if erro = linha.Scan(&usuario.ID, &usuario.Nome, &usuario.Nick, &usuario.Email, &usuario.CriadoEm); erro != nil {
		return nil, fmt.Errorf("erro ao buscar dados do usuário seguido: %v", erro)
	}

	// **Verifica a conexão entre os usuários APÓS a transação**
	var segueVoce, voceSegue bool
	repoErro := repositorio.db.QueryRow(
		`SELECT EXISTS(SELECT 1 FROM seguidores WHERE seguidor_id = $1 AND usuario_id = $2)`, usuario.ID, seguidorID,
	).Scan(&segueVoce)
	if repoErro != nil {
		return nil, repoErro
	}

	repoErro = repositorio.db.QueryRow(
		`SELECT EXISTS(SELECT 1 FROM seguidores WHERE seguidor_id = $1 AND usuario_id = $2)`, seguidorID, usuario.ID,
	).Scan(&voceSegue)
	if repoErro != nil {
		return nil, repoErro
	}

	// Atualiza a conexão do usuário
	if segueVoce && voceSegue {
		usuario.Conexao = "se seguem"
	} else if segueVoce {
		usuario.Conexao = "segue você"
	} else if voceSegue {
		usuario.Conexao = "seguindo"
	} else {
		usuario.Conexao = "nenhuma"
	}

	return []modelos.Usuario{usuario}, nil
}

// BuscarSeguidores traz todos os seguidores de um usuário
func (repositorio Usuarios) BuscarSeguidores(usuarioID uint64) ([]modelos.Usuario, error) {
	// Buscar seguidores do usuário (quem está seguindo o usuário)
	linhas, erro := repositorio.db.Query(`
		SELECT u.id, u.nome, u.nick, u.email, u.criadoEm
		FROM usuarios u 
		INNER JOIN seguidores s ON u.id = s.seguidor_id 
		WHERE s.usuario_id = $1`, usuarioID,
	)
	if erro != nil {
		fmt.Println("Erro ao executar query:", erro)
		return nil, erro
	}
	defer linhas.Close()

	var usuarios []modelos.Usuario
	for linhas.Next() {
		var usuario modelos.Usuario

		if erro = linhas.Scan(
			&usuario.ID,
			&usuario.Nome,
			&usuario.Nick,
			&usuario.Email,
			&usuario.CriadoEm,
		); erro != nil {
			return nil, erro
		}

		usuarios = append(usuarios, usuario)
	}

	return usuarios, nil
}

// BuscarSeguindo traz todos os usuários que um determinado usuário está seguindo
func (repositorio Usuarios) BuscarSeguindo(usuarioID uint64) ([]modelos.Usuario, error) {
	linhas, erro := repositorio.db.Query(`
		SELECT u.id, u.nome, u.nick, u.email, u.criadoEm
		FROM usuarios u 
		INNER JOIN seguindo s ON u.id = s.usuario_id 
		WHERE s.seguindo_id = $1`, usuarioID,
	)
	if erro != nil {
		return nil, erro
	}
	defer linhas.Close()

	var usuarios []modelos.Usuario

	for linhas.Next() {
		var usuario modelos.Usuario

		if erro = linhas.Scan(
			&usuario.ID,
			&usuario.Nome,
			&usuario.Nick,
			&usuario.Email,
			&usuario.CriadoEm,
		); erro != nil {
			return nil, erro
		}

		usuarios = append(usuarios, usuario)
	}

	return usuarios, nil
}

// QuantidadeSeguindo retorna a quantidade de usuários que um usuário está seguindo
func (repositorio Usuarios) QuantidadeSeguindo(usuarioID uint64) (int, error) {
	// Contar o número de pessoas que o usuário está seguindo
	var quantidadeSeguindo int
	erro := repositorio.db.QueryRow(`
		SELECT COUNT(*)
		FROM seguindo s
		WHERE s.seguindo_id = $1`, usuarioID).Scan(&quantidadeSeguindo)

	if erro != nil {
		fmt.Println("Erro ao executar query:", erro)
		return 0, erro
	}

	return quantidadeSeguindo, nil
}

// QuantidadeSeguidores retorna a quantidade de seguidores de um usuário
func (repositorio Usuarios) QuantidadeSeguidores(usuarioID uint64) (int, error) {
	// Contar o número de seguidores de um usuário
	var quantidadeSeguidores int
	erro := repositorio.db.QueryRow(`
		SELECT COUNT(*)
		FROM seguidores s
		WHERE s.usuario_id = $1`, usuarioID).Scan(&quantidadeSeguidores)

	if erro != nil {
		fmt.Println("Erro ao executar query:", erro)
		return 0, erro
	}

	return quantidadeSeguidores, nil
}

// BuscarSenha traz a senha de um usuário pelo ID
func (repositorio Usuarios) BuscarSenha(usuarioID uint64) (string, error) {
	// Consultar a senha do usuário pelo ID
	linha, erro := repositorio.db.Query("select senha from usuarios where id = $1", usuarioID)
	if erro != nil {
		return "", fmt.Errorf("erro ao executar consulta: %v", erro)
	}
	defer linha.Close()

	var senha string

	// Verificar se a consulta retornou resultados
	if linha.Next() {
		// Captura a senha (que deve estar hashada) do usuário
		if erro = linha.Scan(&senha); erro != nil {
			return "", fmt.Errorf("erro ao ler senha: %v", erro)
		}
	} else {
		// Caso o usuário não seja encontrado
		return "", fmt.Errorf("usuário com ID %d não encontrado", usuarioID)
	}

	// Retorna a senha (hash)
	return senha, nil
}

// AtualizarSenha altera a senha de um usuário no banco de dados
func (repositorio Usuarios) AtualizarSenha(usuarioID uint64, senha string) error {
	// Preparar a query de atualização
	statement, erro := repositorio.db.Prepare("update usuarios set senha = $1 where id = $2")
	if erro != nil {
		return fmt.Errorf("erro ao preparar consulta para atualização de senha: %v", erro)
	}
	defer statement.Close()

	// Executar a atualização da senha
	if _, erro = statement.Exec(senha, usuarioID); erro != nil {
		return fmt.Errorf("erro ao executar atualização de senha: %v", erro)
	}

	return nil
}

// NovaSenha altera a senha de um usuário no banco de dados
func (repositorio Usuarios) NovaSenha(usuarioID uint64, senha string) error {
	// Preparar a query de atualização
	statement, erro := repositorio.db.Prepare("update usuarios set senha = $1 where id = $2")
	if erro != nil {
		return fmt.Errorf("erro ao preparar consulta para atualização de senha: %v", erro)
	}
	defer statement.Close()

	// Executar a atualização da senha
	if _, erro = statement.Exec(senha, usuarioID); erro != nil {
		return fmt.Errorf("erro ao executar atualização de senha: %v", erro)
	}

	return nil
}
