package controllers

import (
	"html/template"
	"log"
	"net/http"
)

var templates = template.Must(template.ParseGlob("static/*.html"))

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	err := renderTemplate(w, "index.html", nil)
	if err != nil {
		log.Printf("Erro ao renderizar template index.html: %v", err) // Log detalhado do erro
		http.Error(w, "Erro interno ao carregar a página de home.", http.StatusInternalServerError)
	}
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	err := renderTemplate(w, "login.html", nil)
	if err != nil {
		log.Printf("Erro ao renderizar template login.html: %v", err) // Log detalhado do erro
		http.Error(w, "Erro interno ao carregar a página de login.", http.StatusInternalServerError)
	}
}

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	err := renderTemplate(w, "register.html", nil)
	if err != nil {
		log.Printf("Erro ao renderizar template register.html: %v", err) // Log detalhado do erro
		http.Error(w, "Erro interno ao carregar a página de registro.", http.StatusInternalServerError)
	}
}

func LogadoHandler(w http.ResponseWriter, r *http.Request) {
	err := renderTemplate(w, "logado.html", nil)
	if err != nil {
		log.Printf("Erro ao renderizar template logado.html: %v", err) // Log detalhado do erro
		http.Error(w, "Erro interno ao carregar a página de logado.", http.StatusInternalServerError)
	}
}

func renderTemplate(w http.ResponseWriter, tmpl string, data interface{}) error {
	// Verifica se o template existe
	if err := templates.ExecuteTemplate(w, tmpl, data); err != nil {
		log.Printf("Erro ao executar o template '%s': %v", tmpl, err) // Log detalhado
		return err
	}
	return nil
}
