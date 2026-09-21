// Package catalog provides a gRPC client for the catalog service.
// Пакет catalog предоставляет gRPC-клиент сервиса каталога.
package catalog

import (
	"context"
	"fmt"
	catalogv1 "github.com/repeter513/shop-proto/gen/go/catalog/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

// Product holds catalog product data used by the cart service.
// Product содержит данные товара из каталога для сервиса корзины.
type Product struct {
	// ID is the unique catalog product identifier.
	// ID — уникальный идентификатор товара в каталоге.
	ID int64
	// Name is the human-readable product title.
	// Name — человекочитаемое название товара.
	Name string
	// Price is the unit price in minor currency units.
	// Price — цена за единицу в минимальных единицах валюты.
	Price int64
	// Active indicates whether the product is available for purchase.
	// Active указывает, доступен ли товар для покупки.
	Active bool
}

// Client calls the catalog gRPC service.
// Client вызывает gRPC-сервис каталога.
type Client struct {
	// conn is the persistent gRPC connection to the catalog service.
	// conn — постоянное gRPC-соединение с сервисом каталога.
	conn *grpc.ClientConn
	// client is the generated stub for CatalogService RPCs.
	// client — сгенерированная заглушка для RPC CatalogService.
	client catalogv1.CatalogServiceClient
}

// New dials the catalog service at addr.
// New устанавливает соединение с сервисом каталога по адресу addr.
func New(_ context.Context, addr string) (*Client, error) {
	conn, err := grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("catalog dial %s: %w", addr, err)
	}
	return &Client{
		conn:   conn,
		client: catalogv1.NewCatalogServiceClient(conn),
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

// GetProduct fetches a product by ID; returns nil if not found.
// GetProduct получает товар по ID; возвращает nil, если не найден.
//
// Downstream call: catalog.CatalogService/GetProduct
// Downstream-вызов: catalog.CatalogService/GetProduct
func (c *Client) GetProduct(ctx context.Context, productID int64) (*Product, error) {
	resp, err := c.client.GetProduct(ctx, &catalogv1.GetProductRequest{
		ProductId: productID,
	})
	if err != nil {
		// NotFound from catalog is treated as a missing product, not an error.
		// NotFound от каталога трактуется как отсутствие товара, а не как ошибка.
		if status.Code(err) == codes.NotFound {
			return nil, nil
		}
		return nil, err
	}

	p := resp.GetProduct()
	if p == nil {
		return nil, nil
	}

	return &Product{
		ID:     p.GetId(),
		Name:   p.GetName(),
		Price:  p.GetPrice(),
		Active: p.GetActive(),
	}, nil
}

// GetAvailableStock returns the available stock quantity for a product.
// GetAvailableStock возвращает доступный остаток товара на складе.
//
// Downstream call: catalog.CatalogService/GetStock
// Downstream-вызов: catalog.CatalogService/GetStock
func (c *Client) GetAvailableStock(ctx context.Context, productID int64) (int32, error) {
	resp, err := c.client.GetStock(ctx, &catalogv1.GetStockRequest{
		ProductId: productID,
	})
	if err != nil {
		return 0, err
	}
	if resp.GetStock() == nil {
		return 0, nil
	}
	return resp.GetStock().GetQuantity(), nil
}
