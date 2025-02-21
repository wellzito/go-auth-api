package main

import (
	"api/src/config"
	"api/src/router"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/handlers" // Importando o pacote para lidar com CORS
	"github.com/joho/godotenv"    //remover se tiver em produção

	"api/src/controllers" // Importando o pacote de controllers
)

func main() {
	// Carrega as variáveis do .env remover se tiver em produção
	erro := godotenv.Load()
	if erro != nil {
		log.Fatal("Erro ao carregar o arquivo .env")
	}

	// Carregar as configurações
	config.Carregar()

	r := router.Gerar()

	// Inicializar o Redis
	_, err := config.InicializarRedis()
	if err != nil {
		log.Fatalf("Erro ao conectar ao Redis: %v", err)
	}

	// Rota para servir o arquivo index.html (quando a URL raiz for acessada)
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if _, err := os.Stat("static/index.html"); os.IsNotExist(err) {
			log.Println("Arquivo index.html não encontrado!")
			http.Error(w, "Arquivo não encontrado", http.StatusNotFound)
			return
		}
		http.ServeFile(w, r, "static/index.html") // Certifique-se de que o index.html esteja na raiz do seu projeto ou ajuste o caminho conforme necessário
	})

	// Rota para a página de home
	r.HandleFunc("/home", controllers.HomeHandler)

	// Rota para a página de login
	r.HandleFunc("/login", controllers.LoginHandler)

	// Rota para a página de registro
	r.HandleFunc("/register", controllers.RegisterHandler)

	// Rota para a página de registro
	r.HandleFunc("/logado", controllers.LogadoHandler)

	// Servir arquivos estáticos (HTML, CSS, JS)
	fs := http.FileServer(http.Dir("/app/static"))
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static", fs))

	// Habilitar CORS para permitir requisições de qualquer origem
	corsHandler := handlers.CORS(
		handlers.AllowedOrigins([]string{"*"}),                                       // Permite todas as origens
		handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}), // Permite métodos HTTP
		handlers.AllowedHeaders([]string{"Content-Type", "Authorization"}),           // Permite cabeçalhos específicos
	)(r)

	// Iniciar o servidor HTTP
	fmt.Printf("Escutando na porta %d\n", config.Porta)
	log.Println("Porta configurada:", config.Porta)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", config.Porta), corsHandler))
}
