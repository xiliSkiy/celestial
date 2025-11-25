package handler

import (
	"net/http"
	"strconv"

	"github.com/celestial/gravital-core/internal/service"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// TopologyHandler 拓扑处理器
type TopologyHandler struct {
	topologyService service.TopologyService
	logger          *zap.Logger
}

// NewTopologyHandler 创建拓扑处理器实例
func NewTopologyHandler(topologyService service.TopologyService, logger *zap.Logger) *TopologyHandler {
	return &TopologyHandler{
		topologyService: topologyService,
		logger:          logger,
	}
}

// CreateTopology 创建拓扑
func (h *TopologyHandler) CreateTopology(c *gin.Context) {
	var req service.CreateTopologyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    10001,
			"message": "Invalid request: " + err.Error(),
		})
		return
	}

	topology, err := h.topologyService.CreateTopology(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to create topology", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    10001,
			"message": "Failed to create topology: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": topology,
	})
}

// GetTopology 获取拓扑详情
func (h *TopologyHandler) GetTopology(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    10001,
			"message": "Invalid topology ID",
		})
		return
	}

	topology, err := h.topologyService.GetTopology(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    10001,
			"message": "Topology not found: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": topology,
	})
}

// UpdateTopology 更新拓扑
func (h *TopologyHandler) UpdateTopology(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    10001,
			"message": "Invalid topology ID",
		})
		return
	}

	var req service.UpdateTopologyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    10001,
			"message": "Invalid request: " + err.Error(),
		})
		return
	}

	if err := h.topologyService.UpdateTopology(c.Request.Context(), uint(id), &req); err != nil {
		h.logger.Error("Failed to update topology", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    10001,
			"message": "Failed to update topology: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": "success",
	})
}

// DeleteTopology 删除拓扑
func (h *TopologyHandler) DeleteTopology(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    10001,
			"message": "Invalid topology ID",
		})
		return
	}

	if err := h.topologyService.DeleteTopology(c.Request.Context(), uint(id)); err != nil {
		h.logger.Error("Failed to delete topology", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    10001,
			"message": "Failed to delete topology: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": "success",
	})
}

// ListTopologies 获取拓扑列表
func (h *TopologyHandler) ListTopologies(c *gin.Context) {
	var req service.ListTopologyRequest

	// 解析查询参数
	if page := c.Query("page"); page != "" {
		if p, err := strconv.Atoi(page); err == nil {
			req.Page = p
		}
	}
	if pageSize := c.Query("page_size"); pageSize != "" {
		if ps, err := strconv.Atoi(pageSize); err == nil {
			req.PageSize = ps
		}
	}
	req.Type = c.Query("type")
	req.Scope = c.Query("scope")
	req.Keyword = c.Query("keyword")

	resp, err := h.topologyService.ListTopologies(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to list topologies", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    10001,
			"message": "Failed to list topologies: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": resp,
	})
}

// AddNode 添加节点
func (h *TopologyHandler) AddNode(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    10001,
			"message": "Invalid topology ID",
		})
		return
	}

	var req service.AddNodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    10001,
			"message": "Invalid request: " + err.Error(),
		})
		return
	}

	node, err := h.topologyService.AddNode(c.Request.Context(), uint(id), &req)
	if err != nil {
		h.logger.Error("Failed to add node", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    10001,
			"message": "Failed to add node: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": node,
	})
}

// UpdateNodePosition 更新节点位置
func (h *TopologyHandler) UpdateNodePosition(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    10001,
			"message": "Invalid topology ID",
		})
		return
	}

	nodeID, err := strconv.ParseUint(c.Param("node_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    10001,
			"message": "Invalid node ID",
		})
		return
	}

	var req struct {
		X float64 `json:"x"`
		Y float64 `json:"y"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    10001,
			"message": "Invalid request: " + err.Error(),
		})
		return
	}

	if err := h.topologyService.UpdateNodePosition(c.Request.Context(), uint(id), uint(nodeID), req.X, req.Y); err != nil {
		h.logger.Error("Failed to update node position", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    10001,
			"message": "Failed to update node position: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": "success",
	})
}

// BatchUpdateNodes 批量更新节点
func (h *TopologyHandler) BatchUpdateNodes(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    10001,
			"message": "Invalid topology ID",
		})
		return
	}

	var req service.BatchUpdateNodesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    10001,
			"message": "Invalid request: " + err.Error(),
		})
		return
	}

	if err := h.topologyService.BatchUpdateNodes(c.Request.Context(), uint(id), &req); err != nil {
		h.logger.Error("Failed to batch update nodes", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    10001,
			"message": "Failed to batch update nodes: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": "success",
	})
}

// DeleteNode 删除节点
func (h *TopologyHandler) DeleteNode(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    10001,
			"message": "Invalid topology ID",
		})
		return
	}

	nodeID, err := strconv.ParseUint(c.Param("node_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    10001,
			"message": "Invalid node ID",
		})
		return
	}

	if err := h.topologyService.DeleteNode(c.Request.Context(), uint(id), uint(nodeID)); err != nil {
		h.logger.Error("Failed to delete node", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    10001,
			"message": "Failed to delete node: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": "success",
	})
}

// AddLink 添加链路
func (h *TopologyHandler) AddLink(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    10001,
			"message": "Invalid topology ID",
		})
		return
	}

	var req service.AddLinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    10001,
			"message": "Invalid request: " + err.Error(),
		})
		return
	}

	link, err := h.topologyService.AddLink(c.Request.Context(), uint(id), &req)
	if err != nil {
		h.logger.Error("Failed to add link", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    10001,
			"message": "Failed to add link: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": link,
	})
}

// UpdateLinkStatus 更新链路状态
func (h *TopologyHandler) UpdateLinkStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    10001,
			"message": "Invalid topology ID",
		})
		return
	}

	linkID, err := strconv.ParseUint(c.Param("link_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    10001,
			"message": "Invalid link ID",
		})
		return
	}

	var req service.UpdateLinkStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    10001,
			"message": "Invalid request: " + err.Error(),
		})
		return
	}

	if err := h.topologyService.UpdateLinkStatus(c.Request.Context(), uint(id), uint(linkID), &req); err != nil {
		h.logger.Error("Failed to update link status", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    10001,
			"message": "Failed to update link status: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": "success",
	})
}

// DeleteLink 删除链路
func (h *TopologyHandler) DeleteLink(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    10001,
			"message": "Invalid topology ID",
		})
		return
	}

	linkID, err := strconv.ParseUint(c.Param("link_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    10001,
			"message": "Invalid link ID",
		})
		return
	}

	if err := h.topologyService.DeleteLink(c.Request.Context(), uint(id), uint(linkID)); err != nil {
		h.logger.Error("Failed to delete link", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    10001,
			"message": "Failed to delete link: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": "success",
	})
}

// ApplyLayout 应用布局
func (h *TopologyHandler) ApplyLayout(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    10001,
			"message": "Invalid topology ID",
		})
		return
	}

	var req service.ApplyLayoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    10001,
			"message": "Invalid request: " + err.Error(),
		})
		return
	}

	resp, err := h.topologyService.ApplyLayout(c.Request.Context(), uint(id), &req)
	if err != nil {
		h.logger.Error("Failed to apply layout", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    10001,
			"message": "Failed to apply layout: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": resp,
	})
}

// AnalyzePath 路径分析
func (h *TopologyHandler) AnalyzePath(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    10001,
			"message": "Invalid topology ID",
		})
		return
	}

	var req service.PathAnalysisRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    10001,
			"message": "Invalid request: " + err.Error(),
		})
		return
	}

	resp, err := h.topologyService.AnalyzePath(c.Request.Context(), uint(id), &req)
	if err != nil {
		h.logger.Error("Failed to analyze path", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    10001,
			"message": "Failed to analyze path: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": resp,
	})
}

// AnalyzeImpact 影响分析
func (h *TopologyHandler) AnalyzeImpact(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    10001,
			"message": "Invalid topology ID",
		})
		return
	}

	var req service.ImpactAnalysisRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    10001,
			"message": "Invalid request: " + err.Error(),
		})
		return
	}

	resp, err := h.topologyService.AnalyzeImpact(c.Request.Context(), uint(id), &req)
	if err != nil {
		h.logger.Error("Failed to analyze impact", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    10001,
			"message": "Failed to analyze impact: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": resp,
	})
}

// GetVersions 获取版本列表
func (h *TopologyHandler) GetVersions(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    10001,
			"message": "Invalid topology ID",
		})
		return
	}

	versions, err := h.topologyService.GetVersions(c.Request.Context(), uint(id))
	if err != nil {
		h.logger.Error("Failed to get versions", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    10001,
			"message": "Failed to get versions: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": versions,
	})
}

// CreateSnapshot 创建快照
func (h *TopologyHandler) CreateSnapshot(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    10001,
			"message": "Invalid topology ID",
		})
		return
	}

	var req struct {
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    10001,
			"message": "Invalid request: " + err.Error(),
		})
		return
	}

	// TODO: 从上下文获取用户 ID
	userID := uint(1)

	if err := h.topologyService.CreateSnapshot(c.Request.Context(), uint(id), req.Description, userID); err != nil {
		h.logger.Error("Failed to create snapshot", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    10001,
			"message": "Failed to create snapshot: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": "success",
	})
}

// RestoreVersion 恢复版本
func (h *TopologyHandler) RestoreVersion(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    10001,
			"message": "Invalid topology ID",
		})
		return
	}

	version, err := strconv.Atoi(c.Param("version"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    10001,
			"message": "Invalid version",
		})
		return
	}

	if err := h.topologyService.RestoreVersion(c.Request.Context(), uint(id), version); err != nil {
		h.logger.Error("Failed to restore version", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    10001,
			"message": "Failed to restore version: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": "success",
	})
}

