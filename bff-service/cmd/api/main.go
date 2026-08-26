package main

import (
	"log"
	"time"

	"bff-service/internal/client"
	"bff-service/internal/handler"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {

	// mainservice grpc client başlatma
	grpcAddr := "localhost:50051"
	vmClient, err := client.NewClient(grpcAddr)
	if err != nil {
		log.Fatalf("grpcs client başlatılamadı: %v", err)
	}
	defer vmClient.Close()

	//gin http suncusu
	r := gin.Default()

	//cors ayarları (?) ((Frontend'in localhost:3000'den rahatça bağlanabilmesi için)(?))
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	//endpoint yolları
	vmHandler := handler.NewVMHandler(vmClient)
	api := r.Group("/api")
	{
		api.GET("/vms",vmHandler.ListVMs)
		api.POST("/vms/:id/start", vmHandler.StartVM)
		api.POST("/vms/:id/stop", vmHandler.StopVM)
		api.POST("/vms/:id/shutdown", vmHandler.ShutdownVM)
	}

	log.Println("🚀 BFF Servisi 8080 portunda dinliyor: http://localhost:8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("HTTP sunucu hatası: %v", err)
	}
	
}
