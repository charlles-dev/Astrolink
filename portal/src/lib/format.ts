export function formatCurrency(value: string | number): string {
  const amount = typeof value === 'number' ? value : Number(value)
  return new Intl.NumberFormat('pt-BR', {
    style: 'currency',
    currency: 'BRL'
  })
    .format(Number.isFinite(amount) ? amount : 0)
    .replace(/\u00a0/g, ' ')
}

export function formatCountdown(seconds: number | null | undefined): string {
  const amount = Number(seconds)
  const safeSeconds = Number.isFinite(amount) ? Math.max(0, Math.floor(amount)) : 0
  const days = Math.floor(safeSeconds / 86400)
  const hours = Math.floor((safeSeconds % 86400) / 3600)
  const minutes = Math.floor((safeSeconds % 3600) / 60)
  if (days > 0) {
    return `${days}d ${hours}h`
  }
  if (hours > 0) {
    return `${hours}h ${minutes}m`
  }
  return `${minutes}m`
}

export function formatDuration(minutes: number | null): string {
  if (minutes === null) return 'Por dados'
  if (minutes < 60) return `${minutes} min`
  if (minutes % 1440 === 0) {
    const days = minutes / 1440
    return days === 1 ? '24 horas' : `${days} dias`
  }
  if (minutes % 60 === 0) {
    const hours = minutes / 60
    return hours === 1 ? '1 hora' : `${hours} horas`
  }
  return `${minutes} min`
}

export function maskVoucherCode(value: string): string {
  const normalized = value.toUpperCase().replace(/[^A-Z0-9]/g, '').slice(0, 12)
  if (normalized.length <= 4) return normalized
  return `${normalized.slice(0, 4)}-${normalized.slice(4, 12)}`
}
