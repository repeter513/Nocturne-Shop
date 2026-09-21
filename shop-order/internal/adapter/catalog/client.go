// Package catalog provides a gRPC client for stock reservations.
// Пакет catalog предоставляет gRPC-клиент для резервирования товара.
package catalog

import (
	"context"
	"fmt"
	"sync"

	catalogv1 "github.com/repeter513/shop-proto/gen/go/catalog/v1"
	pkgauth "github.com/repeter513/shop-proto/pkg/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Client calls the remote catalog gRPC service.
// Client вызывает удалённый gRPC-сервис каталога.
type Client struct {
	// conn is the persistent gRPC connection to the catalog service.
	// conn — постоянное gRPC-соединение с сервисом каталога.
	conn *grpc.ClientConn
	// client is the generated stub for CatalogService RPCs.
	// client — сгенерированная заглушка для RPC CatalogService.
	client catalogv1.CatalogServiceClient
	// res tracks reservation IDs per order for confirm/release orchestration.
	// res отслеживает ID резервирований по заказу для оркестрации confirm/release.
	res sync.Map
}

// New dials the catalog service at addr and returns a Client.
// New подключается к сервису каталога по addr и возвращает Client.
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

// Close shuts down the gRPC connection.
// Close закрывает gRPC-соединение.
func (c *Client) Close() error {
	if c.conn == nil {
		return nil
	}
	return c.conn.Close()
}

// ReserveStock reserves product quantity for an order.
// ReserveStock резервирует количество товара для заказа.
//
// Downstream call: catalog.CatalogService/ReserveStock
// Downstream-вызов: catalog.CatalogService/ReserveStock
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
	c.addReservation(orderID, r.GetReservationId())
	return nil
}

// ReleaseOrder releases all stock reservations for an order.
// ReleaseOrder освобождает все резервирования товара для заказа.
//
// Downstream call: catalog.CatalogService/ReleaseStock (per reservation)
// Downstream-вызов: catalog.CatalogService/ReleaseStock (для каждого резервирования)
func (c *Client) ReleaseOrder(ctx context.Context, orderID int64) error {
	return c.releaseOrder(ctx, orderID)
}

// ConfirmReservation finalizes all reservations for an order.
// ConfirmReservation подтверждает все резервирования для заказа.
//
// Downstream call: catalog.CatalogService/ConfirmReservation (per reservation)
// Downstream-вызов: catalog.CatalogService/ConfirmReservation (для каждого резервирования)
func (c *Client) ConfirmReservation(ctx context.Context, orderID int64) error {
	v, ok := c.res.Load(orderID)
	if !ok {
		return fmt.Errorf("confirm reservation failed: order_id=%d", orderID)
	}
	for _, id := range v.([]string) {
		resp, err := c.client.ConfirmReservation(ctx, &catalogv1.ConfirmReservationRequest{
			ReservationId: id,
		})
		if err != nil {
			return err
		}
		if resp.GetReservation() == nil {
			return fmt.Errorf("confirm reservation failed: reservation_id=%s", id)
		}
	}
	c.res.Delete(orderID)
	return nil
}

// addReservation stores a reservation ID keyed by order ID.
// addReservation сохраняет ID резервирования по ID заказа.
func (c *Client) addReservation(orderID int64, reservationID string) {
	var ids []string
	if v, ok := c.res.Load(orderID); ok {
		ids = v.([]string)
	}
	c.res.Store(orderID, append(ids, reservationID))
}

// releaseOrder releases and removes all reservations for an order.
// releaseOrder освобождает и удаляет все резервирования для заказа.
func (c *Client) releaseOrder(ctx context.Context, orderID int64) error {
	v, ok := c.res.LoadAndDelete(orderID)
	if !ok {
		return nil
	}
	for _, id := range v.([]string) {
		if _, err := c.client.ReleaseStock(ctx, &catalogv1.ReleaseStockRequest{
			ReservationId: id,
		}); err != nil {
			return err
		}
	}
	return nil
}
