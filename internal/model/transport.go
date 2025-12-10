package model

import (
	"time"

	"github.com/LFrankl/fdu-lab3/pkg/errno"
	"gorm.io/gorm"
)

// TransportTask ========== 核心实体 ==========
// TransportTask 运输任务（领域实体：有唯一标识，承载核心业务行为）
type TransportTask struct {
	TaskID           string         `gorm:"primaryKey;size:32;comment:运输任务ID"` // 唯一标识
	StartNode        string         `gorm:"size:64;not null;comment:出发节点（分拣中心/网点）"`
	EndNode          string         `gorm:"size:64;not null;comment:到达节点（分拣中心/网点）"`
	Status           string         `gorm:"size:20;not null;default:pending;comment:任务状态（pending/transporting/arrived/completed/abnormal）"`
	VehicleID        string         `gorm:"size:32;not null;comment:运输车辆ID"`
	DriverID         string         `gorm:"size:32;comment:司机ID"`
	DriverName       string         `gorm:"size:64;comment:司机姓名"`
	PackageCount     int            `gorm:"not null;default:0;comment:绑定包裹数量"`
	EstimatedTime    time.Time      `gorm:"comment:预计到达时间"`
	ActualArriveTime time.Time      `gorm:"default:NULL;comment:实际到达时间"`
	CreatedAt        time.Time      `gorm:"autoCreateTime;comment:创建时间"`
	UpdatedAt        time.Time      `gorm:"autoUpdateTime;comment:更新时间"`
	DeletedAt        gorm.DeletedAt `gorm:"index;comment:软删除时间"`
	// 关联值对象
	Route    TransportRoute    `gorm:"embedded;comment:运输路线"`     // 嵌入式值对象
	Abnormal TransportAbnormal `gorm:"embedded;comment:运输异常信息"` // 嵌入式值对象
}

// TransportTaskPackage 运输任务-包裹关联表（关联实体）
type TransportTaskPackage struct {
	ID              uint           `gorm:"primaryKey;autoIncrement;comment:自增ID"`
	TransportTaskID string         `gorm:"size:32;not null;index;comment:运输任务ID"`
	PackageID       string         `gorm:"size:32;not null;index;comment:包裹运单号"`
	AddedTime       time.Time      `gorm:"not null;comment:包裹绑定时间"`
	DeletedAt       gorm.DeletedAt `gorm:"index;comment:软删除时间"`
}

// TransportRoute ========== 核心值对象 ==========
// TransportRoute 运输路线（值对象：无唯一标识，描述运输任务的路线特征）
type TransportRoute struct {
	RouteJSON string  `gorm:"type:json;comment:路线节点JSON（[{address:xxx,longitude:xxx,latitude:xxx}]）"`
	Distance  float64 `gorm:"comment:运输距离（公里）"`
}

// TransportAbnormal 运输异常信息（值对象：描述运输异常特征）
type TransportAbnormal struct {
	AbnormalType   string    `gorm:"size:20;comment:异常类型（route_change/vehicle_fault/delay）"`
	AbnormalReason string    `gorm:"size:512;comment:异常原因"`
	Handler        string    `gorm:"size:64;comment:处理人"`
	HandleTime     time.Time `gorm:"default:NULL;comment:处理时间"`
	HandleResult   string    `gorm:"size:512;comment:处理结果"`
}

// ChangeStatus 运输任务状态变更（核心业务行为）
func (t *TransportTask) ChangeStatus(newStatus string) error {
	// 状态流转规则：pending → transporting → arrived → completed
	statusFlow := map[string][]string{
		"pending":      {"transporting", "abnormal"},
		"transporting": {"arrived", "abnormal"},
		"arrived":      {"completed", "abnormal"},
		"abnormal":     {"transporting", "completed"},
		"completed":    {}, // 已完成状态不可变更
	}
	// 校验状态流转合法性
	allowedStatus := statusFlow[t.Status]
	allow := false
	for _, s := range allowedStatus {
		if s == newStatus {
			allow = true
			break
		}
	}
	if !allow {
		return errno.ErrTransportStatusInvalid
	}
	// 更新状态
	t.Status = newStatus
	// 若状态为arrived，记录实际到达时间
	if newStatus == "arrived" {
		t.ActualArriveTime = time.Now()
	}
	return nil
}

// BindPackage 绑定包裹到运输任务（核心业务行为）
func (t *TransportTask) BindPackage(packageIDs []string) error {
	// 业务规则：仅pending/transporting状态可绑定包裹
	if t.Status != "pending" && t.Status != "transporting" {
		return errno.ErrTransportTaskNotBindable
	}
	// 更新包裹数量
	t.PackageCount += len(packageIDs)
	t.UpdatedAt = time.Now()
	return nil
}

// ReportAbnormal 上报运输异常（核心业务行为）
func (t *TransportTask) ReportAbnormal(abnormalType, reason string, handler string) {
	t.Status = "abnormal"
	t.Abnormal = TransportAbnormal{
		AbnormalType:   abnormalType,
		AbnormalReason: reason,
		Handler:        handler,
		HandleTime:     time.Now(),
	}
	t.UpdatedAt = time.Now()
}

// HandleAbnormal 处理运输异常（核心业务行为）
func (t *TransportTask) HandleAbnormal(result string, newStatus string) error {
	if t.Status != "abnormal" {
		return errno.ErrTransportTaskNotAbnormal
	}
	// 更新异常处理结果
	t.Abnormal.HandleResult = result
	// 恢复任务状态
	if err := t.ChangeStatus(newStatus); err != nil {
		return err
	}
	t.UpdatedAt = time.Now()
	return nil
}

// TableName 表名映射
func (t *TransportTask) TableName() string {
	return "transport_tasks"
}

func (t *TransportTaskPackage) TableName() string {
	return "transport_task_packages"
}
