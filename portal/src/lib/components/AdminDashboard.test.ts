import { fireEvent, render, screen } from '@testing-library/svelte'
import { describe, expect, it, vi } from 'vitest'

import AdminDashboard from './AdminDashboard.svelte'

describe('AdminDashboard', () => {
  it('renders health, plans and connected users', () => {
    render(AdminDashboard, {
      props: {
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

    expect(screen.getByRole('heading', { name: 'Painel local' })).toBeInTheDocument()
    expect(screen.getByText('Banco')).toBeInTheDocument()
    expect(screen.getByText('memory')).toBeInTheDocument()
    expect(screen.getAllByText('Acesso 24 Horas').length).toBeGreaterThan(0)
    expect(screen.getByText('AA:BB:CC:DD:EE:FF')).toBeInTheDocument()
    expect(screen.getByText('VIPA-1234')).toBeInTheDocument()
  })

  it('requests user disconnect from the row action', async () => {
    const onDisconnect = vi.fn()

    render(AdminDashboard, {
      props: {
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
    await fireEvent.input(screen.getByLabelText('Usos maximos'), { target: { value: '25' } })
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
    await fireEvent.input(screen.getByLabelText('Codigo'), { target: { value: 'vipa' } })
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

    await fireEvent.click(screen.getByRole('button', { name: 'Confirmar desativacao VIPA-7777' }))

    expect(onDeactivateVoucher).toHaveBeenCalledWith(7)
  })

  it('requests voucher CSV export with filters', async () => {
    const onExportVouchers = vi.fn()

    render(AdminDashboard, {
      props: {
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
    await fireEvent.input(screen.getByLabelText('Codigo'), { target: { value: 'vip' } })
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
    expect(screen.getByText('R$ 30.00')).toBeInTheDocument()

    await fireEvent.change(screen.getByLabelText('Status do pagamento'), {
      target: { value: 'aprovado' }
    })
    await fireEvent.input(screen.getByLabelText('Inicio'), { target: { value: '2026-05-01' } })
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

    render(AdminDashboard, {
      props: {
        health: null,
        planos: [],
        vouchers: [],
        usuarios: [],
        logs: [
          {
            timestamp: '2026-05-21T10:00:00Z',
            nivel: 'erro',
            tipo: 'backup',
            mensagem: 'Backup indisponivel no modo memory'
          }
        ],
        backupMessage: 'Backup indisponivel neste ambiente',
        loading: false,
        actionMessage: '',
        onRefresh: vi.fn(),
        onDisconnect: vi.fn(),
        onGenerateVouchers: vi.fn(),
        onApplyLogFilters,
        onExportLogs,
        onCreateBackup,
        onLogout: vi.fn()
      }
    })

    expect(screen.getByRole('heading', { name: 'Logs' })).toBeInTheDocument()
    expect(screen.getByText('Backup indisponivel no modo memory')).toBeInTheDocument()
    expect(screen.getByText('Backup indisponivel neste ambiente')).toBeInTheDocument()

    await fireEvent.change(screen.getByLabelText('Nivel'), { target: { value: 'erro' } })
    await fireEvent.input(screen.getByLabelText('Tipo'), { target: { value: 'backup' } })
    await fireEvent.input(screen.getByLabelText('Texto'), { target: { value: 'memory' } })
    await fireEvent.click(screen.getByRole('button', { name: 'Aplicar filtros de logs' }))
    await fireEvent.click(screen.getByRole('button', { name: 'Exportar logs CSV' }))
    await fireEvent.click(screen.getByRole('button', { name: 'Gerar backup' }))

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
  })
})
