package config

import "fmt"

type SetupStatus struct {
	RequiresRestart bool                  `json:"requires_restart"`
	Writable        bool                  `json:"writable"`
	Groups          map[string]SetupGroup `json:"groups"`
}

type SetupGroup struct {
	Label  string       `json:"label"`
	Fields []SetupField `json:"fields"`
}

type SetupField struct {
	Key         string `json:"key"`
	Label       string `json:"label"`
	Description string `json:"description"`
	Secret      bool   `json:"secret"`
	Configured  bool   `json:"configured"`
	Value       string `json:"value,omitempty"`
}

type setupGroupDefinition struct {
	Key   string
	Label string
}

type setupFieldDefinition struct {
	Group       string
	Key         string
	Label       string
	Description string
	Secret      bool
}

var setupGroups = []setupGroupDefinition{
	{Key: "system", Label: "Sistema local"},
	{Key: "security", Label: "Acesso e seguranca"},
	{Key: "database", Label: "Banco local"},
	{Key: "payments", Label: "Mercado Pago"},
	{Key: "opennds", Label: "OpenNDS"},
}

var setupFields = []setupFieldDefinition{
	{Group: "system", Key: "GO_ENV", Label: "Ambiente", Description: "development, staging ou production."},
	{Group: "system", Key: "LOG_LEVEL", Label: "Nivel de log", Description: "debug, info, warn ou error."},
	{Group: "system", Key: "HTTP_ADDR", Label: "Endereco HTTP", Description: "Host e porta do node local."},
	{Group: "system", Key: "NODE_NAME", Label: "Nome do node", Description: "Identificador local mostrado no painel."},
	{Group: "system", Key: "TIMEZONE", Label: "Fuso horario", Description: "Fuso horario usado nos relatorios locais."},
	{Group: "system", Key: EnvAstrolinkAllowEnvWrite, Label: "Escrita pelo painel", Description: "Habilita alteracao do .env pela API admin."},

	{Group: "security", Key: "ADMIN_USUARIO", Label: "Usuario admin", Description: "Login administrativo local."},
	{Group: "security", Key: "ADMIN_SENHA", Label: "Senha admin", Description: "Senha administrativa local.", Secret: true},
	{Group: "security", Key: "ADMIN_TOTP_SECRET", Label: "2FA admin", Description: "Segredo TOTP do administrador.", Secret: true},
	{Group: "security", Key: "JWT_SECRET", Label: "JWT secret", Description: "Chave usada para assinar sessoes locais.", Secret: true},

	{Group: "database", Key: "DATABASE_URL", Label: "Database URL", Description: "String de conexao do Postgres local.", Secret: true},
	{Group: "database", Key: "DB_PASSWORD", Label: "Senha Postgres", Description: "Senha do banco local.", Secret: true},
	{Group: "database", Key: "POSTGRES_PORT", Label: "Porta Postgres", Description: "Porta exposta pelo banco local."},
	{Group: "database", Key: "REDIS_URL", Label: "Redis URL", Description: "String de conexao do Redis local.", Secret: true},
	{Group: "database", Key: "REDIS_PASSWORD", Label: "Senha Redis", Description: "Senha do Redis local.", Secret: true},
	{Group: "database", Key: "RABBITMQ_USER", Label: "Usuario RabbitMQ", Description: "Usuario local do RabbitMQ."},
	{Group: "database", Key: "RABBITMQ_PASS", Label: "Senha RabbitMQ", Description: "Senha local do RabbitMQ.", Secret: true},
	{Group: "database", Key: "AMQP_URL", Label: "AMQP URL", Description: "String de conexao do RabbitMQ local.", Secret: true},

	{Group: "payments", Key: EnvPaymentsProvider, Label: "Provedor", Description: "Use demo para testes ou mercadopago para Pix real."},
	{Group: "payments", Key: EnvMercadoPagoAccessToken, Label: "Access token", Description: "Token privado da conta Mercado Pago.", Secret: true},
	{Group: "payments", Key: EnvMercadoPagoAPIBaseURL, Label: "API base URL", Description: "Sobrescreve a URL da API apenas em testes."},
	{Group: "payments", Key: EnvMercadoPagoPayerEmail, Label: "E-mail pagador", Description: "E-mail padrao para simulacoes locais."},
	{Group: "payments", Key: "MERCADOPAGO_WEBHOOK_SECRET", Label: "Webhook secret", Description: "Segredo usado para validar webhooks.", Secret: true},

	{Group: "opennds", Key: "OPENNDS_ENABLED", Label: "OpenNDS ativo", Description: "Liga a integracao SSH com o roteador."},
	{Group: "opennds", Key: "OPENNDS_SSH_HOST", Label: "Host SSH", Description: "IP ou host do roteador OpenNDS."},
	{Group: "opennds", Key: "OPENNDS_SSH_PORT", Label: "Porta SSH", Description: "Porta SSH do roteador."},
	{Group: "opennds", Key: "OPENNDS_SSH_USER", Label: "Usuario SSH", Description: "Usuario SSH do roteador."},
	{Group: "opennds", Key: "OPENNDS_SSH_KEY_PATH", Label: "Chave SSH", Description: "Caminho local da chave privada SSH."},
	{Group: "opennds", Key: "OPENNDS_SSH_TIMEOUT", Label: "Timeout SSH", Description: "Tempo limite de comandos SSH."},
	{Group: "opennds", Key: "OPENNDS_AUTH_RETRIES", Label: "Tentativas auth", Description: "Tentativas de autenticacao no OpenNDS."},
}

func BuildSetupStatus(file *EnvFile) SetupStatus {
	status := SetupStatus{Groups: make(map[string]SetupGroup, len(setupGroups))}
	for _, group := range setupGroups {
		status.Groups[group.Key] = SetupGroup{Label: group.Label}
	}

	for _, definition := range setupFields {
		value := ""
		if file != nil {
			value = file.Get(definition.Key)
		}
		field := SetupField{
			Key:         definition.Key,
			Label:       definition.Label,
			Description: definition.Description,
			Secret:      definition.Secret,
			Configured:  value != "",
		}
		if !definition.Secret {
			field.Value = value
		}

		group := status.Groups[definition.Group]
		group.Fields = append(group.Fields, field)
		status.Groups[definition.Group] = group
	}

	return status
}

func ApplySetupPatch(file *EnvFile, values map[string]string) error {
	allowed := allowedSetupKeys()
	for _, key := range sortedEnvKeys(values) {
		if _, ok := allowed[key]; !ok {
			return fmt.Errorf("setup key %q is not allowed", key)
		}
		file.Set(key, values[key])
	}
	return nil
}

func allowedSetupKeys() map[string]setupFieldDefinition {
	allowed := make(map[string]setupFieldDefinition, len(setupFields))
	for _, definition := range setupFields {
		allowed[definition.Key] = definition
	}
	return allowed
}
