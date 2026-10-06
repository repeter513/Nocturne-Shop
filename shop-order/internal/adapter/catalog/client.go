// Package catalog provides a gRPC client for stock reservations.
// Пакет catalog предоставляет gRPC-клиент для резервирования товара.
package catalog

import (
	"context"
	"fmt"

	catalogv1 "github.com/repeter513/shop-proto/gen/go/catalog/v1"
	pkgauth "github.com/repeter513/shop-proto/pkg/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	conn   *grpc.ClientConn
	client catalogv1.CatalogServiceClient
}

func New(_ context.Context, addr string) (*Client, error) {
	conn, err := grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(pkgauth.UnaryClientInterceptor()),
	)
	if err != nil {
		return nil, fmt.Errorf("catalog dial %s: %w", addr, err)
	}
	return &Client{
		conn:   conn,
		client: catalogv1.NewCatalogServiceClient(conn),
	}, nil
}

func (c *Client) Close() error {
	if c.conn == nil {
		return nil
	}
	return c.conn.Close()
}

func (c *Client) ReserveStock(ctx context.Context, orderID, productID int64, quantity int32) error {
	resp, err := c.client.ReserveStock(ctx, &catalogv1.ReserveStockRequest{
		ProductId: productID,
		Quantity:  quantity,
		OrderId:   orderID,
	})
	if err != nil {
		return err
	}
	r := resp.GetReservation()
	if r == nil || r.GetReservationId() == "" {
		return fmt.Errorf("reserve stock failed: product_id=%d", productID)
	}
	return nil
}

// ReleaseOrder releases all catalog reservations for the order via order_id (no local reservation cache).
// ReleaseOrder снимает резервы catalog по order_id (без локального кэша reservation_id).
func (c *Client) ReleaseOrder(ctx context.Context, orderID int64) error {
	if orderID == 0 {
		return nil
	}
	_, err := c.client.ReleaseStock(ctx, &catalogv1.ReleaseStockRequest{OrderId: orderID})
	return err
}

// ConfirmReservation confirms reservations for the order via order_id.
// ConfirmReservation подтверждает резервы заказа по order_id.
func (c *Client) ConfirmReservation(ctx context.Context, orderID int64) error {
	resp, err := c.client.ConfirmReservation(ctx, &catalogv1.ConfirmReservationRequest{OrderId: orderID})
	if err != nil {
		return err
	}
	if resp.GetReservation() == nil {
		return fmt.Errorf("confirm reservation failed: order_id=%d", orderID)
	}
	return nil
}
