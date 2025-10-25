package grpc

import (
	"context"

	"google.golang.org/grpc/metadata"
)

const TenantKey = "tenant"

// TenantInterceptor extracts tenant from gRPC metadata
func TenantInterceptor(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", nil
	}

	values := md.Get(TenantKey)
	if len(values) == 0 {
		return "", nil
	}

	return values[0], nil
}

// AddTenantToContext adds tenant to context
func AddTenantToContext(ctx context.Context, tenant string) context.Context {
	return context.WithValue(ctx, TenantKey, tenant)
}
