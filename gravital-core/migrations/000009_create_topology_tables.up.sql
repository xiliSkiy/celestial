-- 创建拓扑表
CREATE TABLE IF NOT EXISTS topologies (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    type VARCHAR(32) NOT NULL,
    scope VARCHAR(32),
    layout_type VARCHAR(32) DEFAULT 'force',
    layout_config JSONB,
    view_config JSONB,
    is_auto_discovery BOOLEAN DEFAULT false,
    discovery_interval INT,
    last_discovery_at TIMESTAMP,
    version INT DEFAULT 1,
    created_by BIGINT,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_topologies_type ON topologies(type);
CREATE INDEX idx_topologies_scope ON topologies(scope);

-- 创建拓扑节点表
CREATE TABLE IF NOT EXISTS topology_nodes (
    id BIGSERIAL PRIMARY KEY,
    topology_id BIGINT NOT NULL REFERENCES topologies(id) ON DELETE CASCADE,
    device_id VARCHAR(64) NOT NULL,
    node_type VARCHAR(32) NOT NULL,
    label VARCHAR(255),
    icon VARCHAR(64),
    position_x FLOAT,
    position_y FLOAT,
    layer INT DEFAULT 0,
    size INT DEFAULT 40,
    color VARCHAR(32),
    shape VARCHAR(32) DEFAULT 'circle',
    properties JSONB,
    is_locked BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(topology_id, device_id)
);

CREATE INDEX idx_topology_nodes_topology_device ON topology_nodes(topology_id, device_id);
CREATE INDEX idx_topology_nodes_layer ON topology_nodes(topology_id, layer);

-- 创建拓扑链路表
CREATE TABLE IF NOT EXISTS topology_links (
    id BIGSERIAL PRIMARY KEY,
    topology_id BIGINT NOT NULL REFERENCES topologies(id) ON DELETE CASCADE,
    source_node_id BIGINT NOT NULL REFERENCES topology_nodes(id) ON DELETE CASCADE,
    target_node_id BIGINT NOT NULL REFERENCES topology_nodes(id) ON DELETE CASCADE,
    link_type VARCHAR(32) NOT NULL,
    source_interface VARCHAR(128),
    target_interface VARCHAR(128),
    bandwidth BIGINT,
    protocol VARCHAR(32),
    status VARCHAR(32) DEFAULT 'unknown',
    utilization FLOAT,
    latency FLOAT,
    packet_loss FLOAT,
    line_style VARCHAR(32) DEFAULT 'solid',
    line_width INT DEFAULT 2,
    color VARCHAR(32),
    label VARCHAR(255),
    properties JSONB,
    discovered_by VARCHAR(32),
    discovered_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_topology_links_topology ON topology_links(topology_id);
CREATE INDEX idx_topology_links_source ON topology_links(source_node_id);
CREATE INDEX idx_topology_links_target ON topology_links(target_node_id);
CREATE INDEX idx_topology_links_status ON topology_links(status);

-- 创建拓扑分组表
CREATE TABLE IF NOT EXISTS topology_groups (
    id BIGSERIAL PRIMARY KEY,
    topology_id BIGINT NOT NULL REFERENCES topologies(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    parent_id BIGINT REFERENCES topology_groups(id),
    position_x FLOAT,
    position_y FLOAT,
    width FLOAT,
    height FLOAT,
    color VARCHAR(32),
    border_color VARCHAR(32),
    is_collapsed BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_topology_groups_topology ON topology_groups(topology_id);
CREATE INDEX idx_topology_groups_parent ON topology_groups(parent_id);

-- 创建拓扑版本历史表
CREATE TABLE IF NOT EXISTS topology_versions (
    id BIGSERIAL PRIMARY KEY,
    topology_id BIGINT NOT NULL REFERENCES topologies(id) ON DELETE CASCADE,
    version INT NOT NULL,
    snapshot JSONB NOT NULL,
    change_description TEXT,
    changed_by BIGINT,
    created_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(topology_id, version)
);

CREATE INDEX idx_topology_versions_topology_version ON topology_versions(topology_id, version);

-- 创建 LLDP 邻居表
CREATE TABLE IF NOT EXISTS lldp_neighbors (
    id BIGSERIAL PRIMARY KEY,
    device_id VARCHAR(64) NOT NULL,
    local_interface VARCHAR(128) NOT NULL,
    neighbor_chassis_id VARCHAR(128),
    neighbor_port_id VARCHAR(128),
    neighbor_system_name VARCHAR(255),
    neighbor_system_desc TEXT,
    neighbor_port_desc VARCHAR(255),
    neighbor_mgmt_addr VARCHAR(64),
    ttl INT,
    discovered_at TIMESTAMP,
    last_seen TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(device_id, local_interface, neighbor_chassis_id, neighbor_port_id)
);

CREATE INDEX idx_lldp_neighbors_device ON lldp_neighbors(device_id);
CREATE INDEX idx_lldp_neighbors_neighbor ON lldp_neighbors(neighbor_chassis_id);

