package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/celestial/gravital-core/internal/model"
	"github.com/celestial/gravital-core/internal/pkg/auth"
)

// UserHandler 用户处理器
type UserHandler struct {
	db         *gorm.DB
	bcryptCost int
}

// NewUserHandler 创建用户处理器
func NewUserHandler(db *gorm.DB, bcryptCost int) *UserHandler {
	return &UserHandler{
		db:         db,
		bcryptCost: bcryptCost,
	}
}

// ListUsers 获取用户列表
func (h *UserHandler) ListUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	keyword := c.Query("keyword")
	roleIDStr := c.Query("role_id")
	enabledStr := c.Query("enabled")

	offset := (page - 1) * size

	query := h.db.Model(&model.User{}).Preload("Role")

	if keyword != "" {
		query = query.Where("username LIKE ? OR email LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	if roleIDStr != "" {
		roleID, _ := strconv.Atoi(roleIDStr)
		query = query.Where("role_id = ?", roleID)
	}

	if enabledStr != "" {
		enabled := enabledStr == "true"
		query = query.Where("enabled = ?", enabled)
	}

	var total int64
	query.Count(&total)

	var users []model.User
	if err := query.Offset(offset).Limit(size).Find(&users).Error; err != nil {
		ErrorResponse(c, http.StatusInternalServerError, 10001, "获取用户列表失败")
		return
	}

	SuccessResponse(c, gin.H{
		"items": users,
		"total": total,
		"page":  page,
		"size":  size,
	})
}

// GetUser 获取用户详情
func (h *UserHandler) GetUser(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)

	var user model.User
	if err := h.db.Preload("Role").First(&user, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			ErrorResponse(c, http.StatusNotFound, 40004, "用户不存在")
			return
		}
		ErrorResponse(c, http.StatusInternalServerError, 10001, "获取用户失败")
		return
	}

	SuccessResponse(c, user)
}

// CreateUser 创建用户
func (h *UserHandler) CreateUser(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=6"`
		RoleID   uint   `json:"role_id" binding:"required"`
		Enabled  bool   `json:"enabled"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorResponse(c, http.StatusBadRequest, 40001, "参数错误: "+err.Error())
		return
	}

	// 检查用户名是否已存在
	var count int64
	h.db.Model(&model.User{}).Where("username = ?", req.Username).Count(&count)
	if count > 0 {
		ErrorResponse(c, http.StatusBadRequest, 40002, "用户名已存在")
		return
	}

	// 检查邮箱是否已存在
	h.db.Model(&model.User{}).Where("email = ?", req.Email).Count(&count)
	if count > 0 {
		ErrorResponse(c, http.StatusBadRequest, 40002, "邮箱已存在")
		return
	}

	// 哈希密码
	passwordHash, err := auth.HashPassword(req.Password, h.bcryptCost)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, 10001, "密码加密失败")
		return
	}

	user := model.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: passwordHash,
		RoleID:       req.RoleID,
		Enabled:      req.Enabled,
	}

	if err := h.db.Create(&user).Error; err != nil {
		ErrorResponse(c, http.StatusInternalServerError, 10001, "创建用户失败")
		return
	}

	// 重新加载用户信息（包含 Role）
	h.db.Preload("Role").First(&user, user.ID)

	SuccessResponse(c, user)
}

// UpdateUser 更新用户
func (h *UserHandler) UpdateUser(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)

	var user model.User
	if err := h.db.First(&user, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			ErrorResponse(c, http.StatusNotFound, 40004, "用户不存在")
			return
		}
		ErrorResponse(c, http.StatusInternalServerError, 10001, "获取用户失败")
		return
	}

	var req struct {
		Username *string `json:"username"`
		Email    *string `json:"email"`
		RoleID   *uint   `json:"role_id"`
		Enabled  *bool   `json:"enabled"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorResponse(c, http.StatusBadRequest, 40001, "参数错误: "+err.Error())
		return
	}

	// 更新字段
	if req.Username != nil {
		user.Username = *req.Username
	}
	if req.Email != nil {
		user.Email = *req.Email
	}
	if req.RoleID != nil {
		user.RoleID = *req.RoleID
	}
	if req.Enabled != nil {
		user.Enabled = *req.Enabled
	}

	if err := h.db.Save(&user).Error; err != nil {
		ErrorResponse(c, http.StatusInternalServerError, 10001, "更新用户失败")
		return
	}

	// 重新加载用户信息（包含 Role）
	h.db.Preload("Role").First(&user, user.ID)

	SuccessResponse(c, user)
}

// DeleteUser 删除用户
func (h *UserHandler) DeleteUser(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)

	// 不能删除 ID 为 1 的管理员
	if id == 1 {
		ErrorResponse(c, http.StatusForbidden, 40003, "不能删除默认管理员")
		return
	}

	if err := h.db.Delete(&model.User{}, id).Error; err != nil {
		ErrorResponse(c, http.StatusInternalServerError, 10001, "删除用户失败")
		return
	}

	SuccessResponse(c, gin.H{"message": "删除成功"})
}

// ToggleUser 启用/禁用用户
func (h *UserHandler) ToggleUser(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)

	// 不能禁用 ID 为 1 的管理员
	if id == 1 {
		ErrorResponse(c, http.StatusForbidden, 40003, "不能禁用默认管理员")
		return
	}

	var req struct {
		Enabled bool `json:"enabled"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorResponse(c, http.StatusBadRequest, 40001, "参数错误: "+err.Error())
		return
	}

	if err := h.db.Model(&model.User{}).Where("id = ?", id).Update("enabled", req.Enabled).Error; err != nil {
		ErrorResponse(c, http.StatusInternalServerError, 10001, "更新用户状态失败")
		return
	}

	SuccessResponse(c, gin.H{"message": "操作成功"})
}

// ResetPassword 重置密码
func (h *UserHandler) ResetPassword(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)

	var req struct {
		Password string `json:"password" binding:"required,min=6"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorResponse(c, http.StatusBadRequest, 40001, "参数错误: "+err.Error())
		return
	}

	// 哈希密码
	passwordHash, err := auth.HashPassword(req.Password, h.bcryptCost)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, 10001, "密码加密失败")
		return
	}

	if err := h.db.Model(&model.User{}).Where("id = ?", id).Update("password_hash", passwordHash).Error; err != nil {
		ErrorResponse(c, http.StatusInternalServerError, 10001, "重置密码失败")
		return
	}

	SuccessResponse(c, gin.H{"message": "密码重置成功"})
}

// ListRoles 获取角色列表
func (h *UserHandler) ListRoles(c *gin.Context) {
	var roles []model.Role
	if err := h.db.Find(&roles).Error; err != nil {
		ErrorResponse(c, http.StatusInternalServerError, 10001, "获取角色列表失败")
		return
	}

	SuccessResponse(c, gin.H{
		"items": roles,
		"total": len(roles),
	})
}

// GetRole 获取角色详情
func (h *UserHandler) GetRole(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)

	var role model.Role
	if err := h.db.First(&role, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			ErrorResponse(c, http.StatusNotFound, 40004, "角色不存在")
			return
		}
		ErrorResponse(c, http.StatusInternalServerError, 10001, "获取角色失败")
		return
	}

	SuccessResponse(c, role)
}

// CreateRole 创建角色
func (h *UserHandler) CreateRole(c *gin.Context) {
	var req struct {
		Name        string   `json:"name" binding:"required"`
		Permissions []string `json:"permissions" binding:"required"`
		Description string   `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorResponse(c, http.StatusBadRequest, 40001, "参数错误: "+err.Error())
		return
	}

	// 检查角色名是否已存在
	var count int64
	h.db.Model(&model.Role{}).Where("name = ?", req.Name).Count(&count)
	if count > 0 {
		ErrorResponse(c, http.StatusBadRequest, 40002, "角色名已存在")
		return
	}

	role := model.Role{
		Name:        req.Name,
		Permissions: req.Permissions,
		Description: req.Description,
	}

	if err := h.db.Create(&role).Error; err != nil {
		ErrorResponse(c, http.StatusInternalServerError, 10001, "创建角色失败")
		return
	}

	SuccessResponse(c, role)
}

// UpdateRole 更新角色
func (h *UserHandler) UpdateRole(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)

	// 不能修改 ID 为 1 的管理员角色
	if id == 1 {
		ErrorResponse(c, http.StatusForbidden, 40003, "不能修改默认管理员角色")
		return
	}

	var role model.Role
	if err := h.db.First(&role, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			ErrorResponse(c, http.StatusNotFound, 40004, "角色不存在")
			return
		}
		ErrorResponse(c, http.StatusInternalServerError, 10001, "获取角色失败")
		return
	}

	var req struct {
		Name        *string   `json:"name"`
		Permissions *[]string `json:"permissions"`
		Description *string   `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorResponse(c, http.StatusBadRequest, 40001, "参数错误: "+err.Error())
		return
	}

	// 更新字段
	if req.Name != nil {
		role.Name = *req.Name
	}
	if req.Permissions != nil {
		role.Permissions = *req.Permissions
	}
	if req.Description != nil {
		role.Description = *req.Description
	}

	if err := h.db.Save(&role).Error; err != nil {
		ErrorResponse(c, http.StatusInternalServerError, 10001, "更新角色失败")
		return
	}

	SuccessResponse(c, role)
}

// DeleteRole 删除角色
func (h *UserHandler) DeleteRole(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)

	// 不能删除 ID 为 1 的管理员角色
	if id == 1 {
		ErrorResponse(c, http.StatusForbidden, 40003, "不能删除默认管理员角色")
		return
	}

	// 检查是否有用户使用该角色
	var count int64
	h.db.Model(&model.User{}).Where("role_id = ?", id).Count(&count)
	if count > 0 {
		ErrorResponse(c, http.StatusBadRequest, 40002, "该角色正在被使用，无法删除")
		return
	}

	if err := h.db.Delete(&model.Role{}, id).Error; err != nil {
		ErrorResponse(c, http.StatusInternalServerError, 10001, "删除角色失败")
		return
	}

	SuccessResponse(c, gin.H{"message": "删除成功"})
}

