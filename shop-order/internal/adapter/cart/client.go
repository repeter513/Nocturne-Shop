package cart

import (
	"context"
	"fmt"

	pkgauth "github.com/repeter513/shop-proto/pkg/auth"
	cartv1 "github.com/repeter513/shop-proto/gen/go/cart/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Item struct {
	ProductID int64
	Quantity  int32
	Name      string
	Price     float64
}

type Cart struct {
	UserID     int64
	Items      []Item
	TotalPrice float64
}

type Client struct {
	conn   *grpc.ClientConn
	client cartv1.CartServiceClient
}

func New(_ context.Context, addr string) (*Client, error) {
	conn, err := grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(pkgauth.UnaryClientInterceptor()),
	)
	if err != nil {
		return nil, fmt.Errorf("cart dial %s: %w", addr, err)
	}
	return &Client{
		conn:   conn,
		client: cartv1.NewCartServiceClient(conn),
	}, nil
}

func (c *Client) Close() error {
	if c.conn == nil {
		return nil
	}
	return c.conn.Close()
}

func (c *Client) GetCart(ctx context.Context, userID int64) (*Cart, error) {
	resp, err := c.client.GetCart(ctx, &cartv1.GetCartRequest{UserId: userID})
	if err != nil {
		return nil, err
	}
	return toCart(resp.GetCart()), nil
}

func (c *Client) ClearCart(ctx context.Context, userID int64) error {
	_, err := c.client.ClearCart(ctx, &cartv1.ClearCartRequest{UserId: userID})
	return err
}

func toCart(c *cartv1.Cart) *Cart {
	if c == nil {
		return &Cart{UserID: 0, Items: nil}
	}
	items := make([]Item, 0, len(c.GetItems()))
	for _, i := range c.GetItems() {
		items = append(items, Item{
			ProductID: i.GetProductId(),
			Quantity:  i.GetQuantity(),
			Name:      i.GetName(),
			Price:     i.GetPrice(),
		})
	}
	return &Cart{
		UserID:     c.GetUserId(),
		Items:      items,
		TotalPrice: c.GetTotalPrice(),
	}
}
