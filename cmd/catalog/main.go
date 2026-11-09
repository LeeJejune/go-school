package main

import (
	pb "github.com/LeeJejune/go-school/gen/school/v1"
	"github.com/LeeJejune/go-school/internal/catalog"
	"github.com/LeeJejune/go-school/internal/platform"
	"google.golang.org/grpc"
)

func main() {
	ctx, stop := platform.SignalContext()
	defer stop()
	platform.LogFatal(platform.ServeGRPC(ctx, platform.Env("CATALOG_ADDR", "127.0.0.1:9001"), func(server *grpc.Server) {
		pb.RegisterCatalogServiceServer(server, catalog.New())
	}))
}
