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

type Clients struct {
	Auth    authv1.AuthServiceClient
	Catalog catalogv1.CatalogServiceClient
	Cart    cartv1.CartServiceClient
	Order   orderv1.OrderServiceClient
	Payment paymentv1.PaymentServiceClient
	conns   []*grpc.ClientConn
}

func New(ctx context.Context, authAddr, catalogAddr, cartAddr, orderAddr, paymentAddr string) (*Clients, error) {
	type target struct {
		addr string
		name string
	}
	targets := []target{
		{authAddr, "auth"},
		{catalogAddr, "catalog"},
		{cartAddr, "cart"},
		{orderAddr, "order"},
		{paymentAddr, "payment"},
	}

	conns := make([]*grpc.ClientConn, 0, len(targets))
	for _, t := range targets {
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

func (c *Clients) Close() {
	closeAll(c.conns)
}

func closeAll(conns []*grpc.ClientConn) {
	for _, conn := range conns {
		if conn != nil {
			_ = conn.Close()
		}
	}
}
