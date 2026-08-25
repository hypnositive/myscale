package repository

import (
	"database/sql"
	"errors"
	"main-service/internal/model"
)

type VMRepository struct {
	db *sql.DB
}

func NewVMRepository(db *sql.DB) *VMRepository {
	return &VMRepository{db: db}
}

func (r *VMRepository) InitSchema() error {
	query := `
	CREATE TABLE IF NOT EXISTS vms (
		id SERIAL PRIMARY KEY,
		node_id INT NOT NULL REFERENCES nodes(id),
		proxmox_vmid INT NOT NULL,
		name VARCHAR(100) NOT NULL,
		status VARCHAR(50) NOT NULL,
		UNIQUE(node_id, proxmox_vmid)
	);`
	_, err := r.db.Exec(query)
	return err
}

func (r *VMRepository) Upsert(vm model.StoredVM) error {
	query := `
	INSERT INTO vms (node_id, proxmox_vmid, name, status)
	VALUES ($1, $2, $3, $4)
	ON CONFLICT (node_id, proxmox_vmid) DO UPDATE
	SET name = EXCLUDED.name,
		status = EXCLUDED.status;`
	_, err := r.db.Exec(query, vm.NodeID, vm.ProxmoxVMID, vm.Name, vm.Status)
	return err
}

func (r *VMRepository) GetAll() ([]model.StoredVM, error) {
	query := `
	SELECT v.id, v.node_id, v.proxmox_vmid, v.name, v.status, n.name
	FROM vms v
	JOIN nodes n ON v.node_id = n.id
	ORDER BY n.name ASC, v.proxmox_vmid ASC;`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []model.StoredVM
	for rows.Next() {
		var vm model.StoredVM
		if err := rows.Scan(&vm.ID, &vm.NodeID, &vm.ProxmoxVMID, &vm.Name, &vm.Status, &vm.NodeName); err != nil {
			return nil, err
		}
		list = append(list, vm)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return list, nil

}

func (r *VMRepository) GetByID(id int) (*model.StoredVM, error) {
	var vm model.StoredVM

	query := `
	SELECT v.id, v.node_id, v.proxmox_vmid, v.name, v.status, n.name
	FROM vms v
	JOIN nodes n ON v.node_id = n.id
	WHERE v.id = $1
	`

	err := r.db.QueryRow(query, id).Scan(
		&vm.ID,
		&vm.NodeID,
		&vm.ProxmoxVMID,
		&vm.Name,
		&vm.Status,
		&vm.NodeName,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &vm, nil
}

func (r *VMRepository) Delete(id int) error {
	query := `DELETE FROM vms WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}