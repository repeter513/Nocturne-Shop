// Package payment provides a gRPC client for the payment service.
// Пакет payment предоставляет gRPC-клиент для сервиса платежей.
package payment

import (
	"context"
	"fmt"

	paymentv1 "github.com/repeter513/shop-proto/gen/go/payment/v1"
	pkgauth "github.com/repeter513/shop-proto/pkg/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Client calls the remote payment gRPC service.
// Client вызывает удалённый gRPC-сервис платежей.
type Client struct {
	// conn is the persistent gRPC connection to the payment service.
	// conn — постоянное gRPC-соединение с сервисом платежей.
	conn *grpc.ClientConn
	// client is the generated stub for PaymentService RPCs.
	// client — сгенерированная заглушка для RPC PaymentService.
	client paymentv1.PaymentServiceClient
}

// New dials the payment service at addr and returns a Client.
// New подключается к сервису платежей по addr и возвращает Client.
func New(_ context.Context, addr string) (*Client, error) {
	conn, err := grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		// Forward JWT so payment service can identify the payer.
		// Проброс JWT, чтобы сервис платежей мог определить плательщика.
		grpc.WithUnaryInterceptor(pkgauth.UnaryClientInterceptor()),
	)
	if err != nil {
		return nil, fmt.Errorf("payment dial %s: %w", addr, err)
	}
	return &Client{conn: conn, client: paymentv1.NewPaymentServiceClient(conn)}, nil
}

// Close shuts down the gRPC connection.
// Close закрывает gRPC-соединение.
func (c *Client) Close() error {
	if c.conn == nil {
		return nil
	}
	return c.conn.Close()
}

// CreatePayment initiates payment for an order and returns the payment ID.
// CreatePayment инициирует оплату заказа и возвращает ID платежа.
//
// Downstream call: payment.PaymentService/CreatePayment — idempotent on duplicate order_id.
// Downstream-вызов: payment.PaymentService/CreatePayment — идемпотентен при дубликате order_id.
func (c *Client) CreatePayment(ctx context.Context, orderID int64, _, _ int64, _ float64) (int64, error) {
	resp, err := c.client.CreatePayment(ctx, &paymentv1.CreatePaymentRequest{
		OrderId: orderID,
	})
	if err != nil {
		return 0, err
	}
	if resp.GetStatus() != paymentv1.PaymentStatus_PAYMENT_STATUS_SUCCESS {
		return 0, fmt.Errorf("payment failed: order_id=%d", orderID)
	}
	return resp.GetPaymentId(), nil
}
