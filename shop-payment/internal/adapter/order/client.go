// Package order provides a gRPC client for the order service.
// Пакет order предоставляет gRPC-клиент сервиса заказов.
package order

import (
	"context"
	"fmt"

	orderv1 "github.com/repeter513/shop-proto/gen/go/order/v1"
	pkgauth "github.com/repeter513/shop-proto/pkg/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Client calls the order gRPC service.
// Client вызывает gRPC-сервис заказов.
type Client struct {
	// conn is the persistent gRPC connection to the order service.
	// conn — постоянное gRPC-соединение с сервисом заказов.
	conn *grpc.ClientConn
	// client is the generated stub for OrderService RPCs.
	// client — сгенерированная заглушка для RPC OrderService.
	client orderv1.OrderServiceClient
}

// New dials the order service at addr.
// New устанавливает соединение с сервисом заказов по адресу addr.
func New(_ context.Context, addr string) (*Client, error) {
	conn, err := grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		// Forward JWT from incoming context to order service for auth.
		// Проброс JWT из входящего контекста в сервис заказов для аутентификации.
		grpc.WithUnaryInterceptor(pkgauth.UnaryClientInterceptor()),
	)
	if err != nil {
		return nil, fmt.Errorf("order dial %s: %w", addr, err)
	}
	return &Client{
		conn:   conn,
		client: orderv1.NewOrderServiceClient(conn),
	}, nil
}

// Close closes the gRPC connection.
// Close закрывает gRPC-соединение.
func (c *Client) Close() error {
	if c.conn == nil {
		return nil
	}
	return c.conn.Close()
}

// OrderTotal returns the total price of an order.
// OrderTotal возвращает итоговую сумму заказа.
//
// Downstream call: order.OrderService/GetOrder — reads TotalPrice from the order aggregate.
// Downstream-вызов: order.OrderService/GetOrder — читает TotalPrice из агрегата заказа.
func (c *Client) OrderTotal(ctx context.Context, orderID int64) (int64, error) {
	resp, err := c.client.GetOrder(ctx, &orderv1.GetOrderRequest{OrderId: orderID})
	if err != nil {
		return 0, err
	}
	o := resp.GetOrder()
	if o == nil {
		return 0, fmt.Errorf("order not found: %d", orderID)
	}
	return o.GetTotalPrice(), nil
}
