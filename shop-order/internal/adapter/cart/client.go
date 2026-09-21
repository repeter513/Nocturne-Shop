// Package cart provides a gRPC client for the cart service.
// Пакет cart предоставляет gRPC-клиент для сервиса корзины.
package cart

import (
	"context"
	"fmt"

	cartv1 "github.com/repeter513/shop-proto/gen/go/cart/v1"
	pkgauth "github.com/repeter513/shop-proto/pkg/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Item is a cart line item with product details.
// Item — позиция корзины с данными о товаре.
type Item struct {
	// ProductID is the catalog product identifier.
	// ProductID — идентификатор товара в каталоге.
	ProductID int64
	// Quantity is the number of units in the cart.
	// Quantity — количество единиц в корзине.
	Quantity int32
	// Name is the product display name from catalog enrichment.
	// Name — отображаемое название товара из обогащения каталогом.
	Name string
	// Price is the unit price in minor currency units.
	// Price — цена за единицу в минимальных единицах валюты.
	Price int64
}

// Cart is the user's shopping cart aggregate.
// Cart — агрегат корзины пользователя.
type Cart struct {
	// UserID is the cart owner.
	// UserID — владелец корзины.
	UserID int64
	// Items is the list of product lines in the cart.
	// Items — список товарных позиций в корзине.
	Items []Item
	// TotalPrice is the sum of (Price * Quantity) across all items.
	// TotalPrice — сумма (Price * Quantity) по всем позициям.
	TotalPrice int64
}

// Client calls the remote cart gRPC service.
// Client вызывает удалённый gRPC-сервис корзины.
type Client struct {
	// conn is the persistent gRPC connection to the cart service.
	// conn — постоянное gRPC-соединение с сервисом корзины.
	conn *grpc.ClientConn
	// client is the generated stub for CartService RPCs.
	// client — сгенерированная заглушка для RPC CartService.
	client cartv1.CartServiceClient
}

// New dials the cart service at addr and returns a Client.
// New подключается к сервису корзины по addr и возвращает Client.
func New(_ context.Context, addr string) (*Client, error) {
	conn, err := grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		// Forward JWT from incoming context so cart service knows the user.
		// Проброс JWT из входящего контекста, чтобы сервис корзины знал пользователя.
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

// Close shuts down the gRPC connection.
// Close закрывает gRPC-соединение.
func (c *Client) Close() error {
	if c.conn == nil {
		return nil
	}
	return c.conn.Close()
}

// GetCart returns the authenticated user's cart from context.
// GetCart возвращает корзину аутентифицированного пользователя из контекста.
//
// Downstream call: cart.CartService/GetCart — user is inferred from forwarded JWT.
// Downstream-вызов: cart.CartService/GetCart — пользователь определяется из проброшенного JWT.
func (c *Client) GetCart(ctx context.Context, _ int64) (*Cart, error) {
	resp, err := c.client.GetCart(ctx, &cartv1.GetCartRequest{})
	if err != nil {
		return nil, err
	}
	return toCart(resp.GetCart()), nil
}

// ClearCart removes all items from the user's cart.
// ClearCart удаляет все позиции из корзины пользователя.
//
// Downstream call: cart.CartService/ClearCart — called after successful payment.
// Downstream-вызов: cart.CartService/ClearCart — вызывается после успешной оплаты.
func (c *Client) ClearCart(ctx context.Context, _ int64) error {
	_, err := c.client.ClearCart(ctx, &cartv1.ClearCartRequest{})
	return err
}

// toCart converts a protobuf cart to the local Cart type.
// toCart преобразует protobuf-корзину в локальный тип Cart.
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
