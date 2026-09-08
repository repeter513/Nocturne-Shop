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

type Handler struct {
	paymentv1.UnimplementedPaymentServiceServer
	svc *service.PaymentService
}

func NewHandler(svc *service.PaymentService) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) CreatePayment(ctx context.Context, req *paymentv1.CreatePaymentRequest) (*paymentv1.CreatePaymentResponse, error) {
	userID, err := userID(ctx)
	if err != nil {
		return nil, err
	}
	p, err := h.svc.Create(ctx, req.GetOrderId(), userID, req.GetAmount(), req.GetSimulateFailure())
	if err != nil {
		return nil, mapError(err)
	}
	return &paymentv1.CreatePaymentResponse{
		PaymentId: p.ID,
		Status:    toProtoStatus(p.Status),
	}, nil
}

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

func userID(ctx context.Context) (int, error) {
	id, err := pkgauth.UserIDFromContext(ctx)
	if err != nil {
		return 0, status.Error(codes.Unauthenticated, err.Error())
	}
	return int(id), nil
}

func toProtoPayment(p *domain.Payment) *paymentv1.Payment {
	if p == nil {
		return nil
	}
	return &paymentv1.Payment{
		PaymentId: p.ID,
		OrderId:   p.OrderID,
		UserId:    p.UserID,
		Amount:    p.Amount,
		Status:    toProtoStatus(p.Status),
		CreatedAt: timestamppb.New(p.CreatedAt),
	}
}

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
