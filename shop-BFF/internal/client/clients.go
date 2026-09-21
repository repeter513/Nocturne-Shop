// Package client dials backend gRPC services for the BFF.
// Пакет client подключается к backend gRPC-сервисам для BFF.
package client

import (
	"context"
	"fmt"

	authv1 "github.com/repeter513/shop-proto/gen/go/auth/v1"
	cartv1 "github.com/repeter513/shop-proto/gen/go/cart/v1"
	catalogv1 "github.com/repeter513/shop-proto/gen/go/catalog/v1"
	orderv1 "github.com/repeter513/shop-proto/gen/go/order/v1"
	paymentv1 "github.com/repeter513/shop-proto/gen/go/payment/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Clients holds gRPC stubs for all backend microservices.
// Clients хранит gRPC-стабы всех backend-микросервисов.
type Clients struct {
	// Auth handles RegisterUser, LoginUser, RefreshToken, ValidateToken, GetUserInfo.
	// Auth обрабатывает RegisterUser, LoginUser, RefreshToken, ValidateToken, GetUserInfo.
	Auth authv1.AuthServiceClient

	// Catalog handles ListProducts, GetProduct, GetStock, ListCategories.
	// Catalog обрабатывает ListProducts, GetProduct, GetStock, ListCategories.
	Catalog catalogv1.CatalogServiceClient

	// Cart handles GetCart, AddToCart, UpdateCartItem, RemoveFromCart, ClearCart.
	// Cart обрабатывает GetCart, AddToCart, UpdateCartItem, RemoveFromCart, ClearCart.
	Cart cartv1.CartServiceClient

	// Order handles CreateOrder, PayOrder, CancelOrder, ListOrders, GetOrder.
	// Order обрабатывает CreateOrder, PayOrder, CancelOrder, ListOrders, GetOrder.
	Order orderv1.OrderServiceClient

	// Payment handles ListPayments, GetPayment.
	// Payment обрабатывает ListPayments, GetPayment.
	Payment paymentv1.PaymentServiceClient

	// conns holds raw connections for Close(); not exposed to handlers.
	// conns хранит сырые соединения для Close(); не экспонируется обработчикам.
	conns []*grpc.ClientConn
}

// New dials all backend services and returns a Clients bundle.
// New подключается ко всем backend-сервисам и возвращает набор Clients.
func New(ctx context.Context, authAddr, catalogAddr, cartAddr, orderAddr, paymentAddr string) (*Clients, error) {
	type target struct {
		addr string
		name string
	}
	// Dial order must match stub assignment below.
	// Порядок подключения должен совпадать с назначением стабов ниже.
	targets := []target{
		{authAddr, "auth"},
		{catalogAddr, "catalog"},
		{cartAddr, "cart"},
		{orderAddr, "order"},
		{paymentAddr, "payment"},
	}

	conns := make([]*grpc.ClientConn, 0, len(targets))
	for _, t := range targets {
		// ponytail: insecure credentials — TLS upgrade path is grpc.WithTransportCredentials(tlsConfig).
		// ponytail: незащищённые credentials — путь апгрейда: grpc.WithTransportCredentials(tlsConfig).
		conn, err := grpc.NewClient(t.addr,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
		if err != nil {
			closeAll(conns)
			return nil, fmt.Errorf("dial %s: %w", t.name, err)
		}
		conns = append(conns, conn)
	}

	return &Clients{
		Auth:    authv1.NewAuthServiceClient(conns[0]),
		Catalog: catalogv1.NewCatalogServiceClient(conns[1]),
		Cart:    cartv1.NewCartServiceClient(conns[2]),
		Order:   orderv1.NewOrderServiceClient(conns[3]),
		Payment: paymentv1.NewPaymentServiceClient(conns[4]),
		conns:   conns,
	}, nil
}

// Close shuts down all gRPC connections.
// Close закрывает все gRPC-соединения.
func (c *Clients) Close() {
	closeAll(c.conns)
}

// closeAll closes every connection in the slice.
// closeAll закрывает каждое соединение в срезе.
func closeAll(conns []*grpc.ClientConn) {
	for _, conn := range conns {
		if conn != nil {
			_ = conn.Close()
		}
	}
}
