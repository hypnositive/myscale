package model

type ProxmoxVM struct {
	VMID   int    `json:"vmid"`
	Name   string `json:"name"`
	Status string `json:"status"`
	CPUs   int    `json:"cpus"`
	Mem    int64  `json:"mem"`
	MaxMem int64  `json:"maxmem"`
	Uptime int64  `json:"uptime"`
}

type StoredVM struct {
	ID          int    `json:"id"`
	NodeID      int    `json:"node_id"`
	ProxmoxVMID int    `json:"proxmox_vmid"`
	Name        string `json:"name"`
	Status      string `json:"status"`
	NodeName    string `json:"node_name"` // Join sorguları için opsiyonel
}