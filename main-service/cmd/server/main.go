package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"main-service/internal/model"
	"main-service/internal/proxmox"
	"main-service/internal/repository"

	_ "github.com/jackc/pgx/v5/stdlib"
)

const (
	dbDSN        = "postgres://postgres:secret@localhost:5432/mydatabase?sslmode=disable"
	proxmoxHost  = "https://100.79.192.82:8006"
	proxmoxNode  = "pve01"
	proxmoxToken = "root@pam!gobackend=27385536-8742-42a8-9754-b98f10a27d59"
	proxmoxName  = "primary-pve"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	db, err := sql.Open("pgx", dbDSN)
	if err != nil {
		log.Fatalf("DB bağlantısı açılamadı: %v", err)
	}
	defer db.Close()

	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("DB ping hatası: %v", err)
	}
	fmt.Println("[+] PostgreSQL bağlantısı başarılı.")

	nodeRepo := repository.NewNodeRepository(db)
	vmRepo := repository.NewVMRepository(db)

	if err := nodeRepo.InitSchema(); err != nil {
		log.Fatalf("Node tablosu oluşturulamadı: %v", err)
	}
	if err := vmRepo.InitSchema(); err != nil {
		log.Fatalf("VM tablosu oluşturulamadı: %v", err)
	}
	fmt.Println("[+] Schema hazır.")

	node := model.Node{
		Name:     proxmoxName,
		HostURL:  proxmoxHost,
		NodeName: proxmoxNode,
		Token:    proxmoxToken,
		IsActive: false,
	}

	if err := nodeRepo.Upsert(node); err != nil {
		log.Fatalf("Node kaydı yazılamadı: %v", err)
	}
	fmt.Println("[+] Node kaydı yazıldı (başlangıç: inactive).")

	client := proxmox.NewClient(node.HostURL, node.NodeName, node.Token)
	vms, err := client.ListVMs(ctx)
	if err != nil {
		node.IsActive = false
		if upsertErr := nodeRepo.Upsert(node); upsertErr != nil {
			log.Fatalf("Node inactive güncellemesi de başarısız: %v (asıl hata: %v)", upsertErr, err)
		}
		log.Fatalf("Proxmox VM listesi çekilemedi: %v", err)
	}

	node.IsActive = true
	if err := nodeRepo.Upsert(node); err != nil {
		log.Fatalf("Node active güncellenemedi: %v", err)
	}
	if err := nodeRepo.DeactivateOtherNodes(node.NodeName, node.HostURL); err != nil {
		log.Fatalf("Aynı node_name için eski host kayıtları pasifleştirilemedi: %v", err)
	}

	storedNode, err := nodeRepo.GetNodeByHostURL(node.HostURL)
	if err != nil {
		log.Fatalf("Node kaydı host_url ile çekilemedi: %v", err)
	}
	if storedNode == nil {
		log.Fatalf("Node kaydı bulunamadı: host=%s", node.HostURL)
	}

	fmt.Printf("[+] Proxmox'tan %d VM listelendi.\n", len(vms))

	var upsertErrCount int
	for _, vm := range vms {
		if err := vmRepo.Upsert(model.StoredVM{
			NodeID:      storedNode.ID,
			ProxmoxVMID: vm.VMID,
			Name:        vm.Name,
			Status:      vm.Status,
			NodeName:    storedNode.NodeName,
		}); err != nil {
			upsertErrCount++
			log.Printf("VM upsert hatası (vmid=%d): %v", vm.VMID, err)
		}
	}
	fmt.Printf("[+] VM kayıtları DB'ye işlendi. başarılı=%d, hatalı=%d\n", len(vms)-upsertErrCount, upsertErrCount)

	allVMs, err := vmRepo.GetAll()
	if err != nil {
		log.Fatalf("DB'den VM'ler çekilemedi: %v", err)
	}

	fmt.Println("\n--- DB'deki VM'ler ---")
	for _, vm := range allVMs {
		fmt.Printf("ID=%d | Node=%s | ProxmoxVMID=%d | Name=%s | Status=%s\n",
			vm.ID, vm.NodeName, vm.ProxmoxVMID, vm.Name, vm.Status)
	}
	fmt.Println("----------------------")

	if len(vms) == 0 {
		fmt.Println("[!] Proxmox'tan hiç VM dönmedi. Node/token bilgilerini kontrol et.")
		return
	}
	fmt.Printf("[+] Proxmox'tan ilk VM örneği: %s (vmid=%d, status=%s)\n", vms[0].Name, vms[0].VMID, vms[0].Status)
}
