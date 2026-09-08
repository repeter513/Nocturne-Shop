package auth

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

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

func UnaryClientInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		return invoker(ForwardAuth(ctx), method, req, reply, cc, opts...)
	}
}
