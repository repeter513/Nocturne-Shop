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

type Product struct {
	ID     int64
	Name   string
	Price  float64
	Active bool
}

type Client struct {
	conn   *grpc.ClientConn
	client catalogv1.CatalogServiceClient
}

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

func (c *Client) Close() error {
	if c.conn == nil {
		return nil
	}
	return c.conn.Close()
}

func (c *Client) GetProduct(ctx context.Context, productID int64) (*Product, error) {
	resp, err := c.client.GetProduct(ctx, &catalogv1.GetProductRequest{
		ProductId: productID,
	})
	if err != nil {
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
