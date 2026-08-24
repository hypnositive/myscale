package model

type Node struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	HostURL   string `json:"host_url"`  // Tailscale IP veya Yerel IP (https://...:8006)
	NodeName  string `json:"node_name"` // PVE içindeki adı: pve01
	Token     string `json:"token"`     // PVEAPIToken=...
	IsActive  bool   `json:"is_active"`
}