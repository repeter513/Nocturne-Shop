// Badge component for order and payment status labels. / Компонент бейджа для статусов заказов и платежей.

interface Props {
  // Proto enum string (e.g. ORDER_STATUS_PENDING). / Proto enum строка (например ORDER_STATUS_PENDING).
  status: string
}

// Human-readable Russian labels for proto status values. / Читаемые русские подписи для proto-статусов.
const LABELS: Record<string, string> = {
  ORDER_STATUS_PENDING: 'Ожидает оплаты',
  ORDER_STATUS_PAID: 'Оплачен',
  ORDER_STATUS_FAILED: 'Ошибка',
  ORDER_STATUS_CANCELLED: 'Отменён',
  PAYMENT_STATUS_PENDING: 'Ожидает',
  PAYMENT_STATUS_SUCCESS: 'Успешно',
  PAYMENT_STATUS_FAILED: 'Ошибка',
}

// Renders a colored badge for a given status string. / Отображает цветной бейдж для переданного статуса.
export function StatusBadge({ status }: Props) {
  const label = LABELS[status] ?? status
  // Map proto status substring to CSS variant. / Сопоставляем подстроку proto-статуса с CSS-вариантом.
  const variant = status.includes('PAID') || status.includes('SUCCESS')
    ? 'success'
    : status.includes('FAILED') || status.includes('CANCELLED')
      ? 'danger'
      : 'muted'

  return <span className={`badge badge-${variant}`}>{label}</span>
}
