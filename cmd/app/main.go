package main

import (
	"context"
	"errors"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	mongorepo "github.com/darkin-kat/user/internal/repository/mongo"
	"github.com/darkin-kat/user/internal/server"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	usrv1 "github.com/darkin-kat/store-api/gen/users/v1"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Config struct {
	MongoURI    string
	MongoDbName string
	GRPCPort    string
}

func loadConfig() Config {
	return Config{
		MongoURI:    getEnv("MONGO_URI", "mongodb://localhost:27017"),
		MongoDbName: getEnv("MONGO_DB_NAME", "project_two"),
		GRPCPort:    getEnv("GRPC_PORT", "50051"),
	}
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func connectToMongo(ctx context.Context, uri string) (*mongo.Client, error) {
	connectCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(connectCtx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}
	err = client.Ping(connectCtx, nil)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func main() {
	cfg := loadConfig()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	mongoClient, err := connectToMongo(ctx, cfg.MongoURI)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer mongoClient.Disconnect(context.Background())

	db := mongoClient.Database(cfg.MongoDbName)
	repo := mongorepo.NewMongoRepository(db)

	srv := server.NewServer(repo)

	grpcServer := grpc.NewServer()
	usrv1.RegisterUserServiceServer(grpcServer, srv)

	reflection.Register(grpcServer) // Register reflection service on gRPC server.

	listener, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		log.Fatalf("Failed to listen on port %s: %v", cfg.GRPCPort, err)
	}

	go func() {
		log.Printf("gRPC user service is listening on port %s", cfg.GRPCPort)
		if err := grpcServer.Serve(listener); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			log.Fatalf("Failed to serve gRPC server: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("shutdown signal received, stopping gracefully...")
	grpcServer.GracefulStop()
	log.Println("gRPC server stopped")
}
