package modelos

// Senha representa o formato da requisição de alteração de senha
type Senha struct {
	Nova  string `json:"nova"`
	Atual string `json:"atual"`
}

// Senha representa o formato da requisição de alteração de senha
type ResetarSenha struct {
	NovaSenha      string `json:"novaSenha"`
	ConfirmarSenha string `json:"confirmarSenha"`
}
