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
	tabelas := []string{"usuarios"}

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
	}
	return nil
}
