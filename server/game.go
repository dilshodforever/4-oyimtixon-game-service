package main

import (
	"log"
	"net"
	pb "github.com/dilshodforever/4-oyimtixon-game-service/genprotos/game"
	"github.com/dilshodforever/4-oyimtixon-game-service/service"
	postgres "github.com/dilshodforever/4-oyimtixon-game-service/storage/postgres"
	"google.golang.org/grpc"
)

func main() {
	db, err := postgres.NewMongoConnecti0n()
	if err != nil {
		log.Fatal("Error while connection on db: ", err.Error())
	}
	liss, err := net.Listen("tcp", ":8087")
	if err != nil {
		log.Fatal("Error while connection on tcp: ", err.Error())
	}

	s := grpc.NewServer()
	pb.RegisterGameServiceServer(s, service.NewGameService(db))
	log.Printf("server listening at %v", liss.Addr())
	if err := s.Serve(liss); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
	
}
