package service

import (
	"context"

	"github.com/repeter513/shop-payment/internal/domain"
	"github.com/repeter513/shop-payment/internal/repository"
)

type PaymentService struct {
	repo repository.PaymentRepository
}

func NewPaymentService(repo repository.PaymentRepository) *PaymentService {
	return &PaymentService{repo: repo}
}

func (s *PaymentService) Create(
	ctx context.Context,
	orderID int64,
	userID int,
	amount float64,
	simulateFailure bool,
) (*domain.Payment, error) {
	if orderID == 0 {
		return nil, domain.ErrOrderIDRequired
	}
	if userID == 0 {
		return nil, domain.ErrUserIDRequired
	}
	if amount <= 0 {
		return nil, domain.ErrInvalidAmount
	}

	p := &domain.Payment{
		OrderID: orderID,
		UserID:  int64(userID),
		Amount:  amount,
		Status:  domain.PaymentStatusSuccess,
	}
	if simulateFailure {
		p.Status = domain.PaymentStatusFailed
	}

	if err := s.repo.Create(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

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
	if p.UserID != int64(userID) {
		return nil, domain.ErrPaymentNotFound
	}
	return p, nil
}

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
