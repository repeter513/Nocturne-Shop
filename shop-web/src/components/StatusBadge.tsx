interface Props {
  status: string
}

const LABELS: Record<string, string> = {
  ORDER_STATUS_PENDING: 'Ожидает оплаты',
  ORDER_STATUS_PAID: 'Оплачен',
  ORDER_STATUS_FAILED: 'Ошибка',
  ORDER_STATUS_CANCELLED: 'Отменён',
  PAYMENT_STATUS_PENDING: 'Ожидает',
  PAYMENT_STATUS_SUCCESS: 'Успешно',
  PAYMENT_STATUS_FAILED: 'Ошибка',
}

export function StatusBadge({ status }: Props) {
  const label = LABELS[status] ?? status
  const variant = status.includes('PAID') || status.includes('SUCCESS')
    ? 'success'
    : status.includes('FAILED') || status.includes('CANCELLED')
      ? 'danger'
      : 'muted'

  return <span className={`badge badge-${variant}`}>{label}</span>
}
