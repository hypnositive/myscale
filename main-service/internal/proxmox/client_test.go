package proxmox_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"main-service/internal/proxmox"
)

func TestClient_ListVMs_Success(t *testing.T) {

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request){

		expectedPath := "/api2/json/nodes/pve-node-1/qemu"
		if r.URL.Path != expectedPath {
			t.Errorf("Beklenen path %s, gelen path %s",expectedPath,r.URL.Path)
		}

		expectedAuth := "PVEAPIToken=test-token-123"
		if r.Header.Get("Authorization") != expectedAuth{
			t.Errorf("Beklenen Auth %s, gelen %s",expectedAuth,r.Header.Get("Authorization"))
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"data": [
					{"vmid": 100, "name": "ubuntu-test", "status": "running"},
					{"vmid": 101, "name": "debian-test", "status": "stopped"}
					]
		}`))
	}))

	defer mockServer.Close()

	client := proxmox.NewClient(mockServer.URL, "pve-node-1", "test-token-123")

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	vms, err := client.ListVMs(ctx)

	if err != nil {
		t.Fatalf("Hata beklenmiyodu ama alındı: %v", err)
	}

	if len(vms) != 2 {
		t.Fatalf("Beklenen VM sayısı 2, gelen %d", len(vms))
	}

	if vms[0].VMID != 100 || vms[0].Name != "ubuntu-test" || vms[0].Status != "running" {
		t.Errorf("İlk VM bilgileri yanlış parse edildi: %+v", vms[0])
	}

	if vms[1].VMID != 101 || vms[1].Name != "debian-test" || vms[1].Status != "stopped" {
		t.Errorf("İkinci VM bilgileri yanlış parse edildi: %+v", vms[0])
	}

}

func TestClient_ListVMs_ServerError(t *testing.T){

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request){
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Sunucu hatası oluştu"))
	}))
	defer mockServer.Close()

	client := proxmox.NewClient(mockServer.URL, "pve-node-1", "testkoo")

	_, err := client.ListVMs(context.Background())
	if err == nil {
		t.Fatal("500 hatasında func error döndürmeliydi")
	}
}


func TestClient_VMAction_Success(t *testing.T){

	commands := []string {"start","stop","shutdown"}

	for _, command := range commands{
		t.Run("action_"+command, func(t *testing.T) {
			mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request){
				
				expectedPath := fmt.Sprintf("/api2/json/nodes/pve-node-1/qemu/100/status/%s",command)
				if r.URL.Path != expectedPath{
					t.Errorf("Beklenen path %s, gelen %s", expectedPath, r.URL.Path)
				}

				if r.Method != http.MethodPost {
					t.Errorf("Beklenen metot POST, gelen %s", r.Method)
				}

				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"data": "UPID:pve-node-1:00001234:`+command+`"}`))
			}))
			defer mockServer.Close()

			client := proxmox.NewClient(mockServer.URL, "pve-node-1","test-token-123")

			taskID, err := client.VMAction(context.Background(), 100, command)
			if err != nil {
				t.Fatalf("Beklenmeyen hata: %v",err)
			}
			
			if taskID != "UPID:pve-node-1:00001234:"+command {
				t.Errorf("dönen taskid uyuşmuor: %s",taskID)
			}
		})

	}
}