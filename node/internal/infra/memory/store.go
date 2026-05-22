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
	"github.com/astrolink/node/internal/payments"
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
	adminFailures map[string][]time.Time
	adminLogs     []store.AdminLog
	nextPlanoID   int
	nextVoucherID int
	nextLoteID    int
}

func NewStore() *Store {
	durationHour := 60
	durationDay := 1440
	durationWeek := 10080
	maxUses := 25
	now := time.Now().UTC()
	expires := now.Add(30 * 24 * time.Hour)

	return &Store{
		settings: store.DefaultSettings(),
		planos: []planos.Plano{
			planos.New(1, "Acesso 1 Hora", "Internet rapida para resolver o essencial.", 5, &durationHour, false, 2),
			planos.New(2, "Acesso 24 Horas", "Um dia completo de internet.", 15, &durationDay, true, 1),
			planos.New(3, "Pacote Semanal", "Sete dias para familia ou equipe.", 50, &durationWeek, false, 3),
		},
		vouchers: map[string]vouchers.Voucher{
			"TEST-1234": {ID: 1, Codigo: "TEST-1234", PlanoID: 2, Tipo: vouchers.TipoSingleUse, Ativo: true, ValidadeEm: &expires, CreatedAt: now.Add(-2 * time.Minute)},
			"UNIV-0000": {ID: 2, Codigo: "UNIV-0000", PlanoID: 1, Tipo: vouchers.TipoUniversal, UsosMaximos: &maxUses, Ativo: true, CreatedAt: now.Add(-time.Minute)},
		},
		usuarios:      map[string]store.Usuario{},
		pix:           map[string]store.PixTransaction{},
		adminSessions: map[string]store.AdminSession{},
		adminFailures: map[string][]time.Time{},
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

func (s *Store) CreatePix(ctx context.Context, input store.CreatePixInput) (store.PixTransaction, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	plano, err := s.findPlano(input.PlanoID)
	if err != nil {
		return store.PixTransaction{}, err
	}
	txid := fmt.Sprintf("ast_%d_%d", time.Now().UnixNano(), len(s.pix)+1)
	now := time.Now().UTC()
	pix, err := payments.NewProvider(payments.ProviderConfig{Name: payments.ProviderDemo}).CreatePix(ctx, payments.CreatePixInput{
		TXID:      txid,
		Valor:     plano.PrecoFormatado,
		Descricao: "Astrolink Wi-Fi - " + plano.Nome,
		ExpiresAt: now.Add(15 * time.Minute),
	})
	if err != nil {
		return store.PixTransaction{}, err
	}
	tx := store.PixTransaction{
		TXID:             txid,
		Valor:            plano.PrecoFormatado,
		Descricao:        "Astrolink Wi-Fi - " + plano.Nome,
		PixCopiaCola:     pix.PixCopiaCola,
		QRCodeBase64:     pix.QRCodeBase64,
		CreatedAt:        now,
		ExpiraEm:         pix.ExpiresAt,
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
	txid = strings.TrimSpace(txid)
	tx, ok := s.pix[txid]
	if !ok {
		for _, candidate := range s.pix {
			if candidate.TXID == txid {
				return candidate, true, nil
			}
		}
	}
	return tx, ok, nil
}

func (s *Store) UpdatePixStatus(_ context.Context, input store.UpdatePixStatusInput) (store.PixTransaction, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	txid := strings.TrimSpace(input.TXID)
	tx, ok := s.pix[txid]
	if !ok {
		return store.PixTransaction{}, false, nil
	}
	status := strings.ToLower(strings.TrimSpace(input.Status))
	switch status {
	case "pendente", "aprovado", "cancelado", "expirado":
	default:
		return store.PixTransaction{}, false, fmt.Errorf("status PIX invalido")
	}
	tx.Status = status
	s.pix[txid] = tx
	return tx, true, nil
}

func (s *Store) AdminPagamentos(_ context.Context, filter store.AdminPagamentoFilter) ([]store.AdminPagamento, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]store.AdminPagamento, 0, len(s.pix))
	for _, tx := range s.pix {
		if !pixMatchesFilter(tx, filter) {
			continue
		}
		plano, err := s.findPlano(tx.PlanoID)
		if err != nil {
			continue
		}
		result = append(result, adminPagamentoFromPix(tx, plano))
	}
	sort.Slice(result, func(i, j int) bool {
		if !result[i].CreatedAt.Equal(result[j].CreatedAt) {
			return result[i].CreatedAt.After(result[j].CreatedAt)
		}
		return result[i].TXID > result[j].TXID
	})
	return result, nil
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

func (s *Store) AdminVouchers(ctx context.Context) ([]store.AdminVoucher, error) {
	return s.AdminVouchersFiltered(ctx, store.AdminVoucherFilter{Limit: 200})
}

func (s *Store) AdminVouchersFiltered(_ context.Context, filter store.AdminVoucherFilter) ([]store.AdminVoucher, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]store.AdminVoucher, 0, len(s.vouchers))
	for _, voucher := range s.vouchers {
		if !voucherMatchesFilter(voucher, filter) {
			continue
		}
		plano, err := s.findPlano(voucher.PlanoID)
		if err != nil {
			continue
		}
		result = append(result, adminVoucherFromDomain(voucher, plano))
	}
	sort.Slice(result, func(i, j int) bool {
		if !result[i].CreatedAt.Equal(result[j].CreatedAt) {
			return result[i].CreatedAt.After(result[j].CreatedAt)
		}
		return result[i].ID > result[j].ID
	})
	if filter.Limit > 0 && len(result) > filter.Limit {
		result = result[:filter.Limit]
	}
	return result, nil
}

func (s *Store) DeactivateVoucher(_ context.Context, id int) (store.AdminVoucher, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for code, voucher := range s.vouchers {
		if voucher.ID != id {
			continue
		}
		voucher.Ativo = false
		s.vouchers[code] = voucher
		plano, err := s.findPlano(voucher.PlanoID)
		if err != nil {
			return store.AdminVoucher{}, err
		}
		return adminVoucherFromDomain(voucher, plano), nil
	}
	return store.AdminVoucher{}, store.ErrVoucherNotFound
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
			LoteID:      &loteID,
			CreatedAt:   time.Now().UTC(),
		}
		s.nextVoucherID++
		s.vouchers[code] = voucher
		created = append(created, adminVoucherFromDomain(voucher, plano))
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

func (s *Store) AdminLoginLocked(_ context.Context, query store.AdminLoginLockoutQuery) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ensureAdminFailures()
	key := adminLoginFailureKey(query.Identity)
	failures := recentAdminLoginFailures(s.adminFailures[key], query.Since)
	s.adminFailures[key] = failures
	return len(failures) >= query.Limit, nil
}

func (s *Store) RecordAdminLoginFailure(_ context.Context, input store.AdminLoginFailureInput) (store.AdminLoginFailureStatus, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ensureAdminFailures()
	key := adminLoginFailureKey(input.Identity)
	failures := recentAdminLoginFailures(s.adminFailures[key], input.At.Add(-input.Window))
	failures = append(failures, input.At.UTC())
	s.adminFailures[key] = failures
	return store.AdminLoginFailureStatus{
		Failures: len(failures),
		Locked:   len(failures) >= input.Limit,
	}, nil
}

func (s *Store) ClearAdminLoginFailures(_ context.Context, identity store.AdminLoginIdentity) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ensureAdminFailures()
	delete(s.adminFailures, adminLoginFailureKey(identity))
	return nil
}

func (s *Store) AppendAdminLog(_ context.Context, input store.AdminLogInput) error {
	log := adminLogFromInput(input)
	s.mu.Lock()
	defer s.mu.Unlock()
	s.adminLogs = append(s.adminLogs, log)
	return nil
}

func (s *Store) AdminLogs(_ context.Context, filter store.AdminLogFilter) ([]store.AdminLog, error) {
	s.mu.RLock()
	logs := make([]store.AdminLog, 0, len(s.adminLogs))
	for _, log := range s.adminLogs {
		if adminLogMatchesFilter(log, filter) {
			logs = append(logs, cloneAdminLog(log))
		}
	}
	s.mu.RUnlock()

	sort.Slice(logs, func(i, j int) bool {
		if !logs[i].Timestamp.Equal(logs[j].Timestamp) {
			return logs[i].Timestamp.After(logs[j].Timestamp)
		}
		return logs[i].Mensagem > logs[j].Mensagem
	})
	return logs, nil
}

func (s *Store) Health(context.Context) store.Health {
	return store.Health{DatabaseStatus: "memory"}
}

func (s *Store) ensureAdminSessions() {
	if s.adminSessions == nil {
		s.adminSessions = map[string]store.AdminSession{}
	}
}

func (s *Store) ensureAdminFailures() {
	if s.adminFailures == nil {
		s.adminFailures = map[string][]time.Time{}
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

func adminLoginFailureKey(identity store.AdminLoginIdentity) string {
	return strings.ToLower(strings.TrimSpace(identity.Usuario)) + "|" + strings.TrimSpace(identity.IP)
}

func recentAdminLoginFailures(failures []time.Time, since time.Time) []time.Time {
	recent := make([]time.Time, 0, len(failures))
	for _, failure := range failures {
		if !failure.Before(since) {
			recent = append(recent, failure)
		}
	}
	return recent
}

func voucherMatchesFilter(voucher vouchers.Voucher, filter store.AdminVoucherFilter) bool {
	switch strings.ToLower(strings.TrimSpace(filter.Status)) {
	case "ativo":
		if !voucher.Ativo {
			return false
		}
	case "inativo":
		if voucher.Ativo {
			return false
		}
	}
	if filter.PlanoID != nil && voucher.PlanoID != *filter.PlanoID {
		return false
	}
	if filter.Codigo != "" && !strings.Contains(strings.ToLower(voucher.Codigo), strings.ToLower(strings.TrimSpace(filter.Codigo))) {
		return false
	}
	if filter.LoteID != nil {
		if voucher.LoteID == nil || *voucher.LoteID != *filter.LoteID {
			return false
		}
	}
	return true
}

func pixMatchesFilter(tx store.PixTransaction, filter store.AdminPagamentoFilter) bool {
	switch strings.ToLower(strings.TrimSpace(filter.Status)) {
	case "", "todos":
	case "pendente", "aprovado", "cancelado", "expirado":
		if tx.Status != strings.ToLower(strings.TrimSpace(filter.Status)) {
			return false
		}
	}
	if filter.Inicio != nil && tx.CreatedAt.Before(*filter.Inicio) {
		return false
	}
	if filter.Fim != nil {
		if filter.FimExclusive {
			if !tx.CreatedAt.Before(*filter.Fim) {
				return false
			}
		} else if tx.CreatedAt.After(*filter.Fim) {
			return false
		}
	}
	return true
}

func adminPagamentoFromPix(tx store.PixTransaction, plano planos.Plano) store.AdminPagamento {
	descricao := tx.Descricao
	if descricao == "" {
		descricao = "Astrolink Wi-Fi - " + plano.Nome
	}
	createdAt := tx.CreatedAt
	if createdAt.IsZero() {
		createdAt = tx.ExpiraEm.Add(-15 * time.Minute)
	}
	expiraEm := tx.ExpiraEm
	if expiraEm.IsZero() {
		expiraEm = createdAt.Add(15 * time.Minute)
	}
	return store.AdminPagamento{
		TXID:      tx.TXID,
		Status:    tx.Status,
		Valor:     tx.Valor,
		Descricao: descricao,
		MAC:       tx.MAC,
		PlanoID:   tx.PlanoID,
		Plano:     store.PlanoResumo{ID: plano.ID, Nome: plano.Nome},
		CreatedAt: createdAt,
		ExpiraEm:  expiraEm,
	}
}

func adminVoucherFromDomain(voucher vouchers.Voucher, plano planos.Plano) store.AdminVoucher {
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
		LoteID:      voucher.LoteID,
		CreatedAt:   voucher.CreatedAt,
	}
}

func adminLogFromInput(input store.AdminLogInput) store.AdminLog {
	timestamp := input.CreatedAt
	if timestamp.IsZero() {
		timestamp = time.Now().UTC()
	}
	return store.AdminLog{
		Timestamp:      timestamp.UTC(),
		Nivel:          normalizeAdminLogNivel(input.Nivel),
		Tipo:           normalizeAdminLogTipo(input.Tipo),
		Mensagem:       strings.TrimSpace(input.Mensagem),
		Detalhes:       cloneRawMessage(input.Detalhes),
		MACRelacionado: normalizeAdminLogMAC(input.MACRelacionado),
	}
}

func normalizeAdminLogNivel(nivel string) string {
	switch strings.ToLower(strings.TrimSpace(nivel)) {
	case "debug":
		return "debug"
	case "warn", "warning", "aviso":
		return "aviso"
	case "error", "erro":
		return "erro"
	default:
		return "info"
	}
}

func normalizeAdminLogTipo(tipo string) string {
	tipo = strings.ToLower(strings.TrimSpace(tipo))
	if tipo == "" {
		return "admin"
	}
	return tipo
}

func normalizeAdminLogMAC(mac string) string {
	mac = strings.TrimSpace(mac)
	if mac == "" {
		return ""
	}
	return strings.ToUpper(mac)
}

func adminLogMatchesFilter(log store.AdminLog, filter store.AdminLogFilter) bool {
	if filter.Nivel != "" && !strings.EqualFold(log.Nivel, filter.Nivel) {
		return false
	}
	if filter.Tipo != "" && !strings.EqualFold(log.Tipo, filter.Tipo) {
		return false
	}
	if filter.Texto != "" {
		haystack := strings.ToLower(strings.Join([]string{
			log.Nivel,
			log.Tipo,
			log.Mensagem,
			string(log.Detalhes),
			log.MACRelacionado,
		}, " "))
		if !strings.Contains(haystack, strings.ToLower(strings.TrimSpace(filter.Texto))) {
			return false
		}
	}
	return true
}

func cloneAdminLog(log store.AdminLog) store.AdminLog {
	log.Detalhes = cloneRawMessage(log.Detalhes)
	return log
}

func cloneRawMessage(raw []byte) []byte {
	if len(raw) == 0 {
		return nil
	}
	cloned := make([]byte, len(raw))
	copy(cloned, raw)
	return cloned
}
