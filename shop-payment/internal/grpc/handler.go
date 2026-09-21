// Package grpc implements the payment gRPC API handlers.
// Пакет grpc реализует gRPC-обработчики API платежей.
package grpc

import (
	"context"
	"errors"
	"github.com/repeter513/shop-payment/internal/domain"
	"github.com/repeter513/shop-payment/internal/service"
	paymentv1 "github.com/repeter513/shop-proto/gen/go/payment/v1"
	pkgauth "github.com/repeter513/shop-proto/pkg/auth"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Handler serves payment RPC methods.
// Handler обслуживает RPC-методы платежей.
type Handler struct {
	paymentv1.UnimplementedPaymentServiceServer
	// svc contains payment business logic invoked by each RPC method.
	// svc содержит бизнес-логику платежей, вызываемую каждым RPC-методом.
	svc *service.PaymentService
}

// NewHandler creates a gRPC handler backed by the payment service.
// NewHandler создаёт gRPC-обработчик на основе сервиса платежей.
func NewHandler(svc *service.PaymentService) *Handler {
	return &Handler{svc: svc}
}

// CreatePayment creates a payment for the given order.
// CreatePayment создаёт платёж по указанному заказу.
func (h *Handler) CreatePayment(ctx context.Context, req *paymentv1.CreatePaymentRequest) (*paymentv1.CreatePaymentResponse, error) {
	// Auth: user_id from JWT context identifies the payer.
	// Аутентификация: user_id из JWT-контекста определяет плательщика.
	userID, err := userID(ctx)
	if err != nil {
		return nil, err // codes.Unauthenticated
	}
	p, err := h.svc.Create(ctx, req.GetOrderId(), userID)
	if err != nil {
		return nil, mapError(err)
	}
	return &paymentv1.CreatePaymentResponse{
		PaymentId: p.ID,
		Status:    toProtoStatus(p.Status),
	}, nil
}

// GetPayment returns a payment by ID for the authenticated user.
// GetPayment возвращает платёж по ID для аутентифицированного пользователя.
func (h *Handler) GetPayment(ctx context.Context, req *paymentv1.GetPaymentRequest) (*paymentv1.GetPaymentResponse, error) {
	userID, err := userID(ctx)
	if err != nil {
		return nil, err
	}
	p, err := h.svc.Get(ctx, userID, req.GetPaymentId())

	if err != nil {
		return nil, mapError(err)
	}
	return &paymentv1.GetPaymentResponse{Payment: toProtoPayment(p)}, nil
}

// ListPayments returns a paginated list of payments.
// ListPayments возвращает постраничный список платежей.
func (h *Handler) ListPayments(ctx context.Context, req *paymentv1.ListPaymentsRequest) (*paymentv1.ListPaymentsResponse, error) {
	userID, err := userID(ctx)
	if err != nil {
		return nil, err
	}
	page := req.GetPage()
	if page == 0 {
		page = 1
	}
	pageSize := req.GetPageSize()
	if pageSize == 0 {
		pageSize = 10
	}

	payments, total, err := h.svc.List(ctx, int64(userID), req.GetOrderId(), page, pageSize)
	if err != nil {
		return nil, mapError(err)
	}

	out := make([]*paymentv1.Payment, 0, len(payments))
	for i := range payments {
		out = append(out, toProtoPayment(&payments[i]))
	}
	return &paymentv1.ListPaymentsResponse{
		Payments:   out,
		TotalCount: total,
	}, nil
}

// userID extracts the authenticated user ID from the request context.
// userID извлекает ID аутентифицированного пользователя из контекста запроса.
//
// The JWT interceptor validates the token (issuer + AudiencePayment) and stores
// user_id in context before this handler runs.
// JWT-интерцептор проверяет токен (issuer + AudiencePayment) и сохраняет
// user_id в контексте до вызова обработчика.
//
// Returns codes.Unauthenticated when the token is missing, expired, or invalid.
// Возвращает codes.Unauthenticated, когда токен отсутствует, просрочен или невалиден.
func userID(ctx context.Context) (int, error) {
	id, err := pkgauth.UserIDFromContext(ctx)
	if err != nil {
		return 0, status.Error(codes.Unauthenticated, err.Error())
	}
	return int(id), nil
}

// toProtoPayment converts a domain payment to its protobuf representation.
// toProtoPayment преобразует доменный платёж в protobuf-представление.
func toProtoPayment(p *domain.Payment) *paymentv1.Payment {
	if p == nil {
		return nil
	}
	return &paymentv1.Payment{
		PaymentId: p.ID,
		OrderId:   p.OrderID,
		UserId:    p.UserID,
		Amount:    int64(p.Amount),
		Status:    toProtoStatus(p.Status),
		CreatedAt: timestamppb.New(p.CreatedAt),
	}
}

// toProtoStatus maps a domain payment status to protobuf.
// toProtoStatus преобразует доменный статус платежа в protobuf.
func toProtoStatus(s domain.PaymentStatus) paymentv1.PaymentStatus {
	switch s {
	case domain.PaymentStatusPending:
		return paymentv1.PaymentStatus_PAYMENT_STATUS_PENDING
	case domain.PaymentStatusSuccess:
		return paymentv1.PaymentStatus_PAYMENT_STATUS_SUCCESS
	case domain.PaymentStatusFailed:
		return paymentv1.PaymentStatus_PAYMENT_STATUS_FAILED
	default:
		return paymentv1.PaymentStatus_PAYMENT_STATUS_UNSPECIFIED
	}
}

// mapError translates domain errors into gRPC status codes.
// mapError преобразует доменные ошибки в коды статуса gRPC.
//
// Status code mapping:
// Сопоставление кодов статуса:
//   - InvalidArgument  — bad input (missing IDs, invalid amount/page)
//   - NotFound         — payment does not exist or belongs to another user
//   - Internal         — unexpected database or downstream order service errors
//
//   - InvalidArgument  — неверный ввод (отсутствующие ID, невалидная сумма/страница)
//   - NotFound         — платёж не найден или принадлежит другому пользователю
//   - Internal         — неожиданные ошибки БД или downstream-сервиса заказов
func mapError(err error) error {
	switch {
	case errors.Is(err, domain.ErrPaymentIDRequired),
		errors.Is(err, domain.ErrOrderIDRequired),
		errors.Is(err, domain.ErrUserIDRequired),
		errors.Is(err, domain.ErrInvalidAmount),
		errors.Is(err, domain.ErrInvalidPage),
		errors.Is(err, domain.ErrInvalidPageSize):
		return status.Error(codes.InvalidArgument, err.Error())

	case errors.Is(err, domain.ErrPaymentNotFound):
		return status.Error(codes.NotFound, err.Error())

	default:
		return status.Error(codes.Internal, err.Error())
	}
}
