package service

import (
	"context"
	"fmt"

	"main-service/internal/proxmox"
	"main-service/internal/repository"
	pb "main-service/proto"

)


type VMService struct {
	pb.UnimplementedVMServiceServer
	vmRepo *repository.VMRepository
	nodeRepo *repository.NodeRepository
	pool *proxmox.ClientPool
}

func NewVMService(
	vmRepo *repository.VMRepository, 
	nodeRepo *repository.NodeRepository, 
	pool *proxmox.ClientPool,
	) *VMService {
		return &VMService{
			vmRepo: vmRepo,
			nodeRepo: nodeRepo,
			pool: pool,
		}
}

func (s *VMService) ListVMs(ctx context.Context, req *pb.ListVMsRequest) (*pb.ListVMsResponse, error) {
	storedVMs, err := s.vmRepo.GetAll()
	if err != nil {
		return nil, fmt.Errorf("failed to get VMs from repository: %v", err)
	}

	var protoVMs []*pb.VM
	for _, vm := range storedVMs {
		protoVMs = append(protoVMs, &pb.VM{
			Id:          int32(vm.ID),
			NodeId:      int32(vm.NodeID),
			ProxmoxVmid: int32(vm.ProxmoxVMID),
			Name:        vm.Name,
			Status:      vm.Status,
			NodeName:    vm.NodeName,
		})
	}

	return &pb.ListVMsResponse{Vms: protoVMs}, nil
}

func (s *VMService) StartVM(ctx context.Context, req *pb.VMActionRequest) (*pb.VMActionResponse, error) {
	return s.executeVMAction(ctx, int(req.GetVmId()), "start")
}

func (s *VMService) StopVM(ctx context.Context, req *pb.VMActionRequest) (*pb.VMActionResponse, error) {
	return s.executeVMAction(ctx, int(req.GetVmId()), "stop")
}

func (s *VMService) ShutdownVM(ctx context.Context, req *pb.VMActionRequest) (*pb.VMActionResponse, error) {
	return s.executeVMAction(ctx, int(req.GetVmId()), "shutdown")
}

func (s *VMService) executeVMAction(ctx context.Context, vmID int, action string) (*pb.VMActionResponse, error) {

	vm, err := s.vmRepo.GetByID(vmID)
	if vm == nil {
		return &pb.VMActionResponse{
			Success: false,
			Message: fmt.Sprintf("ID'ye ait VM bulunamadı (ID: %d): %v", vmID, err),
		}, nil
	}
	if err != nil {
		return &pb.VMActionResponse{
			Success: false,
			Message: fmt.Sprintf("VM veritabanında bulunamadı (ID: %d): %v", vmID, err),
		}, nil
	}

	node, err := s.nodeRepo.GetNodeByID(vm.NodeID)
	if err != nil {
		return &pb.VMActionResponse{
			Success: false,
			Message: fmt.Sprintf("Node bilgisi bulunamadı (NodeID: %d): %v", vm.NodeID, err),
		}, nil
	}

	client := s.pool.GetClient(*node)

	upid, err := client.VMAction(ctx, vm.ProxmoxVMID, action)
	if err != nil {
		return &pb.VMActionResponse{
			Success: false,
			Message: fmt.Sprintf("Proxmox işlemi başarısız (%s): %v", action, err),
		}, nil
	}

	return &pb.VMActionResponse{
		Success:  true,
		Message:  fmt.Sprintf("VM başarıyla %s moduna geçiriliyor.", action),
		TaskUpid: upid,
	}, nil

}