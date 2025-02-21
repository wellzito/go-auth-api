package config

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv" //remover se tiver em produção
)

var (
	DB *pgxpool.Pool

	// StringConexaoBanco é a string de conexão com o PostgreSQL
	StringConexaoBanco = ""

	// Porta onde a API vai estar rodando
	Porta = 8080

	// SecretKey é a chave que vai ser usada para assinar o token
	SecretKey []byte

	// Pool de conexões com o banco de dados
)

// Carregar vai inicializar as variáveis de ambiente
func Carregar() {
	var erro error

	//Descomente para usar local e comente para usar em prod
	// Carregar as variáveis de ambiente primeiro
	//Remover função abaixo se tiver em produção
	if erro = godotenv.Load(); erro != nil {
		log.Fatal(erro)
	}

	Porta, erro = strconv.Atoi(os.Getenv("APP_PORT"))
	if erro != nil {
		Porta = 9000
	}

	/*fmt.Println("APP_PORT:", os.Getenv("APP_PORT"))
	fmt.Println("DB_USUARIO:", os.Getenv("DB_USUARIO"))
	fmt.Println("DB_SENHA:", os.Getenv("DB_SENHA"))*/

	// Alterar a string de conexão para PostgreSQL
	StringConexaoBanco = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", //altere o sslmode para sslmode=require se estiver em produção,
		os.Getenv("DB_USUARIO"),
		os.Getenv("DB_SENHA"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NOME"),
	)
	log.Println(StringConexaoBanco)
	SecretKey = []byte(os.Getenv("SECRET_KEY"))

	// Conecta ao banco de dados usando pgxpool
	DB, erro = pgxpool.New(context.Background(), StringConexaoBanco)
	if erro != nil {
		log.Fatalf("Erro ao conectar ao banco de dados: %v", erro)
	}

	// Verifica se a conexão foi estabelecida com sucesso
	conn, err := DB.Acquire(context.Background())
	if err != nil {
		log.Fatalf("Erro ao adquirir conexão do pool: %v", err)
	}
	conn.Release()

	log.Println("Conexão com o banco de dados estabelecida com sucesso!")

	verificarBanco()
}

// Função para verificar a conexão e a existência das tabelas
// Função para verificar a conexão e a existência das tabelas
func verificarBanco() {
	// Estabelecer a conexão com o banco de dados PostgreSQL
	db, err := sql.Open("postgres", StringConexaoBanco)
	if err != nil {
		log.Fatalf("Erro ao conectar no banco de dados: %v\n", err)
	}
	defer db.Close()

	// Comandos para verificar as tabelas
	tabelas := []string{"usuarios", "seguidores", "publicacoes", "seguindo", "curtidas", "codigos", "eventos", "bilhetes", "lotes_evento", "mensagens_privadas", "mensagens_grupo"}

	// Itera sobre as tabelas e verifica se existem
	for _, tabela := range tabelas {
		var resultado sql.NullString // Use NullString para lidar com valores NULL
		err = db.QueryRow(fmt.Sprintf("SELECT to_regclass('public.%s')", tabela)).Scan(&resultado)
		if err != nil || !resultado.Valid || resultado.String == "" {
			log.Printf("A tabela '%s' não existe ou ocorreu um erro: %v\n", tabela, err)
			log.Printf("Criando a tabela '%s'...\n", tabela)
			// Chama a função para criar a tabela
			criarTabela(db, tabela)
		} else {
			log.Printf("Tabela '%s' existe.\n", tabela)
		}
	}
}

// Função para criar as tabelas caso elas não existam
func criarTabela(db *sql.DB, tabela string) {
	// Consultar os comandos SQL para as tabelas
	queries := obterComandosSQL(tabela)

	for _, query := range queries {
		_, err := db.Exec(query)
		if err != nil {
			log.Fatalf("Erro ao executar a criação da tabela '%s': %v\n", tabela, err)
		}
		log.Printf("Tabela '%s' criada com sucesso.\n", tabela)
	}
}

// Função que retorna os comandos SQL para a criação das tabelas
func obterComandosSQL(tabela string) []string {
	switch tabela {
	case "usuarios":
		return []string{
			`CREATE TABLE IF NOT EXISTS usuarios (
				id serial PRIMARY KEY,
				nome varchar(50) NOT NULL,
				nick varchar(50) NOT NULL UNIQUE,
				email varchar(50) NOT NULL UNIQUE,
				senha varchar(100) NOT NULL,
				criadoEm timestamp default current_timestamp
			);`,
		}
	case "seguidores":
		return []string{
			`CREATE TABLE IF NOT EXISTS seguidores (
				usuario_id int NOT NULL,
				seguidor_id int NOT NULL,
				PRIMARY KEY(usuario_id, seguidor_id),
				FOREIGN KEY (usuario_id) REFERENCES usuarios(id) ON DELETE CASCADE,
				FOREIGN KEY (seguidor_id) REFERENCES usuarios(id) ON DELETE CASCADE,
				criadoEm timestamp default current_timestamp
			);`,
		}
	case "seguindo":
		return []string{
			`CREATE TABLE IF NOT EXISTS seguindo (
				usuario_id int NOT NULL,
				seguindo_id int NOT NULL,
				PRIMARY KEY(usuario_id, seguindo_id),
				FOREIGN KEY (usuario_id) REFERENCES usuarios(id) ON DELETE CASCADE,
				FOREIGN KEY (seguindo_id) REFERENCES usuarios(id) ON DELETE CASCADE,
				criadoEm timestamp default current_timestamp
			);`,
		}
	case "publicacoes":
		return []string{
			`CREATE TABLE IF NOT EXISTS publicacoes (
				id serial PRIMARY KEY,
				titulo varchar(50) NOT NULL,
				conteudo varchar(300) NOT NULL,
				autor_id int NOT NULL,
				FOREIGN KEY (autor_id) REFERENCES usuarios(id) ON DELETE CASCADE,
				curtidas int DEFAULT 0,
				criadaEm timestamp default current_timestamp
			);`,
		}
	case "curtidas":
		return []string{
			`CREATE TABLE IF NOT EXISTS curtidas (
				usuario_id int NOT NULL,
				publicacao_id int NOT NULL,
				PRIMARY KEY(usuario_id, publicacao_id),
				FOREIGN KEY (usuario_id) REFERENCES usuarios(id) ON DELETE CASCADE,
				FOREIGN KEY (publicacao_id) REFERENCES publicacoes(id) ON DELETE CASCADE,
				criadoEm timestamp default current_timestamp
			);`,
		}
	case "codigos":
		return []string{
			`CREATE TABLE IF NOT EXISTS codigos (
				id serial PRIMARY KEY,
				usuario_id BIGINT NOT NULL,
				codigo TEXT NOT NULL,
				expiracao TIMESTAMP NOT NULL,
				CONSTRAINT usuario_id_unique UNIQUE (usuario_id),
				criadoEm timestamp default current_timestamp
			);`,
		}
	case "eventos":
		return []string{
			`CREATE TABLE IF NOT EXISTS eventos (
				id serial PRIMARY KEY,
				titulo varchar(100) NOT NULL,
				conteudo varchar(300) NOT NULL,
				rua varchar(150) NOT NULL,
				numero varchar(20) NOT NULL,
				bairro varchar(100) NOT NULL,
				cidade varchar(100) NOT NULL,
				estado varchar(100) NOT NULL,
				pais varchar(100) NOT NULL,
				cep varchar(20) NOT NULL,
				latitude double precision NOT NULL,
				longitude double precision NOT NULL,
				data_evento timestamp NOT NULL,
				data_evento_end timestamp NOT NULL,
				status varchar(50) NOT NULL,
				usuario_id int NOT NULL,
				categoria varchar(100) NOT NULL,
				criadoEm timestamp default current_timestamp,
				FOREIGN KEY (usuario_id) REFERENCES usuarios(id) ON DELETE CASCADE
			);`,
		}
	case "bilhetes":
		return []string{
			`CREATE TABLE IF NOT EXISTS bilhetes (
				id serial PRIMARY KEY,
				evento_id int NOT NULL,
				usuario_id int NOT NULL,
				forma_pagamento varchar(50) NOT NULL,
				status varchar(50) NOT NULL,
				status_pag varchar(50) NOT NULL,
				qrcode varchar(14) NOT NULL UNIQUE,
				id_lote int NOT NULL,
				valor double precision NOT NULL,
				valor_taxa double precision NOT NULL,
				valor_desconto double precision NOT NULL,
				titulo varchar(100) NOT NULL,
				criadoEm timestamp default current_timestamp,
				usadoEm timestamp,
				FOREIGN KEY (evento_id) REFERENCES eventos(id) ON DELETE CASCADE,
				FOREIGN KEY (usuario_id) REFERENCES usuarios(id) ON DELETE CASCADE
			);`,
		}
	case "lotes_evento":
		return []string{
			`CREATE TABLE IF NOT EXISTS lotes_evento (
				id serial PRIMARY KEY,
				evento_id int NOT NULL,
				usuario_id int NOT NULL,
				titulo varchar(100) NOT NULL,
				conteudo varchar(300) NOT NULL,
				data_inicio timestamp NOT NULL,
				data_fim timestamp NOT NULL,
				valor double precision NOT NULL,
				valor_taxa double precision NOT NULL,
    			quantidade_atual int NOT NULL DEFAULT 0,
    			quantidade_maxima int NOT NULL,
				criadoEm timestamp default current_timestamp,
				FOREIGN KEY (evento_id) REFERENCES eventos(id) ON DELETE CASCADE,
				FOREIGN KEY (usuario_id) REFERENCES usuarios(id) ON DELETE CASCADE
			);`,
		}
	case "mensagens_privadas":
		return []string{
			`CREATE TABLE IF NOT EXISTS mensagens_privadas (
				id SERIAL PRIMARY KEY,
				remetente_id INT NOT NULL,
				destinatario_id INT NOT NULL,
				conteudo TEXT NOT NULL,
				reacoes JSONB DEFAULT '[]',
				visualizado BOOLEAN DEFAULT FALSE,
				criado_em TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				FOREIGN KEY (remetente_id) REFERENCES usuarios(id) ON DELETE CASCADE,
				FOREIGN KEY (destinatario_id) REFERENCES usuarios(id) ON DELETE CASCADE
			);`,
		}
	case "mensagens_grupo":
		return []string{
			`CREATE TABLE IF NOT EXISTS mensagens_grupo (
				id SERIAL PRIMARY KEY,
				remetente_id INT NOT NULL,
				evento_id INT NOT NULL,
				conteudo TEXT NOT NULL,
				reacoes JSONB DEFAULT '[]',
				visualizado BOOLEAN DEFAULT FALSE,
				mencionado_id INT,
				criado_em TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				FOREIGN KEY (remetente_id) REFERENCES usuarios(id) ON DELETE CASCADE,
				FOREIGN KEY (evento_id) REFERENCES eventos(id) ON DELETE CASCADE,
				FOREIGN KEY (mencionado_id) REFERENCES usuarios(id) ON DELETE SET NULL
			);`,
		}

	}
	return nil
}
