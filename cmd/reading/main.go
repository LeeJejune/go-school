package main

import (
	pb "github.com/LeeJejune/go-school/gen/school/v1"
	"github.com/LeeJejune/go-school/internal/catalogclient"
	"github.com/LeeJejune/go-school/internal/mysqlstore"
	"github.com/LeeJejune/go-school/internal/platform"
	"github.com/LeeJejune/go-school/internal/reading"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() { platform.LogFatal(run()) }

func run() error {
	ctx, stop := platform.SignalContext()
	defer stop()
	db, err := platform.OpenMySQL(ctx, platform.Env("MYSQL_DSN", "school:school@tcp(127.0.0.1:3307)/school?parseTime=true&loc=UTC&timeout=3s&readTimeout=3s&writeTimeout=3s"))
	if err != nil {
		return err
	}
	defer db.Close()
	conn, err := grpc.NewClient(platform.Env("CATALOG_TARGET", "127.0.0.1:9001"), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	defer conn.Close()
	service := reading.New(mysqlstore.New(db), catalogclient.New(pb.NewCatalogServiceClient(conn)))
	return platform.ServeGRPC(ctx, platform.Env("READING_ADDR", "127.0.0.1:9002"), func(server *grpc.Server) {
		pb.RegisterReadingServiceServer(server, service)
	})
}
