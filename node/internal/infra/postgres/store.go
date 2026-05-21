package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/astrolink/node/internal/domain/planos"
	"github.com/astrolink/node/internal/domain/vouchers"
	"github.com/astrolink/node/internal/store"
)

type Store struct {
	db    *sql.DB
	clock func() time.Time
}

func NewStore(db *sql.DB, clock func() time.Time) *Store {
	if clock == nil {
		clock = time.Now
	}
	return &Store{db: db, clock: clock}
}

func (s *Store) Settings(ctx context.Context) (store.Settings, error) {
	settings := store.DefaultSettings()
	rows, err := s.db.QueryContext(ctx, `SELECT chave, valor FROM system_settings`)
	if err != nil {
		return store.Settings{}, fmt.Errorf("buscar settings: %w", err)
	}
	defer rows.Close()

	values := map[string]string{}
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return store.Settings{}, fmt.Errorf("ler settings: %w", err)
		}
		values[key] = value
	}
	if err := rows.Err(); err != nil {
		return store.Settings{}, fmt.Errorf("iterar settings: %w", err)
	}

	applySettings(&settings, values)
	return settings, nil
}

func (s *Store) PortalPlanos(ctx context.Context) ([]planos.Plano, error) {
	return s.queryPlanos(ctx, `SELECT id, nome, descricao, preco::text, duracao_minutos, dados_mb, velocidade_down, velocidade_up, recomendado, ativo, visivel_portal, ordem FROM planos WHERE ativo = TRUE AND visivel_portal = TRUE ORDER BY ordem ASC, preco ASC`)
}

func (s *Store) AdminPlanos(ctx context.Context) ([]planos.Plano, error) {
	return s.queryPlanos(ctx, `SELECT id, nome, descricao, preco::text, duracao_minutos, dados_mb, velocidade_down, velocidade_up, recomendado, ativo, visivel_portal, ordem FROM planos ORDER BY ordem ASC, id ASC`)
}

func (s *Store) Usuarios(ctx context.Context) ([]store.Usuario, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT
			u.id, u.mac::text, COALESCE(u.ip_atual::text, ''), COALESCE(u.nome, ''),
			u.status, u.inicio_acesso, u.fim_acesso, u.dados_consumidos_mb,
			p.id, p.nome, r.id, r.nome
		FROM usuarios_mac u
		LEFT JOIN planos p ON p.id = u.plano_id
		LEFT JOIN roteadores r ON r.id = u.roteador_id
		ORDER BY u.updated_at DESC, u.id DESC
		LIMIT 200`)
	if err != nil {
		return nil, fmt.Errorf("buscar usuarios: %w", err)
	}
	defer rows.Close()

	var result []store.Usuario
	for rows.Next() {
		usuario, err := scanUsuario(rows, s.clock())
		if err != nil {
			return nil, err
		}
		result = append(result, usuario)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterar usuarios: %w", err)
	}
	return result, nil
}

func (s *Store) SessaoStatus(ctx context.Context, mac string) (store.Usuario, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT
			u.id, u.mac::text, COALESCE(u.ip_atual::text, ''), COALESCE(u.nome, ''),
			u.status, u.inicio_acesso, u.fim_acesso, u.dados_consumidos_mb,
			p.id, p.nome, r.id, r.nome
		FROM usuarios_mac u
		LEFT JOIN planos p ON p.id = u.plano_id
		LEFT JOIN roteadores r ON r.id = u.roteador_id
		WHERE u.mac = $1::macaddr`, normalizeMAC(mac))

	usuario, err := scanUsuario(row, s.clock())
	if errors.Is(err, sql.ErrNoRows) {
		return store.Usuario{MAC: normalizeMAC(mac), Status: "walled_garden"}, nil
	}
	if err != nil {
		return store.Usuario{}, err
	}
	if usuario.FimAcesso != nil && usuario.FimAcesso.Before(s.clock()) {
		usuario.Status = "expirado"
		usuario.TempoRestanteSegundos = 0
	}
	return usuario, nil
}

func (s *Store) CreatePix(ctx context.Context, input store.CreatePixInput) (store.PixTransaction, error) {
	plano, err := s.findPlano(ctx, input.PlanoID)
	if err != nil {
		return store.PixTransaction{}, err
	}
	now := s.clock().UTC()
	txid := fmt.Sprintf("ast_%d", now.UnixNano())
	tx := store.PixTransaction{
		TXID:             txid,
		Valor:            plano.PrecoFormatado,
		Descricao:        "Astrolink Wi-Fi - " + plano.Nome,
		PixCopiaCola:     "00020126580014br.gov.bcb.pix0136astrolink-demo-" + txid,
		QRCodeBase64:     "data:image/svg+xml;base64,PHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHdpZHRoPSIyNTYiIGhlaWdodD0iMjU2Ij48cmVjdCB3aWR0aD0iMjU2IiBoZWlnaHQ9IjI1NiIgZmlsbD0id2hpdGUiLz48dGV4dCB4PSIxMjgiIHk9IjEyOCIgdGV4dC1hbmNob3I9Im1pZGRsZSIgZmlsbD0iYmxhY2siPkFzdHJvbGluayBQSVg8L3RleHQ+PC9zdmc+",
		ExpiraEm:         now.Add(15 * time.Minute),
		ExpiraEmSegundos: 900,
		Status:           "pendente",
		MAC:              normalizeMAC(input.MAC),
		PlanoID:          plano.ID,
	}

	_, err = s.db.ExecContext(ctx, `
		INSERT INTO transacoes_pix (txid, mac, plano_id, valor, status, pix_copia_cola, qr_code_base64, created_at, updated_at)
		VALUES ($1, $2::macaddr, $3, $4, 'pendente', $5, $6, $7, $7)`,
		tx.TXID, tx.MAC, tx.PlanoID, tx.Valor, tx.PixCopiaCola, tx.QRCodeBase64, now)
	if err != nil {
		return store.PixTransaction{}, fmt.Errorf("criar pix: %w", err)
	}
	if input.Nome != "" {
		_, _ = s.db.ExecContext(ctx, `
			INSERT INTO usuarios_mac (mac, ip_atual, nome, status, created_at, updated_at)
			VALUES ($1::macaddr, $2::inet, $3, 'walled_garden', $4, $4)
			ON CONFLICT (mac) DO UPDATE SET ip_atual = EXCLUDED.ip_atual, nome = EXCLUDED.nome, updated_at = EXCLUDED.updated_at`,
			tx.MAC, input.IP, input.Nome, now)
	}
	return tx, nil
}

func (s *Store) PixStatus(ctx context.Context, txid string) (store.PixTransaction, bool, error) {
	var (
		tx        store.PixTransaction
		createdAt time.Time
	)
	err := s.db.QueryRowContext(ctx, `
		SELECT txid, valor::text, status, COALESCE(pix_copia_cola, ''), COALESCE(qr_code_base64, ''), created_at, mac::text, plano_id
		FROM transacoes_pix
		WHERE txid = $1`, txid).
		Scan(&tx.TXID, &tx.Valor, &tx.Status, &tx.PixCopiaCola, &tx.QRCodeBase64, &createdAt, &tx.MAC, &tx.PlanoID)
	if errors.Is(err, sql.ErrNoRows) {
		return store.PixTransaction{}, false, nil
	}
	if err != nil {
		return store.PixTransaction{}, false, fmt.Errorf("buscar pix: %w", err)
	}
	tx.Descricao = "Astrolink Wi-Fi"
	tx.ExpiraEm = createdAt.Add(15 * time.Minute).UTC()
	tx.ExpiraEmSegundos = int(time.Until(tx.ExpiraEm).Seconds())
	if tx.ExpiraEmSegundos < 0 {
		tx.ExpiraEmSegundos = 0
	}
	return tx, true, nil
}

func (s *Store) RedeemVoucher(ctx context.Context, input store.RedeemVoucherInput) (store.RedeemVoucherResult, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return store.RedeemVoucherResult{}, fmt.Errorf("iniciar transacao voucher: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	voucher, err := selectVoucher(ctx, tx, normalizeVoucherCode(input.Codigo))
	if errors.Is(err, sql.ErrNoRows) {
		return store.RedeemVoucherResult{}, store.ErrVoucherNotFound
	}
	if err != nil {
		return store.RedeemVoucherResult{}, err
	}
	if err := voucher.Validar(s.clock()); err != nil {
		return store.RedeemVoucherResult{}, err
	}
	plano, err := selectPlano(ctx, tx, voucher.PlanoID)
	if err != nil {
		return store.RedeemVoucherResult{}, err
	}

	previous, hadPrevious, err := selectUsuario(ctx, tx, normalizeMAC(input.MAC), s.clock())
	if err != nil {
		return store.RedeemVoucherResult{}, err
	}

	now := s.clock().UTC()
	start := now
	if hadPrevious && previous.FimAcesso != nil && previous.FimAcesso.After(now) {
		start = *previous.FimAcesso
	}
	duration := time.Hour
	if plano.DuracaoMinutos != nil {
		duration = time.Duration(*plano.DuracaoMinutos) * time.Minute
	}
	end := start.Add(duration)

	var userID int
	err = tx.QueryRowContext(ctx, `
		INSERT INTO usuarios_mac (mac, ip_atual, status, plano_id, inicio_acesso, fim_acesso, created_at, updated_at)
		VALUES ($1::macaddr, $2::inet, 'ativo', $3, $4, $5, $4, $4)
		ON CONFLICT (mac) DO UPDATE SET
			ip_atual = EXCLUDED.ip_atual,
			status = 'ativo',
			plano_id = EXCLUDED.plano_id,
			inicio_acesso = EXCLUDED.inicio_acesso,
			fim_acesso = EXCLUDED.fim_acesso,
			updated_at = EXCLUDED.updated_at
		RETURNING id`, normalizeMAC(input.MAC), input.IP, plano.ID, now, end).Scan(&userID)
	if err != nil {
		return store.RedeemVoucherResult{}, fmt.Errorf("atualizar usuario por voucher: %w", err)
	}
	_, err = tx.ExecContext(ctx, `UPDATE vouchers SET usos_atuais = usos_atuais + 1 WHERE id = $1`, voucher.ID)
	if err != nil {
		return store.RedeemVoucherResult{}, fmt.Errorf("atualizar usos do voucher: %w", err)
	}
	minutes := 60
	if plano.DuracaoMinutos != nil {
		minutes = *plano.DuracaoMinutos
	}
	_, err = tx.ExecContext(ctx, `
		INSERT INTO voucher_usos (voucher_id, mac, ip, tempo_adicionado_min, created_at)
		VALUES ($1, $2::macaddr, $3::inet, $4, $5)`, voucher.ID, normalizeMAC(input.MAC), input.IP, minutes, now)
	if err != nil {
		return store.RedeemVoucherResult{}, fmt.Errorf("registrar uso do voucher: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return store.RedeemVoucherResult{}, fmt.Errorf("commit voucher: %w", err)
	}

	usuario := store.Usuario{
		ID:                    userID,
		MAC:                   normalizeMAC(input.MAC),
		IPAtual:               input.IP,
		Status:                "ativo",
		Plano:                 &store.PlanoResumo{ID: plano.ID, Nome: plano.Nome},
		InicioAcesso:          &now,
		FimAcesso:             &end,
		TempoRestanteSegundos: int64(time.Until(end).Seconds()),
	}
	return store.RedeemVoucherResult{
		Usuario:   usuario,
		Plano:     plano,
		HadAccess: hadPrevious && previous.Status == "ativo",
	}, nil
}

func (s *Store) Health(ctx context.Context) store.Health {
	start := time.Now()
	err := s.db.PingContext(ctx)
	latency := time.Since(start).Milliseconds()
	if err != nil {
		return store.Health{DatabaseStatus: "error", DatabaseLatencyMS: latency}
	}
	return store.Health{DatabaseStatus: "ok", DatabaseLatencyMS: latency}
}

func (s *Store) queryPlanos(ctx context.Context, query string) ([]planos.Plano, error) {
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("buscar planos: %w", err)
	}
	defer rows.Close()
	var result []planos.Plano
	for rows.Next() {
		plano, err := scanPlano(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, plano)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterar planos: %w", err)
	}
	return result, nil
}

func (s *Store) findPlano(ctx context.Context, id int) (planos.Plano, error) {
	return selectPlano(ctx, s.db, id)
}

type queryer interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

func selectPlano(ctx context.Context, q queryer, id int) (planos.Plano, error) {
	row := q.QueryRowContext(ctx, `
		SELECT id, nome, descricao, preco::text, duracao_minutos, dados_mb, velocidade_down, velocidade_up, recomendado, ativo, visivel_portal, ordem
		FROM planos
		WHERE id = $1`, id)
	plano, err := scanPlano(row)
	if errors.Is(err, sql.ErrNoRows) {
		return planos.Plano{}, store.ErrPlanoNotFound
	}
	if err != nil {
		return planos.Plano{}, err
	}
	return plano, nil
}

func selectVoucher(ctx context.Context, q queryer, code string) (vouchers.Voucher, error) {
	var (
		voucher    vouchers.Voucher
		tipo       string
		usosMax    sql.NullInt64
		validadeEm sql.NullTime
		prefixo    sql.NullString
	)
	err := q.QueryRowContext(ctx, `
		SELECT id, codigo, plano_id, tipo, usos_maximos, usos_atuais, validade_em, ativo, prefixo
		FROM vouchers
		WHERE codigo = $1
		FOR UPDATE`, code).
		Scan(&voucher.ID, &voucher.Codigo, &voucher.PlanoID, &tipo, &usosMax, &voucher.UsosAtuais, &validadeEm, &voucher.Ativo, &prefixo)
	if err != nil {
		return vouchers.Voucher{}, err
	}
	voucher.Tipo = vouchers.Tipo(tipo)
	if usosMax.Valid {
		value := int(usosMax.Int64)
		voucher.UsosMaximos = &value
	}
	if validadeEm.Valid {
		voucher.ValidadeEm = &validadeEm.Time
	}
	if prefixo.Valid {
		voucher.Prefixo = prefixo.String
	}
	return voucher, nil
}

func selectUsuario(ctx context.Context, q queryer, mac string, now time.Time) (store.Usuario, bool, error) {
	row := q.QueryRowContext(ctx, `
		SELECT
			u.id, u.mac::text, COALESCE(u.ip_atual::text, ''), COALESCE(u.nome, ''),
			u.status, u.inicio_acesso, u.fim_acesso, u.dados_consumidos_mb,
			p.id, p.nome, r.id, r.nome
		FROM usuarios_mac u
		LEFT JOIN planos p ON p.id = u.plano_id
		LEFT JOIN roteadores r ON r.id = u.roteador_id
		WHERE u.mac = $1::macaddr`, mac)
	usuario, err := scanUsuario(row, now)
	if errors.Is(err, sql.ErrNoRows) {
		return store.Usuario{}, false, nil
	}
	if err != nil {
		return store.Usuario{}, false, err
	}
	return usuario, true, nil
}

type scanner interface {
	Scan(...any) error
}

func scanPlano(row scanner) (planos.Plano, error) {
	var (
		plano    planos.Plano
		price    string
		desc     sql.NullString
		duration sql.NullInt64
		dataMB   sql.NullInt64
	)
	if err := row.Scan(
		&plano.ID,
		&plano.Nome,
		&desc,
		&price,
		&duration,
		&dataMB,
		&plano.VelocidadeDown,
		&plano.VelocidadeUp,
		&plano.Recomendado,
		&plano.Ativo,
		&plano.VisivelPortal,
		&plano.Ordem,
	); err != nil {
		return planos.Plano{}, err
	}
	if desc.Valid {
		plano.Descricao = desc.String
	}
	parsedPrice, err := strconv.ParseFloat(price, 64)
	if err != nil {
		return planos.Plano{}, fmt.Errorf("preco invalido %q: %w", price, err)
	}
	plano.Preco = parsedPrice
	plano.PrecoFormatado = price
	if duration.Valid {
		value := int(duration.Int64)
		plano.DuracaoMinutos = &value
	}
	if dataMB.Valid {
		value := int(dataMB.Int64)
		plano.DadosMB = &value
	}
	plano.DuracaoFormatada = planos.FormatDuration(plano.DuracaoMinutos)
	return plano, nil
}

func scanUsuario(row scanner, now time.Time) (store.Usuario, error) {
	var (
		usuario  store.Usuario
		inicio   sql.NullTime
		fim      sql.NullTime
		planoID  sql.NullInt64
		plano    sql.NullString
		routerID sql.NullInt64
		router   sql.NullString
	)
	if err := row.Scan(
		&usuario.ID,
		&usuario.MAC,
		&usuario.IPAtual,
		&usuario.Nome,
		&usuario.Status,
		&inicio,
		&fim,
		&usuario.DadosConsumidosMB,
		&planoID,
		&plano,
		&routerID,
		&router,
	); err != nil {
		return store.Usuario{}, err
	}
	if inicio.Valid {
		usuario.InicioAcesso = &inicio.Time
	}
	if fim.Valid {
		usuario.FimAcesso = &fim.Time
		remaining := fim.Time.Sub(now)
		if remaining > 0 {
			usuario.TempoRestanteSegundos = int64(remaining.Seconds())
		}
	}
	if planoID.Valid {
		usuario.Plano = &store.PlanoResumo{ID: int(planoID.Int64), Nome: plano.String}
	}
	if routerID.Valid {
		usuario.Roteador = &store.RoteadorResumo{ID: int(routerID.Int64), Nome: router.String}
	}
	return usuario, nil
}

func applySettings(settings *store.Settings, values map[string]string) {
	for key, value := range values {
		switch key {
		case "hotspot_nome":
			settings.HotspotNome = value
		case "hotspot_logo_url":
			settings.HotspotLogoURL = value
		case "cor_primaria":
			settings.CorPrimaria = value
		case "cor_secundaria":
			settings.CorSecundaria = value
		case "cor_fundo":
			settings.CorFundo = value
		case "mensagem_boas_vindas":
			settings.MensagemBoasVindas = value
		case "url_pos_conexao":
			settings.URLPosConexao = value
		case "coleta_nome":
			settings.ColetaNome = strings.EqualFold(value, "true")
		case "mostrar_velocidade":
			settings.MostrarVelocidade = !strings.EqualFold(value, "false")
		}
	}
}

func normalizeMAC(mac string) string {
	mac = strings.TrimSpace(mac)
	if mac == "" {
		return "00:00:00:00:00:00"
	}
	return strings.ToUpper(mac)
}

func normalizeVoucherCode(code string) string {
	return strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(code), " ", "-"))
}
