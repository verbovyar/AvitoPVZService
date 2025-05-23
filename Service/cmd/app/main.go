package main

import (
	"AvitoPVZService/Service/config"
	"AvitoPVZService/Service/internal/handlers"
	"AvitoPVZService/Service/internal/repositories/db"
	"AvitoPVZService/Service/internal/repositories/interfaces"
	postgres "AvitoPVZService/Service/pkg"
	"fmt"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"log"
	"net"
	"net/http"
	"path/filepath"
)

func main() {
	fmt.Println("Start server")

	absPath, err := filepath.Abs("config")
	conf, err := config.LoadConfig(absPath)
	if err != nil {
		log.Fatalf("%v", err)
	}

	go startMetrics()
	repo := startRepo(&conf)
	go startGRPC(&conf, repo)
	startHTTP(&conf, repo)
}

func startHTTP(config *config.Config, repo interfaces.Repository) {
	fmt.Println("HTTP server on :8080")
	handler := handlers.NewHttpHandlers(repo)
	log.Fatal(http.ListenAndServe(config.Port, handler))
}

func startRepo(config *config.Config) *db.PostgresRepository {
	pool := postgres.New(config.ConnectingString)
	if pool.Pool == nil {
		fmt.Println("Nil pointer")
	}
	repo := db.New(pool.Pool)

	return repo
}

func startMetrics() {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	log.Println("Metrics on :9000")
	log.Fatal(http.ListenAndServe(":9000", mux))
}

func startGRPC(config *config.Config, repo interfaces.Repository) {
	lis, err := net.Listen(config.NetworkType, config.GrpcPort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	grpcH := handlers.NewGrpcHandlers(repo)
	handlers.RegisterPVZServiceServer(grpcServer, grpcH)
	log.Println("gRPC server listening on :3000")
	grpcServer.Serve(lis)
}
