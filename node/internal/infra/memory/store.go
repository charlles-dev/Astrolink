package memory

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/astrolink/node/internal/domain/planos"
	"github.com/astrolink/node/internal/domain/vouchers"
	"github.com/astrolink/node/internal/store"
)

type Store struct {
	mu            sync.RWMutex
	settings      store.Settings
	planos        []planos.Plano
	vouchers      map[string]vouchers.Voucher
	usuarios      map[string]store.Usuario
	pix           map[string]store.PixTransaction
	adminSessions map[string]store.AdminSession
	nextPlanoID   int
	nextVoucherID int
	nextLoteID    int
}

func NewStore() *Store {
	durationHour := 60
	durationDay := 1440
	durationWeek := 10080
	maxUses := 25
	expires := time.Now().Add(30 * 24 * time.Hour)

	return &Store{
		settings: store.DefaultSettings(),
		planos: []planos.Plano{
			planos.New(1, "Acesso 1 Hora", "Internet rapida para resolver o essencial.", 5, &durationHour, false, 2),
			planos.New(2, "Acesso 24 Horas", "Um dia completo de internet.", 15, &durationDay, true, 1),
			planos.New(3, "Pacote Semanal", "Sete dias para familia ou equipe.", 50, &durationWeek, false, 3),
		},
		vouchers: map[string]vouchers.Voucher{
			"TEST-1234": {ID: 1, Codigo: "TEST-1234", PlanoID: 2, Tipo: vouchers.TipoSingleUse, Ativo: true, ValidadeEm: &expires},
			"UNIV-0000": {ID: 2, Codigo: "UNIV-0000", PlanoID: 1, Tipo: vouchers.TipoUniversal, UsosMaximos: &maxUses, Ativo: true},
		},
		usuarios:      map[string]store.Usuario{},
		pix:           map[string]store.PixTransaction{},
		adminSessions: map[string]store.AdminSession{},
		nextPlanoID:   4,
		nextVoucherID: 3,
		nextLoteID:    1,
	}
}

func (s *Store) Settings(_ context.Context) (store.Settings, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.settings, nil
}

func (s *Store) PortalPlanos(_ context.Context) ([]planos.Plano, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]planos.Plano, 0, len(s.planos))
	for _, plano := range s.planos {
		if plano.Ativo && plano.VisivelPortal {
			result = append(result, plano)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Ordem == result[j].Ordem {
			return result[i].ID < result[j].ID
		}
		return result[i].Ordem < result[j].Ordem
	})
	return result, nil
}

func (s *Store) AdminPlanos(_ context.Context) ([]planos.Plano, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]planos.Plano, len(s.planos))
	copy(result, s.planos)
	sortPlanos(result)
	return result, nil
}

func (s *Store) CreateAdminPlano(_ context.Context, input store.AdminPlanoInput) (planos.Plano, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	plano := planoFromInput(s.nextPlanoID, input)
	s.nextPlanoID++
	s.planos = append(s.planos, plano)
	sortPlanos(s.planos)
	return plano, nil
}

func (s *Store) UpdateAdminPlano(_ context.Context, id int, input store.AdminPlanoInput) (planos.Plano, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.planos {
		if s.planos[i].ID == id {
			plano := planoFromInput(id, input)
			s.planos[i] = plano
			sortPlanos(s.planos)
			return plano, nil
		}
	}
	return planos.Plano{}, store.ErrPlanoNotFound
}

func (s *Store) SetAdminPlanoStatus(_ context.Context, id int, ativo bool) (planos.Plano, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.planos {
		if s.planos[i].ID == id {
			s.planos[i].Ativo = ativo
			return s.planos[i], nil
		}
	}
	return planos.Plano{}, store.ErrPlanoNotFound
}

func (s *Store) Usuarios(_ context.Context) ([]store.Usuario, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]store.Usuario, 0, len(s.usuarios))
	for _, usuario := range s.usuarios {
		result = append(result, usuario)
	}
	return result, nil
}

func (s *Store) SessaoStatus(_ context.Context, mac string) (store.Usuario, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	usuario, ok := s.usuarios[normalizeMAC(mac)]
	if !ok {
		return store.Usuario{MAC: normalizeMAC(mac), Status: "walled_garden"}, nil
	}
	if usuario.FimAcesso != nil {
		remaining := time.Until(*usuario.FimAcesso)
		if remaining > 0 {
			usuario.TempoRestanteSegundos = int64(remaining.Seconds())
			return usuario, nil
		}
	}
	usuario.Status = "expirado"
	usuario.TempoRestanteSegundos = 0
	return usuario, nil
}

func (s *Store) CreatePix(_ context.Context, input store.CreatePixInput) (store.PixTransaction, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	plano, err := s.findPlano(input.PlanoID)
	if err != nil {
		return store.PixTransaction{}, err
	}
	txid := fmt.Sprintf("ast_%d", time.Now().UnixNano())
	tx := store.PixTransaction{
		TXID:             txid,
		Valor:            plano.PrecoFormatado,
		Descricao:        "Astrolink Wi-Fi - " + plano.Nome,
		PixCopiaCola:     "00020126580014br.gov.bcb.pix0136astrolink-demo-" + txid,
		QRCodeBase64:     "data:image/svg+xml;base64,PHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHdpZHRoPSIyNTYiIGhlaWdodD0iMjU2Ij48cmVjdCB3aWR0aD0iMjU2IiBoZWlnaHQ9IjI1NiIgZmlsbD0id2hpdGUiLz48dGV4dCB4PSIxMjgiIHk9IjEyOCIgdGV4dC1hbmNob3I9Im1pZGRsZSIgZmlsbD0iYmxhY2siPkFzdHJvbGluayBQSVg8L3RleHQ+PC9zdmc+",
		ExpiraEm:         time.Now().Add(15 * time.Minute).UTC(),
		ExpiraEmSegundos: 900,
		Status:           "pendente",
		MAC:              normalizeMAC(input.MAC),
		PlanoID:          plano.ID,
	}
	s.pix[txid] = tx
	if input.Nome != "" {
		s.usuarios[normalizeMAC(input.MAC)] = store.Usuario{MAC: normalizeMAC(input.MAC), IPAtual: input.IP, Nome: input.Nome, Status: "walled_garden"}
	}
	return tx, nil
}

func (s *Store) PixStatus(_ context.Context, txid string) (store.PixTransaction, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	tx, ok := s.pix[txid]
	return tx, ok, nil
}

func (s *Store) RedeemVoucher(_ context.Context, input store.RedeemVoucherInput) (store.RedeemVoucherResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	normalizedCode := strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(input.Codigo), " ", "-"))
	voucher, ok := s.vouchers[normalizedCode]
	if !ok {
		return store.RedeemVoucherResult{}, store.ErrVoucherNotFound
	}
	if err := voucher.Validar(time.Now()); err != nil {
		return store.RedeemVoucherResult{}, err
	}
	plano, err := s.findPlano(voucher.PlanoID)
	if err != nil {
		return store.RedeemVoucherResult{}, err
	}
	normalizedMAC := normalizeMAC(input.MAC)
	previous, hadAccess := s.usuarios[normalizedMAC]
	now := time.Now().UTC()
	start := now
	if previous.FimAcesso != nil && previous.FimAcesso.After(now) {
		start = *previous.FimAcesso
	}
	duration := time.Duration(60) * time.Minute
	if plano.DuracaoMinutos != nil {
		duration = time.Duration(*plano.DuracaoMinutos) * time.Minute
	}
	end := start.Add(duration)
	usuario := store.Usuario{
		ID:           previous.ID,
		MAC:          normalizedMAC,
		IPAtual:      input.IP,
		Nome:         previous.Nome,
		Status:       "ativo",
		Plano:        &store.PlanoResumo{ID: plano.ID, Nome: plano.Nome},
		InicioAcesso: &now,
		FimAcesso:    &end,
		Roteador:     &store.RoteadorResumo{ID: 1, Nome: "Roteador Principal"},
	}
	if usuario.ID == 0 {
		usuario.ID = len(s.usuarios) + 1
	}
	usuario.TempoRestanteSegundos = int64(time.Until(end).Seconds())
	s.usuarios[normalizedMAC] = usuario
	voucher.UsosAtuais++
	s.vouchers[normalizedCode] = voucher
	return store.RedeemVoucherResult{
		Usuario:   usuario,
		Plano:     plano,
		HadAccess: hadAccess && previous.Status == "ativo",
	}, nil
}

func (s *Store) AdminVouchers(_ context.Context) ([]store.AdminVoucher, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]store.AdminVoucher, 0, len(s.vouchers))
	for _, voucher := range s.vouchers {
		plano, err := s.findPlano(voucher.PlanoID)
		if err != nil {
			continue
		}
		result = append(result, adminVoucherFromDomain(voucher, plano, nil))
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].ID > result[j].ID
	})
	return result, nil
}

func (s *Store) GenerateVouchers(_ context.Context, input store.GenerateVouchersInput) (store.GenerateVouchersResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if input.Quantidade < 1 || input.Quantidade > 200 {
		return store.GenerateVouchersResult{}, store.ErrInvalidQuantity
	}
	plano, err := s.findPlano(input.PlanoID)
	if err != nil {
		return store.GenerateVouchersResult{}, err
	}
	tipo := vouchers.TipoSingleUse
	if input.Tipo == string(vouchers.TipoUniversal) {
		tipo = vouchers.TipoUniversal
	}
	var validadeEm *time.Time
	if input.ValidadeDias != nil && *input.ValidadeDias > 0 {
		expires := time.Now().UTC().Add(time.Duration(*input.ValidadeDias) * 24 * time.Hour)
		validadeEm = &expires
	}
	loteID := s.nextLoteID
	s.nextLoteID++
	created := make([]store.AdminVoucher, 0, input.Quantidade)
	for len(created) < input.Quantidade {
		code := vouchers.GerarCodigo(input.Prefixo)
		if _, exists := s.vouchers[code]; exists {
			continue
		}
		voucher := vouchers.Voucher{
			ID:          s.nextVoucherID,
			Codigo:      code,
			PlanoID:     plano.ID,
			Tipo:        tipo,
			UsosMaximos: input.UsosMaximos,
			ValidadeEm:  validadeEm,
			Ativo:       true,
			Prefixo:     strings.ToUpper(strings.TrimSpace(input.Prefixo)),
		}
		s.nextVoucherID++
		s.vouchers[code] = voucher
		created = append(created, adminVoucherFromDomain(voucher, plano, &loteID))
	}
	return store.GenerateVouchersResult{
		LoteID:     loteID,
		Quantidade: len(created),
		Vouchers:   created,
	}, nil
}

func (s *Store) CreateAdminSession(_ context.Context, input store.CreateAdminSessionInput) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ensureAdminSessions()
	s.adminSessions[input.RefreshTokenHash] = store.AdminSession{
		Usuario:          input.Usuario,
		RefreshTokenHash: input.RefreshTokenHash,
		IP:               input.IP,
		UserAgent:        input.UserAgent,
		ExpiresAt:        input.ExpiresAt,
		CreatedAt:        time.Now().UTC(),
	}
	return nil
}

func (s *Store) RotateAdminSession(_ context.Context, input store.RotateAdminSessionInput) (store.AdminSession, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ensureAdminSessions()
	current, ok := s.adminSessions[input.CurrentRefreshTokenHash]
	if !ok || current.Revoked || !current.ExpiresAt.After(input.Now) {
		return store.AdminSession{}, false, nil
	}
	current.Revoked = true
	s.adminSessions[input.CurrentRefreshTokenHash] = current
	s.adminSessions[input.NextRefreshTokenHash] = store.AdminSession{
		Usuario:          current.Usuario,
		RefreshTokenHash: input.NextRefreshTokenHash,
		IP:               input.IP,
		UserAgent:        input.UserAgent,
		ExpiresAt:        input.ExpiresAt,
		CreatedAt:        time.Now().UTC(),
	}
	return current, true, nil
}

func (s *Store) RevokeAdminSession(_ context.Context, refreshTokenHash string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ensureAdminSessions()
	current, ok := s.adminSessions[refreshTokenHash]
	if !ok {
		return nil
	}
	current.Revoked = true
	s.adminSessions[refreshTokenHash] = current
	return nil
}

func (s *Store) Health(context.Context) store.Health {
	return store.Health{DatabaseStatus: "memory"}
}

func (s *Store) ensureAdminSessions() {
	if s.adminSessions == nil {
		s.adminSessions = map[string]store.AdminSession{}
	}
}

func (s *Store) findPlano(id int) (planos.Plano, error) {
	for _, plano := range s.planos {
		if plano.ID == id {
			return plano, nil
		}
	}
	return planos.Plano{}, store.ErrPlanoNotFound
}

func planoFromInput(id int, input store.AdminPlanoInput) planos.Plano {
	return planos.FromConfig(planos.Config{
		ID:             id,
		Nome:           input.Nome,
		Descricao:      input.Descricao,
		Preco:          input.Preco,
		DuracaoMinutos: cloneInt(input.DuracaoMinutos),
		DadosMB:        cloneInt(input.DadosMB),
		VelocidadeDown: input.VelocidadeDown,
		VelocidadeUp:   input.VelocidadeUp,
		Recomendado:    input.Recomendado,
		Ativo:          input.Ativo,
		VisivelPortal:  input.VisivelPortal,
		Ordem:          input.Ordem,
	})
}

func cloneInt(value *int) *int {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}

func sortPlanos(items []planos.Plano) {
	sort.Slice(items, func(i, j int) bool {
		if items[i].Ordem == items[j].Ordem {
			return items[i].ID < items[j].ID
		}
		return items[i].Ordem < items[j].Ordem
	})
}

func normalizeMAC(mac string) string {
	mac = strings.TrimSpace(mac)
	if mac == "" {
		return "00:00:00:00:00:00"
	}
	return strings.ToUpper(mac)
}

func adminVoucherFromDomain(voucher vouchers.Voucher, plano planos.Plano, loteID *int) store.AdminVoucher {
	return store.AdminVoucher{
		ID:          voucher.ID,
		Codigo:      voucher.Codigo,
		Plano:       store.PlanoResumo{ID: plano.ID, Nome: plano.Nome},
		Tipo:        string(voucher.Tipo),
		UsosMaximos: voucher.UsosMaximos,
		UsosAtuais:  voucher.UsosAtuais,
		ValidadeEm:  voucher.ValidadeEm,
		Ativo:       voucher.Ativo,
		Prefixo:     voucher.Prefixo,
		LoteID:      loteID,
	}
}
