package lldp

import (
	"bytes"
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/celestial/orbital-sentinels/internal/plugin"
	"github.com/celestial/orbital-sentinels/sdk"
	"github.com/gosnmp/gosnmp"
	"golang.org/x/crypto/ssh"
)

// LLDPPlugin LLDP 插件
type LLDPPlugin struct {
	sdk.BasePlugin
	schema plugin.PluginSchema
}

// LLDPNeighbor LLDP 邻居信息
type LLDPNeighbor struct {
	LocalInterface     string
	NeighborChassisID  string
	NeighborPortID     string
	NeighborSystemName string
	NeighborSystemDesc string
	NeighborPortDesc   string
	NeighborMgmtAddr   string
	TTL                int
}

// NewPlugin 创建插件实例
func NewPlugin() plugin.Plugin {
	return &LLDPPlugin{}
}

// Meta 返回插件元信息
func (p *LLDPPlugin) Meta() plugin.PluginMeta {
	return p.schema.Meta
}

// Schema 返回配置 Schema
func (p *LLDPPlugin) Schema() plugin.PluginSchema {
	return p.schema
}

// Init 初始化插件
func (p *LLDPPlugin) Init(config map[string]interface{}) error {
	// 加载 schema
	p.schema = plugin.PluginSchema{
		Meta: plugin.PluginMeta{
			Name:        "lldp",
			Version:     "1.0.0",
			Description: "LLDP (Link Layer Discovery Protocol) 邻居发现插件",
			Author:      "Celestial Team",
			DeviceTypes: []string{"switch", "router", "network_device"},
		},
		DeviceFields: []plugin.DeviceField{
			{
				Name:        "host",
				Type:        "string",
				Required:    true,
				Description: "设备 IP 地址或主机名",
			},
			{
				Name:        "protocol",
				Type:        "string",
				Required:    false,
				Default:     "snmp",
				Description: "采集协议 (snmp/ssh)",
			},
			{
				Name:        "snmp_community",
				Type:        "string",
				Required:    false,
				Default:     "public",
				Description: "SNMP Community (当 protocol=snmp 时)",
			},
			{
				Name:        "snmp_version",
				Type:        "string",
				Required:    false,
				Default:     "2c",
				Description: "SNMP 版本 (1/2c/3)",
			},
			{
				Name:        "ssh_username",
				Type:        "string",
				Required:    false,
				Description: "SSH 用户名 (当 protocol=ssh 时)",
			},
			{
				Name:        "ssh_password",
				Type:        "password",
				Required:    false,
				Description: "SSH 密码 (当 protocol=ssh 时)",
			},
		},
	}

	return nil
}

// ValidateConfig 验证设备配置
func (p *LLDPPlugin) ValidateConfig(deviceConfig map[string]interface{}) error {
	// 检查必填字段
	if _, ok := deviceConfig["host"]; !ok {
		return fmt.Errorf("host is required")
	}

	protocol := p.getString(deviceConfig, "protocol", "snmp")
	if protocol != "snmp" && protocol != "ssh" {
		return fmt.Errorf("unsupported protocol: %s, must be 'snmp' or 'ssh'", protocol)
	}

	return nil
}

// TestConnection 测试连接
func (p *LLDPPlugin) TestConnection(deviceConfig map[string]interface{}) error {
	protocol := p.getString(deviceConfig, "protocol", "snmp")

	switch protocol {
	case "snmp":
		return p.testSNMPConnection(deviceConfig)
	case "ssh":
		return p.testSSHConnection(deviceConfig)
	default:
		return fmt.Errorf("unsupported protocol: %s", protocol)
	}
}

// Collect 采集数据
func (p *LLDPPlugin) Collect(ctx context.Context, task *plugin.CollectionTask) ([]*plugin.Metric, error) {
	protocol := p.getString(task.DeviceConfig, "protocol", "snmp")

	var neighbors []LLDPNeighbor
	var err error

	switch protocol {
	case "snmp":
		neighbors, err = p.collectViaSNMP(ctx, task)
	case "ssh":
		neighbors, err = p.collectViaSSH(ctx, task)
	default:
		return nil, fmt.Errorf("unsupported protocol: %s", protocol)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to collect LLDP neighbors: %w", err)
	}

	// 将 LLDP 邻居信息编码到 metrics 的 labels 中
	// 这样可以通过 metrics 传递 LLDP 数据
	metrics := make([]*plugin.Metric, 0, len(neighbors))
	for i, neighbor := range neighbors {
		metric := &plugin.Metric{
			Name:      "lldp_neighbor",
			Value:     float64(i + 1), // 使用序号作为值
			Timestamp: time.Now().Unix(),
			Labels: map[string]string{
				"device_id":            task.DeviceID,
				"local_interface":      neighbor.LocalInterface,
				"neighbor_chassis_id":  neighbor.NeighborChassisID,
				"neighbor_port_id":     neighbor.NeighborPortID,
				"neighbor_system_name": neighbor.NeighborSystemName,
				"neighbor_system_desc": neighbor.NeighborSystemDesc,
				"neighbor_port_desc":   neighbor.NeighborPortDesc,
				"neighbor_mgmt_addr":   neighbor.NeighborMgmtAddr,
				"ttl":                  strconv.Itoa(neighbor.TTL),
			},
			Type: plugin.MetricTypeGauge,
		}
		metrics = append(metrics, metric)
	}

	// 如果没有邻居，返回一个空指标表示采集成功但无数据
	if len(metrics) == 0 {
		metrics = append(metrics, &plugin.Metric{
			Name:      "lldp_neighbor_count",
			Value:     0,
			Timestamp: time.Now().Unix(),
			Labels: map[string]string{
				"device_id": task.DeviceID,
			},
			Type: plugin.MetricTypeGauge,
		})
	}

	return metrics, nil
}

// Close 关闭插件
func (p *LLDPPlugin) Close() error {
	return nil
}

// collectViaSNMP 通过 SNMP 采集 LLDP 邻居
func (p *LLDPPlugin) collectViaSNMP(ctx context.Context, task *plugin.CollectionTask) ([]LLDPNeighbor, error) {
	// 获取配置
	host := p.getString(task.DeviceConfig, "host", "")
	if host == "" {
		return nil, fmt.Errorf("host is required")
	}

	port := p.getInt(task.DeviceConfig, "port", 161)
	snmpVersion := p.getString(task.DeviceConfig, "snmp_version", "2c")

	// 从 connection_config 中获取认证信息
	var community string
	var snmpv3Config map[string]interface{}

	// 尝试从新的 auth 结构读取
	if authRaw, ok := task.DeviceConfig["auth"]; ok {
		if authMap, ok := authRaw.(map[string]interface{}); ok {
			if configRaw, ok := authMap["config"]; ok {
				if configMap, ok := configRaw.(map[string]interface{}); ok {
					if snmpVersion == "3" {
						snmpv3Config = configMap
					} else {
						if comm, ok := configMap["community"].(string); ok {
							community = comm
						}
					}
				}
			}
		}
	}

	// 兼容旧格式：从 device_config 直接读取
	if community == "" && snmpVersion != "3" {
		community = p.getString(task.DeviceConfig, "snmp_community", "public")
	}

	// 创建 SNMP 客户端
	snmp := &gosnmp.GoSNMP{
		Target:  host,
		Port:    uint16(port),
		Timeout: time.Duration(10) * time.Second,
		Retries: 3,
		Context: ctx,
	}

	// 设置 SNMP 版本和认证
	switch snmpVersion {
	case "1":
		snmp.Version = gosnmp.Version1
		snmp.Community = community
	case "2c":
		snmp.Version = gosnmp.Version2c
		snmp.Community = community
	case "3":
		snmp.Version = gosnmp.Version3
		snmp.SecurityModel = gosnmp.UserSecurityModel

		// SNMP v3 配置
		username := p.getString(snmpv3Config, "username", "")
		securityLevel := p.getString(snmpv3Config, "security_level", "noAuthNoPriv")

		usmSecurityParameters := &gosnmp.UsmSecurityParameters{
			UserName: username,
		}

		// 根据安全级别设置认证和加密
		switch securityLevel {
		case "authNoPriv", "authPriv":
			authProtocol := p.getString(snmpv3Config, "auth_protocol", "MD5")
			authPassword := p.getString(snmpv3Config, "auth_password", "")

			switch authProtocol {
			case "MD5":
				usmSecurityParameters.AuthenticationProtocol = gosnmp.MD5
			case "SHA":
				usmSecurityParameters.AuthenticationProtocol = gosnmp.SHA
			case "SHA224":
				usmSecurityParameters.AuthenticationProtocol = gosnmp.SHA224
			case "SHA256":
				usmSecurityParameters.AuthenticationProtocol = gosnmp.SHA256
			case "SHA384":
				usmSecurityParameters.AuthenticationProtocol = gosnmp.SHA384
			case "SHA512":
				usmSecurityParameters.AuthenticationProtocol = gosnmp.SHA512
			default:
				usmSecurityParameters.AuthenticationProtocol = gosnmp.MD5
			}
			usmSecurityParameters.AuthenticationPassphrase = authPassword
		}

		if securityLevel == "authPriv" {
			privProtocol := p.getString(snmpv3Config, "priv_protocol", "DES")
			privPassword := p.getString(snmpv3Config, "priv_password", "")

			switch privProtocol {
			case "DES":
				usmSecurityParameters.PrivacyProtocol = gosnmp.DES
			case "AES":
				usmSecurityParameters.PrivacyProtocol = gosnmp.AES
			case "AES192":
				usmSecurityParameters.PrivacyProtocol = gosnmp.AES192
			case "AES256":
				usmSecurityParameters.PrivacyProtocol = gosnmp.AES256
			default:
				usmSecurityParameters.PrivacyProtocol = gosnmp.DES
			}
			usmSecurityParameters.PrivacyPassphrase = privPassword
		}

		snmp.SecurityParameters = usmSecurityParameters
		snmp.MsgFlags = p.getSNMPv3MsgFlags(securityLevel)
	default:
		return nil, fmt.Errorf("unsupported SNMP version: %s", snmpVersion)
	}

	// 连接
	if err := snmp.Connect(); err != nil {
		return nil, fmt.Errorf("failed to connect to SNMP device: %w", err)
	}
	defer snmp.Conn.Close()

	// LLDP MIB OIDs
	// lldpRemTable: 1.0.8802.1.1.2.1.4.1
	// lldpRemLocalPortNum: 1.0.8802.1.1.2.1.4.1.1.2
	// lldpRemChassisIdSubtype: 1.0.8802.1.1.2.1.4.1.1.3
	// lldpRemChassisId: 1.0.8802.1.1.2.1.4.1.1.4
	// lldpRemPortIdSubtype: 1.0.8802.1.1.2.1.4.1.1.5
	// lldpRemPortId: 1.0.8802.1.1.2.1.4.1.1.6
	// lldpRemSysName: 1.0.8802.1.1.2.1.4.1.1.9
	// lldpRemSysDesc: 1.0.8802.1.1.2.1.4.1.1.10
	// lldpRemPortDesc: 1.0.8802.1.1.2.1.4.1.1.8
	// lldpRemManAddr: 1.0.8802.1.1.2.1.4.2.1.5

	baseOID := "1.0.8802.1.1.2.1.4.1.1"
	neighbors := make([]LLDPNeighbor, 0)

	// 获取本地端口号索引
	portNumOID := baseOID + ".2"
	err := snmp.BulkWalk(portNumOID, func(pdu gosnmp.SnmpPDU) error {
		// 从 OID 中提取索引
		oid := pdu.Name
		parts := strings.Split(oid, ".")
		if len(parts) < 2 {
			return nil
		}

		// 构建邻居信息的 OID 索引
		remTimeMark := parts[len(parts)-2]
		remLocalPortNum := parts[len(parts)-1]

		// 获取邻居信息
		neighbor, err := p.getLLDPNeighborViaSNMP(snmp, remTimeMark, remLocalPortNum)
		if err == nil && neighbor != nil {
			neighbors = append(neighbors, *neighbor)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk LLDP table: %w", err)
	}

	return neighbors, nil
}

// getLLDPNeighborViaSNMP 获取单个 LLDP 邻居信息
func (p *LLDPPlugin) getLLDPNeighborViaSNMP(snmp *gosnmp.GoSNMP, remTimeMark, remLocalPortNum string) (*LLDPNeighbor, error) {
	baseOID := "1.0.8802.1.1.2.1.4.1.1"
	index := remTimeMark + "." + remLocalPortNum

	// 获取本地接口名称
	localPortNumOID := "1.0.8802.1.1.2.1.3.7.1.3." + remLocalPortNum
	localPortResult, err := snmp.Get([]string{localPortNumOID})
	if err != nil || len(localPortResult.Variables) == 0 {
		return nil, fmt.Errorf("failed to get local port: %w", err)
	}
	localInterface := p.formatSNMPValue(localPortResult.Variables[0].Value)

	// 获取邻居 Chassis ID
	chassisIDOID := baseOID + ".4." + index
	chassisIDResult, err := snmp.Get([]string{chassisIDOID})
	if err != nil || len(chassisIDResult.Variables) == 0 {
		return nil, fmt.Errorf("failed to get chassis ID: %w", err)
	}
	chassisID := p.formatSNMPValue(chassisIDResult.Variables[0].Value)

	// 获取邻居 Port ID
	portIDOID := baseOID + ".6." + index
	portIDResult, err := snmp.Get([]string{portIDOID})
	if err != nil || len(portIDResult.Variables) == 0 {
		return nil, fmt.Errorf("failed to get port ID: %w", err)
	}
	portID := p.formatSNMPValue(portIDResult.Variables[0].Value)

	// 获取系统名称
	sysNameOID := baseOID + ".9." + index
	sysNameResult, err := snmp.Get([]string{sysNameOID})
	systemName := ""
	if err == nil && len(sysNameResult.Variables) > 0 {
		systemName = p.formatSNMPValue(sysNameResult.Variables[0].Value)
	}

	// 获取系统描述
	sysDescOID := baseOID + ".10." + index
	sysDescResult, err := snmp.Get([]string{sysDescOID})
	systemDesc := ""
	if err == nil && len(sysDescResult.Variables) > 0 {
		systemDesc = p.formatSNMPValue(sysDescResult.Variables[0].Value)
	}

	// 获取端口描述
	portDescOID := baseOID + ".8." + index
	portDescResult, err := snmp.Get([]string{portDescOID})
	portDesc := ""
	if err == nil && len(portDescResult.Variables) > 0 {
		portDesc = p.formatSNMPValue(portDescResult.Variables[0].Value)
	}

	// 获取管理地址（可选）
	mgmtAddrOID := "1.0.8802.1.1.2.1.4.2.1.5." + index
	mgmtAddrResult, err := snmp.Get([]string{mgmtAddrOID})
	mgmtAddr := ""
	if err == nil && len(mgmtAddrResult.Variables) > 0 {
		mgmtAddr = p.formatSNMPValue(mgmtAddrResult.Variables[0].Value)
	}

	return &LLDPNeighbor{
		LocalInterface:     localInterface,
		NeighborChassisID:  chassisID,
		NeighborPortID:     portID,
		NeighborSystemName: systemName,
		NeighborSystemDesc: systemDesc,
		NeighborPortDesc:   portDesc,
		NeighborMgmtAddr:   mgmtAddr,
		TTL:                120, // 默认 TTL
	}, nil
}

// formatSNMPValue 格式化 SNMP 值
func (p *LLDPPlugin) formatSNMPValue(value interface{}) string {
	switch v := value.(type) {
	case []byte:
		// MAC 地址格式（6字节）
		if len(v) == 6 {
			parts := make([]string, 6)
			for i, b := range v {
				parts[i] = fmt.Sprintf("%02x", b)
			}
			return strings.Join(parts, ":")
		}
		return string(v)
	case string:
		return v
	case int:
		return strconv.Itoa(v)
	case uint:
		return strconv.FormatUint(uint64(v), 10)
	default:
		return fmt.Sprintf("%v", v)
	}
}

// getSNMPv3MsgFlags 获取 SNMP v3 消息标志
func (p *LLDPPlugin) getSNMPv3MsgFlags(securityLevel string) gosnmp.SnmpV3MsgFlags {
	switch securityLevel {
	case "noAuthNoPriv":
		return gosnmp.NoAuthNoPriv
	case "authNoPriv":
		return gosnmp.AuthNoPriv
	case "authPriv":
		return gosnmp.AuthPriv
	default:
		return gosnmp.NoAuthNoPriv
	}
}

// collectViaSSH 通过 SSH 采集 LLDP 邻居
func (p *LLDPPlugin) collectViaSSH(ctx context.Context, task *plugin.CollectionTask) ([]LLDPNeighbor, error) {
	// 获取配置
	host := p.getString(task.DeviceConfig, "host", "")
	if host == "" {
		return nil, fmt.Errorf("host is required")
	}

	port := p.getInt(task.DeviceConfig, "port", 22)
	timeout := p.getInt(task.DeviceConfig, "timeout", 30)

	// 从 connection_config 中获取认证信息
	var username, password, privateKey, passphrase string
	var authMethod string

	// 尝试从新的 auth 结构读取
	if authRaw, ok := task.DeviceConfig["auth"]; ok {
		if authMap, ok := authRaw.(map[string]interface{}); ok {
			if configRaw, ok := authMap["config"]; ok {
				if configMap, ok := configRaw.(map[string]interface{}); ok {
					username = p.getString(configMap, "username", "")
					password = p.getString(configMap, "password", "")
					privateKey = p.getString(configMap, "private_key", "")
					passphrase = p.getString(configMap, "passphrase", "")
					authMethod = p.getString(configMap, "auth_method", "password")
				}
			}
		}
	}

	// 兼容旧格式：从 device_config 直接读取
	if username == "" {
		username = p.getString(task.DeviceConfig, "ssh_username", "")
		password = p.getString(task.DeviceConfig, "ssh_password", "")
	}

	if username == "" {
		return nil, fmt.Errorf("SSH username is required")
	}

	// 创建 SSH 客户端配置
	config := &ssh.ClientConfig{
		User:            username,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // 生产环境应验证主机密钥
		Timeout:         time.Duration(timeout) * time.Second,
	}

	// 设置认证方法
	if authMethod == "key" || (authMethod == "both" && privateKey != "") {
		// 密钥认证
		signer, err := p.parsePrivateKey(privateKey, passphrase)
		if err != nil {
			return nil, fmt.Errorf("failed to parse private key: %w", err)
		}
		config.Auth = []ssh.AuthMethod{ssh.PublicKeys(signer)}
	} else if authMethod == "password" || (authMethod == "both" && password != "") {
		// 密码认证
		config.Auth = []ssh.AuthMethod{ssh.Password(password)}
	} else {
		return nil, fmt.Errorf("no valid authentication method provided")
	}

	// 连接
	address := fmt.Sprintf("%s:%d", host, port)
	client, err := ssh.Dial("tcp", address, config)
	if err != nil {
		return nil, fmt.Errorf("failed to connect via SSH: %w", err)
	}
	defer client.Close()

	// 创建会话
	session, err := client.NewSession()
	if err != nil {
		return nil, fmt.Errorf("failed to create SSH session: %w", err)
	}
	defer session.Close()

	// 检测设备类型并执行相应的命令
	deviceType := p.getString(task.DeviceConfig, "device_type", "")
	command := p.getLLDPCommand(deviceType)

	// 执行命令
	var stdout, stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr

	if err := session.Run(command); err != nil {
		return nil, fmt.Errorf("failed to execute command '%s': %w, stderr: %s", command, err, stderr.String())
	}

	// 解析输出
	output := stdout.String()
	neighbors := p.parseLLDPOutput(output, deviceType)

	return neighbors, nil
}

// parsePrivateKey 解析私钥
func (p *LLDPPlugin) parsePrivateKey(keyData, passphrase string) (ssh.Signer, error) {
	var signer ssh.Signer
	var err error

	if passphrase != "" {
		signer, err = ssh.ParsePrivateKeyWithPassphrase([]byte(keyData), []byte(passphrase))
	} else {
		signer, err = ssh.ParsePrivateKey([]byte(keyData))
	}

	if err != nil {
		return nil, err
	}

	return signer, nil
}

// getLLDPCommand 根据设备类型获取 LLDP 命令
func (p *LLDPPlugin) getLLDPCommand(deviceType string) string {
	// 根据设备类型返回相应的命令
	switch deviceType {
	case "cisco", "ios", "ios-xe":
		return "show lldp neighbors detail"
	case "juniper", "junos":
		return "show lldp neighbors"
	case "huawei", "vrp":
		return "display lldp neighbor-information"
	case "h3c", "comware":
		return "display lldp neighbor-information"
	case "arista", "eos":
		return "show lldp neighbors detail"
	default:
		// 默认尝试 Cisco 命令
		return "show lldp neighbors detail"
	}
}

// testSNMPConnection 测试 SNMP 连接
func (p *LLDPPlugin) testSNMPConnection(deviceConfig map[string]interface{}) error {
	host := p.getString(deviceConfig, "host", "")
	if host == "" {
		return fmt.Errorf("host is required")
	}

	port := p.getInt(deviceConfig, "port", 161)
	snmpVersion := p.getString(deviceConfig, "snmp_version", "2c")

	// 创建 SNMP 客户端（简化版，只测试连接）
	snmp := &gosnmp.GoSNMP{
		Target:  host,
		Port:    uint16(port),
		Timeout: 5 * time.Second,
		Retries: 1,
	}

	switch snmpVersion {
	case "1":
		snmp.Version = gosnmp.Version1
		snmp.Community = p.getString(deviceConfig, "snmp_community", "public")
	case "2c":
		snmp.Version = gosnmp.Version2c
		snmp.Community = p.getString(deviceConfig, "snmp_community", "public")
	case "3":
		// SNMP v3 测试需要完整配置，这里简化处理
		return fmt.Errorf("SNMP v3 connection test requires full authentication configuration")
	default:
		return fmt.Errorf("unsupported SNMP version: %s", snmpVersion)
	}

	if err := snmp.Connect(); err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer snmp.Conn.Close()

	// 尝试获取系统描述（标准 OID）
	sysDescOID := "1.3.6.1.2.1.1.1.0"
	_, err := snmp.Get([]string{sysDescOID})
	if err != nil {
		return fmt.Errorf("failed to query device: %w", err)
	}

	return nil
}

// testSSHConnection 测试 SSH 连接
func (p *LLDPPlugin) testSSHConnection(deviceConfig map[string]interface{}) error {
	host := p.getString(deviceConfig, "host", "")
	if host == "" {
		return fmt.Errorf("host is required")
	}

	port := p.getInt(deviceConfig, "port", 22)
	timeout := p.getInt(deviceConfig, "timeout", 10)

	username := p.getString(deviceConfig, "ssh_username", "")
	password := p.getString(deviceConfig, "ssh_password", "")

	if username == "" {
		return fmt.Errorf("SSH username is required")
	}

	// 创建 SSH 客户端配置
	config := &ssh.ClientConfig{
		User:            username,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         time.Duration(timeout) * time.Second,
	}

	if password != "" {
		config.Auth = []ssh.AuthMethod{ssh.Password(password)}
	} else {
		return fmt.Errorf("SSH password or private key is required")
	}

	// 尝试连接
	address := fmt.Sprintf("%s:%d", host, port)
	client, err := ssh.Dial("tcp", address, config)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer client.Close()

	return nil
}

// getString 获取字符串配置
func (p *LLDPPlugin) getString(config map[string]interface{}, key string, defaultValue string) string {
	if val, ok := config[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return defaultValue
}

// getInt 获取整数配置
func (p *LLDPPlugin) getInt(config map[string]interface{}, key string, defaultValue int) int {
	if val, ok := config[key]; ok {
		switch v := val.(type) {
		case int:
			return v
		case float64:
			return int(v)
		case string:
			if i, err := strconv.Atoi(v); err == nil {
				return i
			}
		}
	}
	return defaultValue
}

// parseLLDPOutput 解析 LLDP 输出（通用方法，供 SNMP 和 SSH 使用）
func (p *LLDPPlugin) parseLLDPOutput(output, deviceType string) []LLDPNeighbor {
	neighbors := make([]LLDPNeighbor, 0)

	switch deviceType {
	case "cisco", "ios", "ios-xe", "arista", "eos":
		neighbors = p.parseCiscoLLDPOutput(output)
	case "juniper", "junos":
		neighbors = p.parseJuniperLLDPOutput(output)
	case "huawei", "vrp", "h3c", "comware":
		neighbors = p.parseHuaweiLLDPOutput(output)
	default:
		// 尝试通用解析
		neighbors = p.parseGenericLLDPOutput(output)
	}

	return neighbors
}

// parseCiscoLLDPOutput 解析 Cisco 设备输出
func (p *LLDPPlugin) parseCiscoLLDPOutput(output string) []LLDPNeighbor {
	neighbors := make([]LLDPNeighbor, 0)

	// Cisco 输出格式示例：
	// Capability codes:
	//     (R) Router, (B) Bridge, (T) Telephone, (C) DOCSIS Cable Device
	//     (W) WLAN Access Point, (P) Repeater, (S) Station, (O) Other
	//
	// Device ID           Local Intf     Hold-time  Capability      Port ID
	// Switch-02           Gi0/1         120        B               Gi0/2
	// Router-01           Gi0/3         180        R               Fa0/1

	lines := strings.Split(output, "\n")
	inTable := false

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// 检测表头
		if strings.Contains(line, "Device ID") && strings.Contains(line, "Local Intf") {
			inTable = true
			continue
		}

		if !inTable {
			continue
		}

		// 解析表行
		fields := strings.Fields(line)
		if len(fields) >= 5 {
			neighbor := LLDPNeighbor{
				NeighborSystemName: fields[0],
				LocalInterface:     fields[1],
				NeighborPortID:     fields[4],
				TTL:                120,
			}

			// 尝试解析 TTL
			if len(fields) >= 3 {
				if ttl, err := strconv.Atoi(fields[2]); err == nil {
					neighbor.TTL = ttl
				}
			}

			neighbors = append(neighbors, neighbor)
		}
	}

	return neighbors
}

// parseJuniperLLDPOutput 解析 Juniper 设备输出
func (p *LLDPPlugin) parseJuniperLLDPOutput(output string) []LLDPNeighbor {
	neighbors := make([]LLDPNeighbor, 0)

	// Juniper 输出格式示例：
	// Local Interface    Parent Interface    Chassis Id          Port info          System Name
	// ge-0/0/0.0         -                  00:11:22:33:44:55   ge-0/0/1.0         Switch-02

	lines := strings.Split(output, "\n")
	inTable := false

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if strings.Contains(line, "Local Interface") && strings.Contains(line, "Chassis Id") {
			inTable = true
			continue
		}

		if !inTable {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) >= 4 {
			neighbor := LLDPNeighbor{
				LocalInterface:    fields[0],
				NeighborChassisID: fields[2],
				NeighborPortID:    fields[3],
				TTL:               120,
			}

			if len(fields) >= 5 {
				neighbor.NeighborSystemName = fields[4]
			}

			neighbors = append(neighbors, neighbor)
		}
	}

	return neighbors
}

// parseHuaweiLLDPOutput 解析 Huawei/H3C 设备输出
func (p *LLDPPlugin) parseHuaweiLLDPOutput(output string) []LLDPNeighbor {
	neighbors := make([]LLDPNeighbor, 0)

	// Huawei 输出格式示例：
	// Local Interface: GigabitEthernet0/0/1
	// Chassis ID: 00-11-22-33-44-55
	// Port ID: GigabitEthernet0/0/2
	// System Name: Switch-02

	// 使用正则表达式解析
	re := regexp.MustCompile(`Local Interface:\s*(.+?)\s*\n.*?Chassis ID:\s*(.+?)\s*\n.*?Port ID:\s*(.+?)\s*\n(?:.*?System Name:\s*(.+?)\s*\n)?`)

	matches := re.FindAllStringSubmatch(output, -1)
	for _, match := range matches {
		if len(match) >= 4 {
			neighbor := LLDPNeighbor{
				LocalInterface:    strings.TrimSpace(match[1]),
				NeighborChassisID: strings.TrimSpace(match[2]),
				NeighborPortID:    strings.TrimSpace(match[3]),
				TTL:               120,
			}

			if len(match) >= 5 && match[4] != "" {
				neighbor.NeighborSystemName = strings.TrimSpace(match[4])
			}

			neighbors = append(neighbors, neighbor)
		}
	}

	return neighbors
}

// parseGenericLLDPOutput 通用解析（尝试多种格式）
func (p *LLDPPlugin) parseGenericLLDPOutput(output string) []LLDPNeighbor {
	// 先尝试 Cisco 格式
	neighbors := p.parseCiscoLLDPOutput(output)
	if len(neighbors) > 0 {
		return neighbors
	}

	// 再尝试 Juniper 格式
	neighbors = p.parseJuniperLLDPOutput(output)
	if len(neighbors) > 0 {
		return neighbors
	}

	// 最后尝试 Huawei 格式
	return p.parseHuaweiLLDPOutput(output)
}

// 确保实现了接口
var _ plugin.Plugin = (*LLDPPlugin)(nil)
