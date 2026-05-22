package admin_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/astrolink/node/internal/api/admin"
	adminauth "github.com/astrolink/node/internal/auth"
	"github.com/astrolink/node/internal/config"
	"github.com/astrolink/node/internal/domain/planos"
	"github.com/astrolink/node/internal/gateway"
	"github.com/astrolink/node/internal/store"
	"github.com/gofiber/fiber/v2"
)

func TestDesconectarUsuario_ChamaGatewayDeauthorize(t *testing.T) {
	app := fiber.New()
	repo := &fakeStore{}
	router := &fakeGateway{}
	admin.Register(app, admin.Dependencies{
		Config:  testConfig(),
		Store:   repo,
		Gateway: router,
	})
	tokens := loginAdmin(t, app)

	req := httptest.NewRequest("POST", "/admin/usuarios/AA:BB:CC:DD:EE:FF/desconectar", nil)
	req.Header.Set("Authorization", "Bearer "+tokens.AccessToken)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		t.Fatalf("status = %d body=%s", resp.StatusCode, string(body))
	}
	if !strings.Contains(string(body), `"sucesso":true`) {
		t.Fatalf("resposta inesperada: %s", string(body))
	}
	if len(router.deauths) != 1 || router.deauths[0] != "AA:BB:CC:DD:EE:FF" {
		t.Fatalf("deauths = %#v", router.deauths)
	}
}

func TestListarVouchers_RetornaVouchersDoStore(t *testing.T) {
	app := fiber.New()
	repo := &fakeStore{
		vouchers: []store.AdminVoucher{
			{
				ID:     1,
				Codigo: "VIPA-1234",
				Plano:  store.PlanoResumo{ID: 2, Nome: "Acesso 24 Horas"},
				Tipo:   "single_use",
				Ativo:  true,
			},
		},
	}
	admin.Register(app, admin.Dependencies{
		Config:  testConfig(),
		Store:   repo,
		Gateway: &fakeGateway{},
	})
	tokens := loginAdmin(t, app)

	req := httptest.NewRequest("GET", "/admin/vouchers", nil)
	req.Header.Set("Authorization", "Bearer "+tokens.AccessToken)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		t.Fatalf("status = %d body=%s", resp.StatusCode, string(body))
	}
	if !strings.Contains(string(body), `"codigo":"VIPA-1234"`) {
		t.Fatalf("resposta inesperada: %s", string(body))
	}
}

func TestGerarVouchers_RetornaCodigosCriados(t *testing.T) {
	app := fiber.New()
	repo := &fakeStore{
		generated: store.GenerateVouchersResult{
			LoteID:     7,
			Quantidade: 2,
			Vouchers: []store.AdminVoucher{
				{ID: 3, Codigo: "VIPA-1111", Plano: store.PlanoResumo{ID: 2, Nome: "Acesso 24 Horas"}, Tipo: "single_use", Ativo: true},
				{ID: 4, Codigo: "VIPA-2222", Plano: store.PlanoResumo{ID: 2, Nome: "Acesso 24 Horas"}, Tipo: "single_use", Ativo: true},
			},
		},
	}
	admin.Register(app, admin.Dependencies{
		Config:  testConfig(),
		Store:   repo,
		Gateway: &fakeGateway{},
	})
	tokens := loginAdmin(t, app)

	body := strings.NewReader(`{"plano_id":2,"quantidade":2,"prefixo":"VIPA"}`)
	req := httptest.NewRequest("POST", "/admin/vouchers/gerar", body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+tokens.AccessToken)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		payload, _ := io.ReadAll(resp.Body)
		t.Fatalf("status = %d body=%s", resp.StatusCode, string(payload))
	}
	if repo.generateInput.PlanoID != 2 || repo.generateInput.Quantidade != 2 || repo.generateInput.Prefixo != "VIPA" {
		t.Fatalf("input recebido = %+v", repo.generateInput)
	}
	var got store.GenerateVouchersResult
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	if got.LoteID != 7 || got.Quantidade != 2 || len(got.Vouchers) != 2 {
		t.Fatalf("resposta inesperada: %+v", got)
	}
}

func TestCriarPlano_RetornaPlanoCriado(t *testing.T) {
	app := fiber.New()
	repo := &fakeStore{
		createdPlano: planos.Plano{
			ID:               10,
			Nome:             "Noite",
			Descricao:        "Acesso noturno",
			Preco:            12.5,
			PrecoFormatado:   "12.50",
			DuracaoMinutos:   intPtr(480),
			DuracaoFormatada: "8 horas",
			DadosMB:          intPtr(2048),
			VelocidadeDown:   30,
			VelocidadeUp:     10,
			Recomendado:      true,
			Ativo:            true,
			VisivelPortal:    true,
			Ordem:            4,
		},
	}
	admin.Register(app, admin.Dependencies{Config: testConfig(), Store: repo, Gateway: &fakeGateway{}})
	tokens := loginAdmin(t, app)

	body := strings.NewReader(`{"nome":"Noite","descricao":"Acesso noturno","preco":12.5,"duracao_minutos":480,"dados_mb":2048,"velocidade_down":30,"velocidade_up":10,"recomendado":true,"ativo":true,"visivel_portal":true,"ordem":4}`)
	req := httptest.NewRequest("POST", "/admin/planos", body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+tokens.AccessToken)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		payload, _ := io.ReadAll(resp.Body)
		t.Fatalf("status = %d body=%s", resp.StatusCode, string(payload))
	}
	if repo.createPlanoInput.Nome != "Noite" || repo.createPlanoInput.Preco != 12.5 || repo.createPlanoInput.VelocidadeDown != 30 || repo.createPlanoInput.VelocidadeUp != 10 {
		t.Fatalf("input recebido = %+v", repo.createPlanoInput)
	}
	var got struct {
		Plano planos.Plano `json:"plano"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	if got.Plano.ID != 10 || got.Plano.PrecoFormatado != "12.50" || got.Plano.DadosMB == nil || *got.Plano.DadosMB != 2048 {
		t.Fatalf("plano inesperado: %+v", got.Plano)
	}
}

func TestAtualizarPlano_RetornaPlanoEditado(t *testing.T) {
	app := fiber.New()
	repo := &fakeStore{
		updatedPlano: planos.Plano{
			ID:               2,
			Nome:             "Dia inteiro",
			Descricao:        "24 horas",
			Preco:            18,
			PrecoFormatado:   "18.00",
			DuracaoMinutos:   intPtr(1440),
			DuracaoFormatada: "24 horas",
			VelocidadeDown:   50,
			VelocidadeUp:     20,
			Recomendado:      true,
			Ativo:            true,
			VisivelPortal:    false,
			Ordem:            1,
		},
	}
	admin.Register(app, admin.Dependencies{Config: testConfig(), Store: repo, Gateway: &fakeGateway{}})
	tokens := loginAdmin(t, app)

	body := strings.NewReader(`{"nome":"Dia inteiro","descricao":"24 horas","preco":18,"duracao_minutos":1440,"velocidade_down":50,"velocidade_up":20,"recomendado":true,"ativo":true,"visivel_portal":false,"ordem":1}`)
	req := httptest.NewRequest("PUT", "/admin/planos/2", body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+tokens.AccessToken)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		payload, _ := io.ReadAll(resp.Body)
		t.Fatalf("status = %d body=%s", resp.StatusCode, string(payload))
	}
	if repo.updatePlanoID != 2 || repo.updatePlanoInput.Nome != "Dia inteiro" || repo.updatePlanoInput.VisivelPortal {
		t.Fatalf("update recebido id=%d input=%+v", repo.updatePlanoID, repo.updatePlanoInput)
	}
	var got struct {
		Plano planos.Plano `json:"plano"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	if got.Plano.ID != 2 || got.Plano.VelocidadeDown != 50 || got.Plano.VisivelPortal {
		t.Fatalf("plano inesperado: %+v", got.Plano)
	}
}

func TestAlterarStatusPlano_RetornaPlanoAtualizado(t *testing.T) {
	app := fiber.New()
	repo := &fakeStore{
		statusPlano: planos.Plano{
			ID:               3,
			Nome:             "Semanal",
			PrecoFormatado:   "50.00",
			DuracaoFormatada: "7 dias",
			Ativo:            false,
			VisivelPortal:    true,
		},
	}
	admin.Register(app, admin.Dependencies{Config: testConfig(), Store: repo, Gateway: &fakeGateway{}})
	tokens := loginAdmin(t, app)

	req := httptest.NewRequest("PATCH", "/admin/planos/3/status", strings.NewReader(`{"ativo":false}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+tokens.AccessToken)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		payload, _ := io.ReadAll(resp.Body)
		t.Fatalf("status = %d body=%s", resp.StatusCode, string(payload))
	}
	if repo.statusPlanoID != 3 || repo.statusPlanoAtivo {
		t.Fatalf("status recebido id=%d ativo=%v", repo.statusPlanoID, repo.statusPlanoAtivo)
	}
	var got struct {
		Plano planos.Plano `json:"plano"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	if got.Plano.ID != 3 || got.Plano.Ativo {
		t.Fatalf("plano inesperado: %+v", got.Plano)
	}
}

func TestCriarPlano_ValidaPayload(t *testing.T) {
	cases := []struct {
		name string
		body string
	}{
		{name: "nome obrigatorio", body: `{"preco":1,"duracao_minutos":60,"velocidade_down":1,"velocidade_up":1,"ordem":1}`},
		{name: "preco nao negativo", body: `{"nome":"x","preco":-1,"duracao_minutos":60,"velocidade_down":1,"velocidade_up":1,"ordem":1}`},
		{name: "duracao positiva", body: `{"nome":"x","preco":1,"duracao_minutos":0,"velocidade_down":1,"velocidade_up":1,"ordem":1}`},
		{name: "velocidade nao negativa", body: `{"nome":"x","preco":1,"duracao_minutos":60,"velocidade_down":-1,"velocidade_up":1,"ordem":1}`},
		{name: "ordem razoavel", body: `{"nome":"x","preco":1,"duracao_minutos":60,"velocidade_down":1,"velocidade_up":1,"ordem":1000000}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			app := fiber.New()
			repo := &fakeStore{}
			admin.Register(app, admin.Dependencies{Config: testConfig(), Store: repo, Gateway: &fakeGateway{}})
			tokens := loginAdmin(t, app)

			req := httptest.NewRequest("POST", "/admin/planos", strings.NewReader(tc.body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+tokens.AccessToken)
			resp, err := app.Test(req, -1)
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != 400 {
				payload, _ := io.ReadAll(resp.Body)
				t.Fatalf("status = %d body=%s", resp.StatusCode, string(payload))
			}
			if repo.createPlanoCalled {
				t.Fatal("store nao deve ser chamado para payload invalido")
			}
		})
	}
}

func TestLogin_RetornaJWTAssinadoERefreshOpaco(t *testing.T) {
	app := fiber.New()
	repo := &fakeStore{}
	admin.Register(app, admin.Dependencies{
		Config:  testConfig(),
		Store:   repo,
		Gateway: &fakeGateway{},
	})

	tokens := loginAdmin(t, app)

	if strings.Count(tokens.AccessToken, ".") != 2 {
		t.Fatalf("access_token = %q, want JWT with 3 segments", tokens.AccessToken)
	}
	if tokens.RefreshToken == "" {
		t.Fatal("refresh_token vazio")
	}
	if tokens.RefreshToken == tokens.AccessToken || strings.Count(tokens.RefreshToken, ".") == 2 || strings.HasSuffix(tokens.RefreshToken, ".refresh") {
		t.Fatalf("refresh_token = %q, want opaque token", tokens.RefreshToken)
	}
	if tokens.TokenType != "Bearer" || tokens.ExpiresIn <= 0 {
		t.Fatalf("resposta auth inesperada: %+v", tokens)
	}
	if len(repo.createdSessions) != 1 {
		t.Fatalf("sessoes criadas = %d, want 1", len(repo.createdSessions))
	}
	session := repo.createdSessions[0]
	if session.Usuario != "admin" {
		t.Fatalf("usuario da sessao = %q", session.Usuario)
	}
	if session.RefreshTokenHash == "" || session.RefreshTokenHash == tokens.RefreshToken {
		t.Fatalf("refresh token deve ser armazenado como hash, got %q", session.RefreshTokenHash)
	}
	if !session.ExpiresAt.After(time.Now().UTC()) {
		t.Fatalf("expiracao da sessao = %s, want futuro", session.ExpiresAt)
	}
}

func TestLogin_ComSecretTOTPExigeCodigo(t *testing.T) {
	cases := []struct {
		name string
		body string
	}{
		{name: "ausente", body: `{"usuario":"admin","senha":"admin123"}`},
		{name: "em branco", body: `{"usuario":"admin","senha":"admin123","totp_codigo":"   "}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			app := fiber.New()
			repo := &fakeStore{}
			cfg := testConfig()
			cfg.AdminTOTPSecret = "JBSWY3DPEHPK3PXP"
			admin.Register(app, admin.Dependencies{
				Config:  cfg,
				Store:   repo,
				Gateway: &fakeGateway{},
			})

			resp := postAdminLogin(t, app, tc.body)
			defer resp.Body.Close()
			body, _ := io.ReadAll(resp.Body)

			if resp.StatusCode != 428 {
				t.Fatalf("status = %d body=%s, want 428", resp.StatusCode, string(body))
			}
			if !strings.Contains(string(body), "totp_obrigatorio") {
				t.Fatalf("erro = %s, want totp_obrigatorio", string(body))
			}
			if len(repo.createdSessions) != 0 {
				t.Fatalf("sessoes criadas = %d, want 0", len(repo.createdSessions))
			}
			if len(repo.loginFailures[adminLoginFailureKey(store.AdminLoginIdentity{Usuario: "admin", IP: "0.0.0.0"})]) != 1 {
				t.Fatalf("falhas registradas = %+v, want 1 for missing totp", repo.loginFailures)
			}
		})
	}
}

func TestLogin_ComSecretTOTPCodigoInvalidoContaFalha(t *testing.T) {
	app := fiber.New()
	repo := &fakeStore{}
	cfg := testConfig()
	cfg.AdminTOTPSecret = "JBSWY3DPEHPK3PXP"
	admin.Register(app, admin.Dependencies{
		Config:  cfg,
		Store:   repo,
		Gateway: &fakeGateway{},
	})

	resp := postAdminLogin(t, app, `{"usuario":"admin","senha":"admin123","totp_codigo":"000000"}`)
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 401 {
		t.Fatalf("status = %d body=%s, want 401", resp.StatusCode, string(body))
	}
	if !strings.Contains(string(body), "codigo 2FA invalido") {
		t.Fatalf("erro = %s, want codigo 2FA invalido", string(body))
	}
	if len(repo.createdSessions) != 0 {
		t.Fatalf("sessoes criadas = %d, want 0", len(repo.createdSessions))
	}
	failures := repo.loginFailures[adminLoginFailureKey(store.AdminLoginIdentity{Usuario: "admin", IP: "0.0.0.0"})]
	if len(failures) != 1 {
		t.Fatalf("falhas registradas = %d, want 1", len(failures))
	}
}

func TestLogin_ComSecretTOTPCodigoValidoEmiteTokensELimpaFalhas(t *testing.T) {
	app := fiber.New()
	identity := store.AdminLoginIdentity{Usuario: "admin", IP: "0.0.0.0"}
	repo := &fakeStore{
		loginFailures: map[string][]time.Time{
			adminLoginFailureKey(identity): []time.Time{time.Now().UTC()},
		},
	}
	cfg := testConfig()
	cfg.AdminTOTPSecret = "JBSWY3DPEHPK3PXP"
	admin.Register(app, admin.Dependencies{
		Config:  cfg,
		Store:   repo,
		Gateway: &fakeGateway{},
	})
	code, err := adminauth.GenerateTOTPCodeAt(cfg.AdminTOTPSecret, time.Now().UTC())
	if err != nil {
		t.Fatalf("GenerateTOTPCodeAt() error = %v", err)
	}

	resp := postAdminLogin(t, app, `{"usuario":"admin","senha":"admin123","totp_codigo":`+strconvQuote(code)+`}`)
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		t.Fatalf("status = %d body=%s, want 200", resp.StatusCode, string(body))
	}
	if len(repo.createdSessions) != 1 {
		t.Fatalf("sessoes criadas = %d, want 1", len(repo.createdSessions))
	}
	if len(repo.loginFailures[adminLoginFailureKey(identity)]) != 0 {
		t.Fatalf("falhas restantes = %+v, want none", repo.loginFailures)
	}
}

func TestLogin_BloqueiaAposCincoFalhasRecentes(t *testing.T) {
	app := fiber.New()
	repo := &fakeStore{}
	admin.Register(app, admin.Dependencies{
		Config:  testConfig(),
		Store:   repo,
		Gateway: &fakeGateway{},
	})

	for i := 0; i < 4; i++ {
		resp := postAdminLogin(t, app, `{"usuario":"admin","senha":"errada"}`)
		if resp.StatusCode != 401 {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			t.Fatalf("falha %d status = %d body=%s, want 401", i+1, resp.StatusCode, string(body))
		}
		resp.Body.Close()
	}

	resp := postAdminLogin(t, app, `{"usuario":"admin","senha":"errada"}`)
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 429 {
		t.Fatalf("quinta falha status = %d body=%s, want 429", resp.StatusCode, string(body))
	}
	if len(repo.createdSessions) != 0 {
		t.Fatalf("sessoes criadas = %d, want 0", len(repo.createdSessions))
	}
	if !strings.Contains(string(body), "login bloqueado") {
		t.Fatalf("erro de lockout pouco claro: %s", string(body))
	}
}

func TestAdminRotasProtegidas_ExigemBearerToken(t *testing.T) {
	app := fiber.New()
	admin.Register(app, admin.Dependencies{
		Config:  testConfig(),
		Store:   &fakeStore{},
		Gateway: &fakeGateway{},
	})

	req := httptest.NewRequest("GET", "/admin/vouchers", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 401 {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("status = %d body=%s, want 401", resp.StatusCode, string(body))
	}
}

func TestAuthMe_RetornaUsuarioDoJWT(t *testing.T) {
	app := fiber.New()
	admin.Register(app, admin.Dependencies{
		Config:  testConfig(),
		Store:   &fakeStore{},
		Gateway: &fakeGateway{},
	})
	tokens := loginAdmin(t, app)

	req := httptest.NewRequest("GET", "/admin/auth/me", nil)
	req.Header.Set("Authorization", "Bearer "+tokens.AccessToken)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("status = %d body=%s", resp.StatusCode, string(body))
	}
	var got struct {
		Usuario string `json:"usuario"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	if got.Usuario != "admin" {
		t.Fatalf("usuario = %q", got.Usuario)
	}
}

func TestRefresh_RotacionaRefreshToken(t *testing.T) {
	app := fiber.New()
	repo := &fakeStore{}
	admin.Register(app, admin.Dependencies{
		Config:  testConfig(),
		Store:   repo,
		Gateway: &fakeGateway{},
	})
	tokens := loginAdmin(t, app)

	req := httptest.NewRequest("POST", "/admin/auth/refresh", strings.NewReader(`{"refresh_token":`+strconvQuote(tokens.RefreshToken)+`}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("status = %d body=%s", resp.StatusCode, string(body))
	}
	var refreshed authResponse
	if err := json.NewDecoder(resp.Body).Decode(&refreshed); err != nil {
		t.Fatal(err)
	}
	if refreshed.AccessToken == "" || strings.Count(refreshed.AccessToken, ".") != 2 {
		t.Fatalf("access_token renovado invalido: %+v", refreshed)
	}
	if refreshed.RefreshToken == "" || refreshed.RefreshToken == tokens.RefreshToken {
		t.Fatalf("refresh_token renovado = %q, antigo = %q", refreshed.RefreshToken, tokens.RefreshToken)
	}
	if len(repo.rotatedRefreshTokenHashes) != 1 {
		t.Fatalf("rotacoes = %d, want 1", len(repo.rotatedRefreshTokenHashes))
	}
	if repo.rotatedRefreshTokenHashes[0] == tokens.RefreshToken {
		t.Fatalf("refresh antigo foi enviado ao store sem hash")
	}
}

func TestLogout_RevogaRefreshToken(t *testing.T) {
	app := fiber.New()
	repo := &fakeStore{}
	admin.Register(app, admin.Dependencies{
		Config:  testConfig(),
		Store:   repo,
		Gateway: &fakeGateway{},
	})
	tokens := loginAdmin(t, app)

	req := httptest.NewRequest("POST", "/admin/auth/logout", strings.NewReader(`{"refresh_token":`+strconvQuote(tokens.RefreshToken)+`}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+tokens.AccessToken)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 204 {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("status = %d body=%s", resp.StatusCode, string(body))
	}
	if len(repo.revokedRefreshTokenHashes) != 1 {
		t.Fatalf("revogacoes = %d, want 1", len(repo.revokedRefreshTokenHashes))
	}
	if repo.revokedRefreshTokenHashes[0] == tokens.RefreshToken {
		t.Fatalf("refresh token foi enviado ao store sem hash")
	}
}

type fakeStore struct {
	vouchers                  []store.AdminVoucher
	generated                 store.GenerateVouchersResult
	generateInput             store.GenerateVouchersInput
	createdPlano              planos.Plano
	updatedPlano              planos.Plano
	statusPlano               planos.Plano
	createPlanoInput          store.AdminPlanoInput
	updatePlanoInput          store.AdminPlanoInput
	updatePlanoID             int
	statusPlanoID             int
	statusPlanoAtivo          bool
	createPlanoCalled         bool
	createdSessions           []store.CreateAdminSessionInput
	rotatedRefreshTokenHashes []string
	revokedRefreshTokenHashes []string
	loginFailures             map[string][]time.Time
	sessions                  map[string]store.AdminSession
}

func (fakeStore) Settings(context.Context) (store.Settings, error)     { return store.Settings{}, nil }
func (fakeStore) PortalPlanos(context.Context) ([]planos.Plano, error) { return nil, nil }
func (fakeStore) AdminPlanos(context.Context) ([]planos.Plano, error)  { return nil, nil }
func (fakeStore) Usuarios(context.Context) ([]store.Usuario, error)    { return nil, nil }
func (fakeStore) SessaoStatus(context.Context, string) (store.Usuario, error) {
	return store.Usuario{}, nil
}
func (fakeStore) CreatePix(context.Context, store.CreatePixInput) (store.PixTransaction, error) {
	return store.PixTransaction{}, nil
}
func (fakeStore) PixStatus(context.Context, string) (store.PixTransaction, bool, error) {
	return store.PixTransaction{}, false, nil
}
func (fakeStore) UpdatePixStatus(context.Context, store.UpdatePixStatusInput) (store.PixTransaction, bool, error) {
	return store.PixTransaction{}, false, nil
}
func (fakeStore) RedeemVoucher(context.Context, store.RedeemVoucherInput) (store.RedeemVoucherResult, error) {
	return store.RedeemVoucherResult{}, nil
}
func (fakeStore) Health(context.Context) store.Health { return store.Health{} }

func (f fakeStore) AdminVouchers(context.Context) ([]store.AdminVoucher, error) {
	return f.vouchers, nil
}

func (f *fakeStore) GenerateVouchers(_ context.Context, input store.GenerateVouchersInput) (store.GenerateVouchersResult, error) {
	f.generateInput = input
	return f.generated, nil
}

func (f *fakeStore) CreateAdminPlano(_ context.Context, input store.AdminPlanoInput) (planos.Plano, error) {
	f.createPlanoCalled = true
	f.createPlanoInput = input
	return f.createdPlano, nil
}

func (f *fakeStore) UpdateAdminPlano(_ context.Context, id int, input store.AdminPlanoInput) (planos.Plano, error) {
	f.updatePlanoID = id
	f.updatePlanoInput = input
	return f.updatedPlano, nil
}

func (f *fakeStore) SetAdminPlanoStatus(_ context.Context, id int, ativo bool) (planos.Plano, error) {
	f.statusPlanoID = id
	f.statusPlanoAtivo = ativo
	return f.statusPlano, nil
}

func (f *fakeStore) CreateAdminSession(_ context.Context, input store.CreateAdminSessionInput) error {
	f.createdSessions = append(f.createdSessions, input)
	f.ensureSessions()
	f.sessions[input.RefreshTokenHash] = store.AdminSession{
		Usuario:          input.Usuario,
		RefreshTokenHash: input.RefreshTokenHash,
		IP:               input.IP,
		UserAgent:        input.UserAgent,
		ExpiresAt:        input.ExpiresAt,
		CreatedAt:        time.Now().UTC(),
	}
	return nil
}

func (f *fakeStore) RotateAdminSession(_ context.Context, input store.RotateAdminSessionInput) (store.AdminSession, bool, error) {
	f.rotatedRefreshTokenHashes = append(f.rotatedRefreshTokenHashes, input.CurrentRefreshTokenHash)
	f.ensureSessions()
	current, ok := f.sessions[input.CurrentRefreshTokenHash]
	if !ok || current.Revoked || !current.ExpiresAt.After(input.Now) {
		return store.AdminSession{}, false, nil
	}
	current.Revoked = true
	f.sessions[input.CurrentRefreshTokenHash] = current
	if err := f.CreateAdminSession(context.Background(), store.CreateAdminSessionInput{
		Usuario:          current.Usuario,
		RefreshTokenHash: input.NextRefreshTokenHash,
		IP:               input.IP,
		UserAgent:        input.UserAgent,
		ExpiresAt:        input.ExpiresAt,
	}); err != nil {
		return store.AdminSession{}, false, err
	}
	return current, true, nil
}

func (f *fakeStore) RevokeAdminSession(_ context.Context, refreshTokenHash string) error {
	f.revokedRefreshTokenHashes = append(f.revokedRefreshTokenHashes, refreshTokenHash)
	f.ensureSessions()
	current, ok := f.sessions[refreshTokenHash]
	if !ok {
		return nil
	}
	current.Revoked = true
	f.sessions[refreshTokenHash] = current
	return nil
}

func (f *fakeStore) AdminLoginLocked(_ context.Context, query store.AdminLoginLockoutQuery) (bool, error) {
	key := adminLoginFailureKey(query.Identity)
	failures := f.recentLoginFailures(key, query.Since)
	f.loginFailures[key] = failures
	return len(failures) >= query.Limit, nil
}

func (f *fakeStore) RecordAdminLoginFailure(_ context.Context, input store.AdminLoginFailureInput) (store.AdminLoginFailureStatus, error) {
	key := adminLoginFailureKey(input.Identity)
	failures := f.recentLoginFailures(key, input.At.Add(-input.Window))
	failures = append(failures, input.At)
	f.loginFailures[key] = failures
	return store.AdminLoginFailureStatus{Failures: len(failures), Locked: len(failures) >= input.Limit}, nil
}

func (f *fakeStore) ClearAdminLoginFailures(_ context.Context, identity store.AdminLoginIdentity) error {
	f.ensureLoginFailures()
	delete(f.loginFailures, adminLoginFailureKey(identity))
	return nil
}

func (f *fakeStore) ensureSessions() {
	if f.sessions == nil {
		f.sessions = map[string]store.AdminSession{}
	}
}

func (f *fakeStore) recentLoginFailures(key string, since time.Time) []time.Time {
	f.ensureLoginFailures()
	recent := make([]time.Time, 0, len(f.loginFailures[key]))
	for _, failure := range f.loginFailures[key] {
		if !failure.Before(since) {
			recent = append(recent, failure)
		}
	}
	return recent
}

func (f *fakeStore) ensureLoginFailures() {
	if f.loginFailures == nil {
		f.loginFailures = map[string][]time.Time{}
	}
}

func adminLoginFailureKey(identity store.AdminLoginIdentity) string {
	return strings.ToLower(strings.TrimSpace(identity.Usuario)) + "|" + strings.TrimSpace(identity.IP)
}

type fakeGateway struct {
	authorizations []gateway.Authorization
	deauths        []string
}

func (f *fakeGateway) Authorize(_ context.Context, input gateway.Authorization) error {
	f.authorizations = append(f.authorizations, input)
	return nil
}

func (f *fakeGateway) Deauthorize(_ context.Context, mac string) error {
	f.deauths = append(f.deauths, mac)
	return nil
}

func (*fakeGateway) Ping(context.Context) (time.Duration, error) {
	return 0, nil
}

type authResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

func testConfig() config.Config {
	return config.Config{
		AdminUser:     "admin",
		AdminPassword: "admin123",
		JWTSecret:     "test-jwt-secret-com-mais-de-32-bytes",
	}
}

func loginAdmin(t *testing.T, app *fiber.App) authResponse {
	t.Helper()
	resp := postAdminLogin(t, app, `{"usuario":"admin","senha":"admin123"}`)
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("login status = %d body=%s", resp.StatusCode, string(body))
	}
	var tokens authResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokens); err != nil {
		t.Fatal(err)
	}
	return tokens
}

func postAdminLogin(t *testing.T, app *fiber.App, body string) *http.Response {
	t.Helper()
	req := httptest.NewRequest("POST", "/admin/auth/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}
	return resp
}

func strconvQuote(value string) string {
	encoded, _ := json.Marshal(value)
	return string(encoded)
}

func intPtr(value int) *int {
	return &value
}
