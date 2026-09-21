// Package service implements payment business logic.
// Пакет service реализует бизнес-логику платежей.
package service

import (
	"context"

	"github.com/repeter513/shop-payment/internal/adapter/order"
	"github.com/repeter513/shop-payment/internal/domain"
	"github.com/repeter513/shop-payment/internal/repository"
)

// PaymentService orchestrates payment creation and queries.
// PaymentService координирует создание и запросы платежей.
type PaymentService struct {
	// repo persists payment records in PostgreSQL.
	// repo сохраняет записи платежей в PostgreSQL.
	repo repository.PaymentRepository
	// order is the downstream gRPC client to fetch order totals.
	// order — downstream gRPC-клиент для получения суммы заказа.
	order *order.Client
}

// NewPaymentService creates a PaymentService with the given repository and order client.
// NewPaymentService создаёт PaymentService с указанным репозиторием и клиентом заказов.
func NewPaymentService(repo repository.PaymentRepository, orderClient *order.Client) *PaymentService {
	return &PaymentService{repo: repo, order: orderClient}
}

// Create records a successful payment for the given order.
// Create регистрирует успешный платёж по указанному заказу.
func (s *PaymentService) Create(ctx context.Context, orderID int64, userID int) (*domain.Payment, error) {
	// Step 1: validate required identifiers.
	// Шаг 1: проверка обязательных идентификаторов.
	if orderID == 0 {
		return nil, domain.ErrOrderIDRequired
	}
	if userID == 0 {
		return nil, domain.ErrUserIDRequired
	}

	// Step 2: downstream call — order.OrderTotal to fetch the authoritative charge amount.
	// Шаг 2: downstream-вызов — order.OrderTotal для получения авторитетной суммы списания.
	amount, err := s.order.OrderTotal(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if amount <= 0 {
		return nil, domain.ErrInvalidAmount
	}

	// Step 3: build payment entity with success status (MVP: instant success).
	// Шаг 3: формирование сущности платежа со статусом success (MVP: мгновенный успех).
	p := &domain.Payment{
		OrderID: orderID,
		UserID:  int64(userID),
		Amount:  float64(amount),
		Status:  domain.PaymentStatusSuccess,
	}

	// Step 4: persist with idempotency — duplicate order_id returns existing record.
	// Шаг 4: сохранение с идемпотентностью — дубликат order_id возвращает существующую запись.
	if err := s.repo.Create(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

// Get returns a payment by ID if it belongs to the given user.
// Get возвращает платёж по ID, если он принадлежит указанному пользователю.
func (s *PaymentService) Get(ctx context.Context, userID int, id int64) (*domain.Payment, error) {
	if userID == 0 {
		return nil, domain.ErrUserIDRequired
	}
	if id == 0 {
		return nil, domain.ErrPaymentIDRequired
	}
	p, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	// Ownership check: users may only read their own payments.
	// Проверка владения: пользователи могут читать только свои платежи.
	if p.UserID != int64(userID) {
		return nil, domain.ErrPaymentNotFound
	}
	return p, nil
}

// List returns a paginated list of payments filtered by user and order.
// List возвращает постраничный список платежей с фильтрацией по пользователю и заказу.
func (s *PaymentService) List(
	ctx context.Context,
	userID, orderID int64,
	page, pageSize int32,
) ([]domain.Payment, int32, error) {
	if page < 1 {
		return nil, 0, domain.ErrInvalidPage
	}
	if pageSize < 1 {
		return nil, 0, domain.ErrInvalidPageSize
	}
	return s.repo.List(ctx, userID, orderID, page, pageSize)
}
