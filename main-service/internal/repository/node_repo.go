package repository

import (
	"database/sql"
	"errors"
	"main-service/internal/model"
)

type NodeRepository struct {
	db *sql.DB
}

func NewNodeRepository(db *sql.DB) *NodeRepository {
	return &NodeRepository{db: db}
}

func (r *NodeRepository) InitSchema() error {
	query := `
	CREATE TABLE IF NOT EXISTS nodes (
		id SERIAL PRIMARY KEY,
		name VARCHAR(100) NOT NULL,
		host_url VARCHAR(255) NOT NULL UNIQUE,
		node_name VARCHAR(100) NOT NULL,
		token TEXT NOT NULL,
		is_active BOOLEAN DEFAULT true
	);`
	_, err := r.db.Exec(query)
	return err
}

func (r *NodeRepository) Upsert(n model.Node) error {
	query := `
    INSERT INTO nodes (name, host_url, node_name, token, is_active)
    VALUES ($1, $2, $3, $4, $5)
    ON CONFLICT (host_url) DO UPDATE
    SET name = EXCLUDED.name,
        node_name = EXCLUDED.node_name,
        token = EXCLUDED.token,
        is_active = EXCLUDED.is_active;`

	_, err := r.db.Exec(query, n.Name, n.HostURL, n.NodeName, n.Token, n.IsActive)
	return err
}

func (r *NodeRepository) DeactivateOtherNodes(nodeName, keepHostURL string) error {
	query := `
	UPDATE nodes
	SET is_active = false
	WHERE node_name = $1
	  AND host_url <> $2;
	`
	_, err := r.db.Exec(query, nodeName, keepHostURL)
	return err
}

func (r *NodeRepository) GetNodeByHostURL(hostURL string) (*model.Node, error) {
	var n model.Node

	query := `
        SELECT id, name, host_url, node_name, token, is_active
        FROM nodes
        WHERE host_url = $1
    `

	err := r.db.QueryRow(query, hostURL).Scan(&n.ID, &n.Name, &n.HostURL, &n.NodeName, &n.Token, &n.IsActive)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &n, nil
}

func (r *NodeRepository) GetActiveNodes() ([]model.Node, error) {
	rows, err := r.db.Query(`
        SELECT id, name, host_url, node_name, token, is_active
        FROM nodes
        WHERE is_active = true
    `)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var nodes []model.Node
	for rows.Next() {
		var n model.Node
		if err := rows.Scan(&n.ID, &n.Name, &n.HostURL, &n.NodeName, &n.Token, &n.IsActive); err != nil {
			return nil, err
		}
		nodes = append(nodes, n)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return nodes, nil
}

func (r *NodeRepository) GetAllNodes() ([]model.Node, error) {
	rows, err := r.db.Query(`
        SELECT id, name, host_url, node_name, token, is_active
        FROM nodes
    `)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var nodes []model.Node
	for rows.Next() {
		var n model.Node
		if err := rows.Scan(&n.ID, &n.Name, &n.HostURL, &n.NodeName, &n.Token, &n.IsActive); err != nil {
			return nil, err
		}
		nodes = append(nodes, n)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return nodes, nil
}

func (r *NodeRepository) GetNodeByID(id int) (*model.Node, error) {
	var n model.Node

	query := `
        SELECT id, name, host_url, node_name, token, is_active
        FROM nodes
        WHERE id = $1
    `

	err := r.db.QueryRow(query, id).Scan(&n.ID, &n.Name, &n.HostURL, &n.NodeName, &n.Token, &n.IsActive)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &n, nil
}
