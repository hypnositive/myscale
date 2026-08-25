package main

import (
	"database/sql"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"main-service/internal/config"
	"main-service/internal/proxmox"
	"main-service/internal/repository"
	"main-service/internal/service"
	pb "main-service/proto"

	_ "github.com/jackc/pgx/v5/stdlib"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)


func main() {

	cfg := config.Load()
	log.Println("Yapılandırma yüklendi")


	//db
	db, err := sql.Open("pgx", cfg.DBDSN)
	if err != nil {
		log.Fatalf("Veritabanı bağlantı hatası: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Veritabanına ulaşılamıyor: %v", err)
	}
	log.Println("Veritabanı bağlantısı başarılı.")


	//repository katmanları
	nodeRepo := repository.NewNodeRepository(db)
	if err := nodeRepo.InitSchema(); err != nil {
		log.Fatalf("Nodes tablosu oluşturulamadı: %v", err)
	}

	vmRepo := repository.NewVMRepository(db)
		if err := vmRepo.InitSchema(); err != nil {
		log.Fatalf("Vms tablosu oluşturulamadı: %v", err)
	}
	log.Println("Veritabanı şemaları (nodes, vms) hazır.")


	//proxmox havuzu
	clientPool := proxmox.NewClientPool()
	vmService := service.NewVMService(vmRepo,nodeRepo,clientPool) 


	//tcp listener
	lis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		log.Fatalf("Port dinlenemiyor (%s): %v",cfg.GRPCPort, err)
	}

	//grpc sunucusu
	grpcServer := grpc.NewServer()
	pb.RegisterVMServiceServer(grpcServer, vmService)


	//postman veya grpcurl ile kolay test edebilmek için reflectionmış ne olduğu hakkında en ufak bir fikrimd ahi yok
	reflection.Register(grpcServer)

	//graceful shutdown
	go func ()  {
		stopChan := make(chan os.Signal, 1)
		signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)
		<-stopChan

		log.Println("sunucu kapatılıyor...")
		grpcServer.GracefulStop()
		log.Println("Sunucu tereyağından kıl çekermişcesine durdurulmuş öğren burayı aq")
	}()
	
	log.Printf("gRPC Sunucusu :%s portunda dinlemede...\n", cfg.GRPCPort)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("gRPC sunucu hatası: %v", err)
	}


}