package banco

import (
	"api/src/config"
	"database/sql"

	_ "github.com/lib/pq" // Driver PostgreSQL
)

// Conectar abre a conexão com o banco de dados e a retorna
func Conectar() (*sql.DB, error) {
	// Usando o driver PostgreSQL em vez de MySQL
	db, erro := sql.Open("postgres", config.StringConexaoBanco)
	if erro != nil {
		return nil, erro
	}

	// Verifica a conexão com o banco de dados
	if erro = db.Ping(); erro != nil {
		db.Close()
		return nil, erro
	}

	return db, nil
}
