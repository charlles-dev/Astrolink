import { fireEvent, render, screen } from '@testing-library/svelte'
import { afterEach, describe, expect, it, vi } from 'vitest'

import AdminDashboard from './AdminDashboard.svelte'

describe('AdminDashboard', () => {
  afterEach(() => {
    localStorage.removeItem('astrolink.admin.theme')
    delete document.documentElement.dataset.adminTheme
    document.documentElement.style.colorScheme = ''
  })

  it('renders health, plans and connected users', () => {
    render(AdminDashboard, {
      props: {
        activePage: 'overview',
        health: {
          status: 'healthy',
          versao: '0.1.0',
          uptime_segundos: 0,
          checks: {
            banco_dados: { status: 'memory', latencia_ms: 0 },
            redis: { status: 'mock' },
            rabbitmq: { status: 'mock' },
            mercadopago: { status: 'mock' },
            roteadores: { total: 1, online: 1, offline: 0 }
          }
        },
        planos: [
          {
            id: 1,
            nome: 'Acesso 24 Horas',
            preco: '15.00',
            duracao_minutos: 1440,
            duracao_formatada: '24 horas',
            dados_mb: null,
            velocidade_down: 10,
            velocidade_up: 5,
            recomendado: true,
            ativo: true,
            visivel_portal: true,
            ordem: 1
          }
        ],
        vouchers: [
          {
            id: 1,
            codigo: 'VIPA-1234',
            plano: { id: 1, nome: 'Acesso 24 Horas' },
            tipo: 'single_use',
            usos_atuais: 0,
            ativo: true
          }
        ],
        usuarios: [
          {
            id: 1,
            mac: 'AA:BB:CC:DD:EE:FF',
            ip_atual: '192.168.1.50',
            status: 'ativo',
            plano: { id: 1, nome: 'Acesso 24 Horas' },
            tempo_restante_segundos: 3600,
            dados_consumidos_mb: 120
          }
        ],
        loading: false,
        actionMessage: '',
        onRefresh: vi.fn(),
        onDisconnect: vi.fn(),
        onGenerateVouchers: vi.fn(),
        onLogout: vi.fn()
      }
    })

    expect(screen.getByRole('heading', { name: 'Operação local' })).toBeInTheDocument()
    expect(screen.getByText('Banco')).toBeInTheDocument()
    expect(screen.getByText('memory')).toBeInTheDocument()
    expect(screen.getByText('Usuários ativos')).toBeInTheDocument()
    expect(screen.getByText('AA:BB:CC:DD:EE:FF')).toBeInTheDocument()
    expect(screen.getByRole('link', { name: 'Ver usuários' })).toHaveAttribute(
      'href',
      '/painel/usuarios'
    )
    expect(
      screen.queryByRole('button', { name: 'Desconectar AA:BB:CC:DD:EE:FF' })
    ).not.toBeInTheDocument()
  })

  it('marks the active page in shell navigation', () => {
    render(AdminDashboard, {
      props: {
        activePage: 'vouchers',
        health: null,
        planos: [],
        vouchers: [],
        usuarios: [],
        loading: false,
        actionMessage: '',
        onRefresh: vi.fn(),
        onDisconnect: vi.fn(),
        onGenerateVouchers: vi.fn(),
        onLogout: vi.fn()
      }
    })

    const voucherLinks = screen.getAllByRole('link', { name: 'Vouchers' })
    expect(voucherLinks).toHaveLength(1)
    expect(voucherLinks[0]).toHaveAttribute('aria-current', 'page')
    expect(screen.getByRole('heading', { name: 'Emissão de vouchers' })).toBeInTheDocument()

    screen.getAllByRole('link', { name: 'Usuários' }).forEach((link) => {
      expect(link).not.toHaveAttribute('aria-current')
    })
  })

  it('renders the admin shell with the Astrolink DaisyUI theme and controls', async () => {
    render(AdminDashboard, {
      props: {
        activePage: 'overview',
        health: null,
        planos: [],
        vouchers: [],
        usuarios: [],
        loading: false,
        actionMessage: '',
        onRefresh: vi.fn(),
        onDisconnect: vi.fn(),
        onGenerateVouchers: vi.fn(),
        onLogout: vi.fn()
      }
    })

    expect(screen.getByTestId('admin-shell')).toHaveAttribute('data-theme', 'astrolink')
    await fireEvent.click(screen.getByRole('button', { name: 'Ativar modo escuro' }))
    expect(screen.getByTestId('admin-shell')).toHaveAttribute('data-theme', 'astrolink-dark')
    expect(localStorage.getItem('astrolink.admin.theme')).toBe('dark')
    expect(document.documentElement.dataset.adminTheme).toBe('dark')
    expect(screen.getByRole('button', { name: 'Atualizar' })).toHaveClass('btn')
    expect(screen.getByRole('button', { name: 'Sair' })).toHaveClass('btn-primary')
  })

  it('requests user disconnect from the row action', async () => {
    const onDisconnect = vi.fn()

    render(AdminDashboard, {
      props: {
        activePage: 'usuarios',
        health: null,
        planos: [],
        vouchers: [],
        usuarios: [
          {
            id: 1,
            mac: 'AA:BB:CC:DD:EE:FF',
            status: 'ativo',
            tempo_restante_segundos: 0,
            dados_consumidos_mb: 0
          }
        ],
        loading: false,
        actionMessage: '',
        onRefresh: vi.fn(),
        onDisconnect,
        onGenerateVouchers: vi.fn(),
        onLogout: vi.fn()
      }
    })

    await fireEvent.click(screen.getByRole('button', { name: 'Desconectar AA:BB:CC:DD:EE:FF' }))

    expect(onDisconnect).toHaveBeenCalledWith('AA:BB:CC:DD:EE:FF')
  })

  it('submits single-use voucher generation with validity days', async () => {
    const onGenerateVouchers = vi.fn()

    render(AdminDashboard, {
      props: {
        activePage: 'vouchers',
        health: null,
        planos: [
          {
            id: 2,
            nome: 'Acesso 24 Horas',
            preco: '15.00',
            duracao_minutos: 1440,
            duracao_formatada: '24 horas',
            dados_mb: null,
            velocidade_down: 10,
            velocidade_up: 5,
            recomendado: true,
            ativo: true,
            visivel_portal: true,
            ordem: 1
          }
        ],
        vouchers: [],
        usuarios: [],
        loading: false,
        actionMessage: '',
        onRefresh: vi.fn(),
        onDisconnect: vi.fn(),
        onGenerateVouchers,
        onLogout: vi.fn()
      }
    })

    await fireEvent.input(screen.getByLabelText('Prefixo'), { target: { value: 'BARCO' } })
    await fireEvent.input(screen.getByLabelText('Quantidade'), { target: { value: '3' } })
    await fireEvent.input(screen.getByLabelText('Validade (dias)'), { target: { value: '7' } })
    await fireEvent.click(screen.getByRole('button', { name: 'Gerar vouchers' }))

    expect(onGenerateVouchers).toHaveBeenCalledWith({
      plano_id: 2,
      quantidade: 3,
      tipo: 'single_use',
      validade_dias: 7,
      prefixo: 'BARCO'
    })
  })

  it('submits universal voucher generation with max uses', async () => {
    const onGenerateVouchers = vi.fn()

    render(AdminDashboard, {
      props: {
        activePage: 'vouchers',
        health: null,
        planos: [
          {
            id: 2,
            nome: 'Acesso 24 Horas',
            preco: '15.00',
            duracao_minutos: 1440,
            duracao_formatada: '24 horas',
            dados_mb: null,
            velocidade_down: 10,
            velocidade_up: 5,
            recomendado: true,
            ativo: true,
            visivel_portal: true,
            ordem: 1
          }
        ],
        vouchers: [],
        usuarios: [],
        loading: false,
        actionMessage: '',
        onRefresh: vi.fn(),
        onDisconnect: vi.fn(),
        onGenerateVouchers,
        onLogout: vi.fn()
      }
    })

    await fireEvent.click(screen.getByLabelText('Universal'))
    await fireEvent.input(screen.getByLabelText('Prefixo'), { target: { value: 'pub' } })
    await fireEvent.input(screen.getByLabelText('Quantidade'), { target: { value: '4' } })
    await fireEvent.input(screen.getByLabelText('Usos máximos'), { target: { value: '25' } })
    await fireEvent.click(screen.getByRole('button', { name: 'Gerar vouchers' }))

    expect(onGenerateVouchers).toHaveBeenCalledWith({
      plano_id: 2,
      quantidade: 4,
      tipo: 'universal',
      usos_maximos: 25,
      prefixo: 'PUB'
    })
  })

  it('applies voucher filters', async () => {
    const onApplyVoucherFilters = vi.fn()

    render(AdminDashboard, {
      props: {
        activePage: 'vouchers',
        health: null,
        planos: [
          {
            id: 2,
            nome: 'Acesso 24 Horas',
            preco: '15.00',
            duracao_minutos: 1440,
            duracao_formatada: '24 horas',
            dados_mb: null,
            velocidade_down: 10,
            velocidade_up: 5,
            recomendado: true,
            ativo: true,
            visivel_portal: true,
            ordem: 1
          }
        ],
        vouchers: [],
        usuarios: [],
        loading: false,
        actionMessage: '',
        onRefresh: vi.fn(),
        onDisconnect: vi.fn(),
        onGenerateVouchers: vi.fn(),
        onApplyVoucherFilters,
        onLogout: vi.fn()
      }
    })

    await fireEvent.change(screen.getByLabelText('Status'), { target: { value: 'inativo' } })
    await fireEvent.change(screen.getByLabelText('Plano do filtro'), { target: { value: '2' } })
    await fireEvent.input(screen.getByLabelText('Código'), { target: { value: 'vipa' } })
    await fireEvent.input(screen.getByLabelText('Lote'), { target: { value: '12' } })
    await fireEvent.click(screen.getByRole('button', { name: 'Aplicar filtros' }))

    expect(onApplyVoucherFilters).toHaveBeenCalledWith({
      status: 'inativo',
      plano_id: 2,
      codigo: 'VIPA',
      lote_id: 12
    })
  })

  it('confirms and requests voucher deactivation from the row action', async () => {
    const onDeactivateVoucher = vi.fn()

    render(AdminDashboard, {
      props: {
        activePage: 'vouchers',
        health: null,
        planos: [],
        vouchers: [
          {
            id: 7,
            codigo: 'VIPA-7777',
            plano: { id: 1, nome: 'Acesso 24 Horas' },
            tipo: 'single_use',
            usos_atuais: 0,
            ativo: true
          }
        ],
        usuarios: [],
        loading: false,
        actionMessage: '',
        onRefresh: vi.fn(),
        onDisconnect: vi.fn(),
        onGenerateVouchers: vi.fn(),
        onDeactivateVoucher,
        onLogout: vi.fn()
      }
    })

    await fireEvent.click(screen.getByRole('button', { name: 'Desativar VIPA-7777' }))
    expect(onDeactivateVoucher).not.toHaveBeenCalled()

    await fireEvent.click(screen.getByRole('button', { name: 'Confirmar desativação de VIPA-7777' }))

    expect(onDeactivateVoucher).toHaveBeenCalledWith(7)
  })

  it('requests voucher CSV export with filters', async () => {
    const onExportVouchers = vi.fn()

    render(AdminDashboard, {
      props: {
        activePage: 'vouchers',
        health: null,
        planos: [],
        vouchers: [],
        usuarios: [],
        loading: false,
        actionMessage: '',
        onRefresh: vi.fn(),
        onDisconnect: vi.fn(),
        onGenerateVouchers: vi.fn(),
        onExportVouchers,
        onLogout: vi.fn()
      }
    })

    await fireEvent.change(screen.getByLabelText('Status'), { target: { value: 'ativo' } })
    await fireEvent.input(screen.getByLabelText('Código'), { target: { value: 'vip' } })
    await fireEvent.click(screen.getByRole('button', { name: 'Exportar CSV' }))

    expect(onExportVouchers).toHaveBeenCalledWith({
      status: 'ativo',
      codigo: 'VIP'
    })
  })

  it('renders payments and requests filtered export', async () => {
    const onApplyPaymentFilters = vi.fn()
    const onExportPayments = vi.fn()

    render(AdminDashboard, {
      props: {
        activePage: 'pagamentos',
        health: null,
        planos: [],
        vouchers: [],
        usuarios: [],
        pagamentos: [
          {
            txid: 'pix-123',
            status: 'aprovado',
            valor: '15.00',
            descricao: 'Acesso 24 Horas',
            mac: 'AA:BB:CC:DD:EE:FF',
            plano_id: 1,
            plano: { id: 1, nome: 'Acesso 24 Horas' },
            created_at: '2026-05-21T10:00:00Z',
            expira_em: '2026-05-21T10:30:00Z'
          }
        ],
        pagamentosTotais: {
          pendente: 1,
          aprovado: 2,
          cancelado: 0,
          expirado: 0,
          valor_total: '30.00'
        },
        loading: false,
        actionMessage: '',
        onRefresh: vi.fn(),
        onDisconnect: vi.fn(),
        onGenerateVouchers: vi.fn(),
        onApplyPaymentFilters,
        onExportPayments,
        onLogout: vi.fn()
      }
    })

    expect(screen.getByRole('heading', { name: 'Pagamentos' })).toBeInTheDocument()
    expect(screen.getByText('pix-123')).toBeInTheDocument()
    expect(screen.getByText('R$ 30,00')).toBeInTheDocument()

    await fireEvent.change(screen.getByLabelText('Status do pagamento'), {
      target: { value: 'aprovado' }
    })
    await fireEvent.input(screen.getByLabelText('Início'), { target: { value: '2026-05-01' } })
    await fireEvent.input(screen.getByLabelText('Fim'), { target: { value: '2026-05-21' } })
    await fireEvent.click(screen.getByRole('button', { name: 'Exportar pagamentos CSV' }))

    expect(onExportPayments).toHaveBeenCalledWith({
      status: 'aprovado',
      inicio: '2026-05-01',
      fim: '2026-05-21'
    })
  })

  it('renders logs and backup operations', async () => {
    const onApplyLogFilters = vi.fn()
    const onExportLogs = vi.fn()
    const onCreateBackup = vi.fn()
    const onRestoreBackup = vi.fn()

    render(AdminDashboard, {
      props: {
        activePage: 'logs',
        health: null,
        planos: [],
        vouchers: [],
        usuarios: [],
        logs: [
          {
            timestamp: '2026-05-21T10:00:00Z',
            nivel: 'erro',
            tipo: 'backup',
            mensagem: 'Backup indisponível no modo memory'
          }
        ],
        backupMessage: 'Backup indisponível neste ambiente',
        loading: false,
        actionMessage: '',
        onRefresh: vi.fn(),
        onDisconnect: vi.fn(),
        onGenerateVouchers: vi.fn(),
        onApplyLogFilters,
        onExportLogs,
        onCreateBackup,
        onRestoreBackup,
        onLogout: vi.fn()
      }
    })

    expect(screen.getByRole('heading', { name: 'Logs' })).toBeInTheDocument()
    expect(screen.getByText('Backup indisponível no modo memory')).toBeInTheDocument()
    expect(screen.getByText('Backup indisponível neste ambiente')).toBeInTheDocument()

    await fireEvent.change(screen.getByLabelText('Nível'), { target: { value: 'erro' } })
    await fireEvent.input(screen.getByLabelText('Tipo'), { target: { value: 'backup' } })
    await fireEvent.input(screen.getByLabelText('Buscar texto'), { target: { value: 'memory' } })
    await fireEvent.click(screen.getByRole('button', { name: 'Aplicar filtros de logs' }))
    await fireEvent.click(screen.getByRole('button', { name: 'Exportar logs CSV' }))
    await fireEvent.click(screen.getByRole('button', { name: 'Gerar backup' }))
    await fireEvent.input(screen.getByLabelText('Arquivo do backup'), {
      target: { value: 'backup.sql' }
    })
    await fireEvent.input(screen.getByLabelText('Confirmação RESTAURAR'), {
      target: { value: 'RESTAURAR' }
    })
    await fireEvent.click(screen.getByRole('button', { name: 'Validar restore protegido' }))

    expect(onApplyLogFilters).toHaveBeenCalledWith({
      nivel: 'erro',
      tipo: 'backup',
      texto: 'memory'
    })
    expect(onExportLogs).toHaveBeenCalledWith({
      nivel: 'erro',
      tipo: 'backup',
      texto: 'memory'
    })
    expect(onCreateBackup).toHaveBeenCalled()
    expect(onRestoreBackup).toHaveBeenCalledWith({
      arquivo: 'backup.sql',
      confirmacao: 'RESTAURAR'
    })
  })

  it('renders live events from the dashboard stream state', () => {
    render(AdminDashboard, {
      props: {
        activePage: 'logs',
        health: null,
        planos: [],
        vouchers: [],
        usuarios: [],
        liveConnected: true,
        liveLastEventAt: '2026-05-21T10:35:00Z',
        liveSnapshot: {
          usuarios: { ativos: 1, total: 3 },
          vouchers: { ativos: 2, total: 4 },
          pix: { pendente: 1, aprovado: 5 },
          logs: 8
        },
        liveEvents: [
          {
            id: 'evt-1',
            tipo: 'snapshot',
            mensagem: 'Estado operacional atualizado',
            timestamp: '2026-05-21T10:35:00Z'
          }
        ],
        loading: false,
        actionMessage: '',
        onRefresh: vi.fn(),
        onDisconnect: vi.fn(),
        onGenerateVouchers: vi.fn(),
        onLogout: vi.fn()
      }
    })

    expect(screen.getByRole('heading', { name: 'Eventos ao vivo' })).toBeInTheDocument()
    expect(screen.getByText('Conectado')).toBeInTheDocument()
    expect(screen.getByText('1/3')).toBeInTheDocument()
    expect(screen.getByText('Estado operacional atualizado')).toBeInTheDocument()
  })
})
