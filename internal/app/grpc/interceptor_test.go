package grpc

import (
	"context"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestSubnetCheckInterceptor_EmptySubnet(t *testing.T) {
	interceptor := SubnetCheckInterceptor("")
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "success", nil
	}

	ctx := context.Background()
	resp, err := interceptor(ctx, nil, &grpc.UnaryServerInfo{}, handler)

	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
	if resp != "success" {
		t.Errorf("expected 'success', got: %v", resp)
	}
}

func TestSubnetCheckInterceptor_ValidIP(t *testing.T) {
	interceptor := SubnetCheckInterceptor("192.168.1.0/24")
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "success", nil
	}

	md := metadata.New(map[string]string{"x-real-ip": "192.168.1.10"})
	ctx := metadata.NewIncomingContext(context.Background(), md)

	resp, err := interceptor(ctx, nil, &grpc.UnaryServerInfo{}, handler)

	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
	if resp != "success" {
		t.Errorf("expected 'success', got: %v", resp)
	}
}

func TestSubnetCheckInterceptor_InvalidIP(t *testing.T) {
	interceptor := SubnetCheckInterceptor("192.168.1.0/24")
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "success", nil
	}

	md := metadata.New(map[string]string{"x-real-ip": "10.0.0.1"})
	ctx := metadata.NewIncomingContext(context.Background(), md)

	_, err := interceptor(ctx, nil, &grpc.UnaryServerInfo{}, handler)

	if err == nil {
		t.Error("expected error, got nil")
	}
	st, ok := status.FromError(err)
	if !ok {
		t.Error("expected status error")
	}
	if st.Code() != codes.PermissionDenied {
		t.Errorf("expected PermissionDenied, got: %v", st.Code())
	}
}

func TestSubnetCheckInterceptor_MissingMetadata(t *testing.T) {
	interceptor := SubnetCheckInterceptor("192.168.1.0/24")
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "success", nil
	}

	ctx := context.Background()
	_, err := interceptor(ctx, nil, &grpc.UnaryServerInfo{}, handler)

	if err == nil {
		t.Error("expected error, got nil")
	}
	st, ok := status.FromError(err)
	if !ok {
		t.Error("expected status error")
	}
	if st.Code() != codes.PermissionDenied {
		t.Errorf("expected PermissionDenied, got: %v", st.Code())
	}
}

func TestSubnetCheckInterceptor_MissingXRealIP(t *testing.T) {
	interceptor := SubnetCheckInterceptor("192.168.1.0/24")
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "success", nil
	}

	md := metadata.New(map[string]string{})
	ctx := metadata.NewIncomingContext(context.Background(), md)

	_, err := interceptor(ctx, nil, &grpc.UnaryServerInfo{}, handler)

	if err == nil {
		t.Error("expected error, got nil")
	}
	st, ok := status.FromError(err)
	if !ok {
		t.Error("expected status error")
	}
	if st.Code() != codes.PermissionDenied {
		t.Errorf("expected PermissionDenied, got: %v", st.Code())
	}
}

func TestSubnetCheckInterceptor_InvalidIPFormat(t *testing.T) {
	interceptor := SubnetCheckInterceptor("192.168.1.0/24")
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "success", nil
	}

	md := metadata.New(map[string]string{"x-real-ip": "invalid-ip"})
	ctx := metadata.NewIncomingContext(context.Background(), md)

	_, err := interceptor(ctx, nil, &grpc.UnaryServerInfo{}, handler)

	if err == nil {
		t.Error("expected error, got nil")
	}
	st, ok := status.FromError(err)
	if !ok {
		t.Error("expected status error")
	}
	if st.Code() != codes.PermissionDenied {
		t.Errorf("expected PermissionDenied, got: %v", st.Code())
	}
}

func TestSubnetCheckInterceptor_InvalidCIDR(t *testing.T) {
	interceptor := SubnetCheckInterceptor("invalid-cidr")
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "success", nil
	}

	md := metadata.New(map[string]string{"x-real-ip": "192.168.1.10"})
	ctx := metadata.NewIncomingContext(context.Background(), md)

	_, err := interceptor(ctx, nil, &grpc.UnaryServerInfo{}, handler)

	if err == nil {
		t.Error("expected error, got nil")
	}
	st, ok := status.FromError(err)
	if !ok {
		t.Error("expected status error")
	}
	if st.Code() != codes.Internal {
		t.Errorf("expected Internal, got: %v", st.Code())
	}
}
