// Global shop configuration constants. / Глобальные константы конфигурации магазина.

// Display name shown in the header. / Отображаемое имя в шапке сайта.
export const SHOP_NAME = 'Nocturne'

// Payment reservation window in minutes (must match backend order service). / Окно резервирования оплаты в минутах (должно совпадать с order-сервисом).
export const PAYMENT_RESERVE_MINUTES = 15

// Payment reservation window in milliseconds (used by countdown timer). / Окно резервирования оплаты в миллисекундах (для таймера обратного отсчёта).
export const PAYMENT_RESERVE_MS = PAYMENT_RESERVE_MINUTES * 60 * 1000
