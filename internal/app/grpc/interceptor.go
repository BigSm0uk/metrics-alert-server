package grpc

import (
	"context"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func SubnetCheckInterceptor(trustedSubnet string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if trustedSubnet == "" {
			return handler(ctx, req)
		}

		_, ipNet, err := net.ParseCIDR(trustedSubnet)
		if err != nil {
			return nil, status.Error(codes.Internal, "invalid trusted subnet configuration")
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.PermissionDenied, "missing metadata")
		}

		ips := md.Get("x-real-ip")
		if len(ips) == 0 {
			return nil, status.Error(codes.PermissionDenied, "missing x-real-ip in metadata")
		}

		clientIP := net.ParseIP(ips[0])
		if clientIP == nil {
			return nil, status.Error(codes.PermissionDenied, "invalid IP address")
		}

		if !ipNet.Contains(clientIP) {
			return nil, status.Error(codes.PermissionDenied, "IP address not in trusted subnet")
		}

		return handler(ctx, req)
	}
}
