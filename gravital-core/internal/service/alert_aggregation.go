package service

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/celestial/gravital-core/internal/model"
)

// AlertAggregation 告警聚合信息
type AlertAggregation struct {
	RuleID      uint   `json:"rule_id"`
	RuleName    string `json:"rule_name"`
	Severity    string `json:"severity"`
	Description string `json:"description"`
	TotalCount  int    `json:"total_count"`   // 总告警数
	FiringCount int    `json:"firing_count"`  // 告警中数量
	AckedCount  int    `json:"acked_count"`   // 已确认数量
	FirstFired  string `json:"first_fired"`   // 最早触发时间
	LastFired   string `json:"last_fired"`    // 最近触发时间
	Devices     []struct {
		DeviceID    string `json:"device_id"`
		DeviceName  string `json:"device_name"`
		Status      string `json:"status"`
		TriggeredAt string `json:"triggered_at"`
	} `json:"devices"` // 受影响的设备列表
}

// GetAlertAggregations 获取告警聚合信息
func GetAlertAggregations(ctx context.Context, db *gorm.DB) ([]*AlertAggregation, error) {
	// 查询所有活跃的告警事件（firing 和 acknowledged）
	var events []*model.AlertEvent
	err := db.WithContext(ctx).
		Preload("Rule").
		Where("status IN ?", []string{"firing", "acknowledged"}).
		Order("triggered_at DESC").
		Find(&events).Error
	
	if err != nil {
		return nil, err
	}

	// 按规则聚合
	aggregationMap := make(map[uint]*AlertAggregation)
	
	for _, event := range events {
		if event.Rule == nil {
			continue
		}

		agg, exists := aggregationMap[event.RuleID]
		if !exists {
			agg = &AlertAggregation{
				RuleID:      event.RuleID,
				RuleName:    event.Rule.RuleName,
				Severity:    event.Rule.Severity,
				Description: event.Rule.Description,
				TotalCount:  0,
				FiringCount: 0,
				AckedCount:  0,
				FirstFired:  event.TriggeredAt.Format("2006-01-02 15:04:05"),
				LastFired:   event.TriggeredAt.Format("2006-01-02 15:04:05"),
				Devices:     make([]struct {
					DeviceID    string `json:"device_id"`
					DeviceName  string `json:"device_name"`
					Status      string `json:"status"`
					TriggeredAt string `json:"triggered_at"`
				}, 0),
			}
			aggregationMap[event.RuleID] = agg
		}

		// 更新统计
		agg.TotalCount++
		if event.Status == "firing" {
			agg.FiringCount++
		} else if event.Acknowledged {
			agg.AckedCount++
		}

		// 更新时间范围
		triggeredAt := event.TriggeredAt.Format("2006-01-02 15:04:05")
		if triggeredAt < agg.FirstFired {
			agg.FirstFired = triggeredAt
		}
		if triggeredAt > agg.LastFired {
			agg.LastFired = triggeredAt
		}

		// 添加设备信息
		agg.Devices = append(agg.Devices, struct {
			DeviceID    string `json:"device_id"`
			DeviceName  string `json:"device_name"`
			Status      string `json:"status"`
			TriggeredAt string `json:"triggered_at"`
		}{
			DeviceID:    event.DeviceID,
			DeviceName:  event.DeviceID, // TODO: 从设备表查询设备名称
			Status:      event.Status,
			TriggeredAt: triggeredAt,
		})
	}

	// 转换为切片
	result := make([]*AlertAggregation, 0, len(aggregationMap))
	for _, agg := range aggregationMap {
		result = append(result, agg)
	}

	return result, nil
}

// BatchAcknowledgeEvents 批量确认告警事件
func BatchAcknowledgeEvents(ctx context.Context, db *gorm.DB, ids []uint, userID uint, comment string) error {
	now := time.Now()
	return db.WithContext(ctx).
		Model(&model.AlertEvent{}).
		Where("id IN ?", ids).
		Updates(map[string]interface{}{
			"acknowledged":    true,
			"acknowledged_by": userID,
			"acknowledged_at": now,
			"comment":         comment,
		}).Error
}

// BatchResolveEvents 批量解决告警事件
func BatchResolveEvents(ctx context.Context, db *gorm.DB, ids []uint, comment string) error {
	now := time.Now()
	return db.WithContext(ctx).
		Model(&model.AlertEvent{}).
		Where("id IN ?", ids).
		Updates(map[string]interface{}{
			"status":      "resolved",
			"resolved_at": now,
			"comment":     comment,
		}).Error
}

// ResolveEventsByRule 解决某个规则的所有告警
func ResolveEventsByRule(ctx context.Context, db *gorm.DB, ruleID uint, comment string) error {
	now := time.Now()
	return db.WithContext(ctx).
		Model(&model.AlertEvent{}).
		Where("rule_id = ? AND status IN ?", ruleID, []string{"firing", "acknowledged"}).
		Updates(map[string]interface{}{
			"status":      "resolved",
			"resolved_at": now,
			"comment":     comment,
		}).Error
}

