package proxmox

import "main-service/internal/model"

// ClientPool: Çoklu sunucular için istemci yönetim katmanı
type ClientPool struct{}

func NewClientPool() *ClientPool {
	return &ClientPool{}
}

// GetClient: Veritabanından gelen model.Node nesnesini alıp ona özel Proxmox Client üretir
func (p *ClientPool) GetClient(node model.Node) *Client {
	return NewClient(node.HostURL, node.NodeName, node.Token)
}