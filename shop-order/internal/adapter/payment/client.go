package payment

import (
	"context"
	"fmt"

	pkgauth "github.com/repeter513/shop-proto/pkg/auth"
	paymentv1 "github.com/repeter513/shop-proto/gen/go/payment/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	conn   *grpc.ClientConn
	client paymentv1.PaymentServiceClient
}

func New(_ context.Context, addr string) (*Client, error) {
	conn, err := grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(pkgauth.UnaryClientInterceptor()),
	)
	if err != nil {
		return nil, fmt.Errorf("payment dial %s: %w", addr, err)
	}
	return &Client{conn: conn, client: paymentv1.NewPaymentServiceClient(conn)}, nil
}

func (c *Client) Close() error {
	if c.conn == nil {
		return nil
	}
	return c.conn.Close()
}

func (c *Client) CreatePayment(ctx context.Context, orderID, userID int64, amount float64) (int64, error) {
	resp, err := c.client.CreatePayment(ctx, &paymentv1.CreatePaymentRequest{
		OrderId: orderID,
		UserId:  userID,
		Amount:  amount,
	})
	if err != nil {
		return 0, err
	}
	if resp.GetStatus() != paymentv1.PaymentStatus_PAYMENT_STATUS_SUCCESS {
		return 0, fmt.Errorf("payment failed: order_id=%d", orderID)
	}
	return resp.GetPaymentId(), nil
}
