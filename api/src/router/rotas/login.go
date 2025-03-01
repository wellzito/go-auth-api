package rotas

import (
	"api/src/controllers"
	"net/http"
)

var rotaLogin = []Rota{
	{
		URI:                "/login",
		Metodo:             http.MethodPost,
		Funcao:             controllers.Login,
		RequerAutenticacao: false,
	},
	{
		URI:                "/anonimo",
		Metodo:             http.MethodPost,
		Funcao:             controllers.LoginAnonimo,
		RequerAutenticacao: false,
	},
}
