// grpc_klog provides a logging interceptor for the klog logger.
package grpc_klog

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
	"k8s.io/klog/v2"
)

// LogGRPC writes a log with the name of every gRPC call or error it receives.
// Request parameters or responses are NOT logged.
func LogGRPC(level klog.Level) func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		// log the requests method name
		var addr string
		peer, ok := peer.FromContext(ctx)
		if ok {
			addr = peer.Addr.String()
		}
		klog.V(level).Infof("GRPC call from peer: %q: %s", addr, info.FullMethod)

		// log errors, if any
		resp, err := handler(ctx, req)
		if err != nil {
			klog.Errorf("GRPC error: %v", err)
		}
		return resp, err
	}
}
