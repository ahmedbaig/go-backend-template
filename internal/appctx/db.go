package appctx

import (
	"context"
	"net/http"
	"time"

	"github.com/ahmedbaig/go-backend-db"
	"github.com/ahmedbaig/go-backend-db/dbgrpc"
	"github.com/ahmedbaig/go-backend-http"
	"github.com/ahmedbaig/go-backend-proto/grpcx"
	"github.com/ahmedbaig/go-backend-template/internal/config"
	"github.com/ahmedbaig/go-backend-template/internal/constants"
	"google.golang.org/grpc"
)

func connectDB(h config.DBHandler) (*db.Client, *grpc.ClientConn, error) {
	var opts []db.ClientOption
	if h.Instance != "" {
		opts = append(opts, db.WithInstance(h.Instance))
	}
	if h.Transport == constants.TransportHTTP {
		opts = append(opts, db.WithHTTPClient(&http.Client{Transport: httpx.PropagatingTransport(nil), Timeout: h.Timeout.Std()}))
		if h.APIKey != "" {
			opts = append(opts, db.WithHeader(constants.HeaderAPIKey, h.APIKey))
		}
		return db.NewClient(h.URL, opts...), nil, nil
	}
	return dbgrpc.Dial(h.GRPCAddr, grpcx.ClientOptions{
		ServiceToken: h.APIKey,
		Extra:        []grpc.DialOption{grpc.WithUnaryInterceptor(timeout(h.Timeout.Std()))},
	}, opts...)
}

func timeout(d time.Duration) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		if _, ok := ctx.Deadline(); !ok && d > 0 {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, d)
			defer cancel()
		}
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}
