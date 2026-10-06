// Forwarding Authorization header from incoming to outgoing gRPC metadata.
// Проброс заголовка Authorization из входящих в исходящие gRPC metadata.
package auth

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// ForwardAuth copies the Bearer token from incoming metadata to outgoing context.
// ForwardAuth копирует Bearer-токен из входящих metadata в исходящий context.
// Used by BFF and order service when calling downstream gRPC services.
// Используется BFF и order-сервисом при вызове downstream gRPC-сервисов.
func ForwardAuth(ctx context.Context) context.Context {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ctx
	}
	if vals := md.Get("authorization"); len(vals) > 0 {
		return metadata.AppendToOutgoingContext(ctx, "authorization", vals[0])
	}
	return ctx
}

// UnaryClientInterceptor automatically forwards auth on every outbound gRPC call.
// UnaryClientInterceptor автоматически пробрасывает auth при каждом исходящем gRPC-вызове.
func UnaryClientInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		return invoker(ForwardAuth(ctx), method, req, reply, cc, opts...)
	}
}
