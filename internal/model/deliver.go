package model

import (
	"time"

	"github.com/LFrankl/fdu-lab3/pkg/errno"
	"gorm.io/gorm"
)

// DeliveryTask ========== 核心实体 ==========
// DeliveryTask 派送任务（领域实体：有唯一标识，承载核心业务行为）
type DeliveryTask struct {
	TaskID       string         `gorm:"primaryKey;size:32;comment:派送任务ID"` // 唯一标识
	DeliveryArea string         `gorm:"size:128;not null;comment:派送区域（如深圳南山区科技园）"`
	CourierID    string         `gorm:"size:32;not null;comment:派送员ID"`
	CourierName  string         `gorm:"size:64;not null;comment:派送员姓名"`
	Status       string         `gorm:"size:20;not null;default:pending;comment:任务状态（pending/delivering/completed/abnormal）"`
	PackageCount int            `gorm:"not null;default:0;comment:绑定包裹数量"`
	StartTime    time.Time      `gorm:"comment:派送开始时间"`
	CompleteTime time.Time      `gorm:"comment:派送完成时间"`
	StartNode    string         `gorm:"size:64;not null;comment:派送起点（派送网点）"`
	CreatedAt    time.Time      `gorm:"autoCreateTime;comment:创建时间"`
	UpdatedAt    time.Time      `gorm:"autoUpdateTime;comment:更新时间"`
	DeletedAt    gorm.DeletedAt `gorm:"index;comment:软删除时间"`
	// 关联值对象
	Abnormal DeliveryAbnormal `gorm:"embedded;comment:派送异常信息"` // 嵌入式值对象
}

// DeliveryTaskPackage 派送任务-包裹关联表（关联实体）
type DeliveryTaskPackage struct {
	ID             uint           `gorm:"primaryKey;autoIncrement;comment:自增ID"`
	DeliveryTaskID string         `gorm:"size:32;not null;index;comment:派送任务ID"`
	PackageID      string         `gorm:"size:32;not null;index;comment:包裹运单号"`
	DeliveryOrder  int            `gorm:"not null;default:0;comment:派送顺序"`
	SignInfo       SignInfo       `gorm:"embedded;comment:签收信息"` // 嵌入式值对象
	AddedTime      time.Time      `gorm:"not null;comment:包裹绑定时间"`
	DeletedAt      gorm.DeletedAt `gorm:"index;comment:软删除时间"`
}

// DeliveryAbnormal ========== 核心值对象 ==========
// DeliveryAbnormal 派送异常信息（值对象：无唯一标识，描述异常特征）
type DeliveryAbnormal struct {
	AbnormalType   string    `gorm:"size:20;comment:异常类型（receiver_absent/address_error/package_damage）"`
	AbnormalReason string    `gorm:"size:512;comment:异常原因"`
	Handler        string    `gorm:"size:64;comment:处理人"`
	HandleTime     time.Time `gorm:"comment:处理时间"`
	HandleResult   string    `gorm:"size:512;comment:处理结果（如二次派送/退回）"`
}

// SignInfo 签收信息（值对象：描述包裹签收特征）
type SignInfo struct {
	SignerName  string    `gorm:"size:64;comment:签收人姓名"`
	SignerPhone string    `gorm:"size:20;comment:签收人电话（脱敏）"`
	SignTime    time.Time `gorm:"comment:签收时间"`
	SignType    string    `gorm:"size:20;comment:签收类型（person/signboard/agent）"` // 本人/柜机/代签
	SignRemark  string    `gorm:"size:512;comment:签收备注"`
}

// ChangeStatus ========== 领域行为方法 ==========
// ChangeStatus 派送任务状态变更（核心业务行为）
func (d *DeliveryTask) ChangeStatus(newStatus string) error {
	// 状态流转规则：pending → delivering → completed
	statusFlow := map[string][]string{
		"pending":    {"delivering", "abnormal"},
		"delivering": {"completed", "abnormal"},
		"abnormal":   {"delivering", "completed"},
		"completed":  {}, // 已完成状态不可变更
	}
	// 校验状态合法性
	allowedStatus := statusFlow[d.Status]
	allow := false
	for _, s := range allowedStatus {
		if s == newStatus {
			allow = true
			break
		}
	}
	if !allow {
		return errno.ErrDeliveryStatusInvalid
	}
	// 更新状态及时间
	d.Status = newStatus
	switch newStatus {
	case "delivering":
		d.StartTime = time.Now() // 开始派送，记录开始时间
	case "completed":
		d.CompleteTime = time.Now() // 完成派送，记录完成时间
	}
	return nil
}

// BindPackage 绑定包裹到派送任务（核心业务行为）
func (d *DeliveryTask) BindPackage(packageIDs []string) error {
	// 业务规则：仅pending状态可绑定包裹
	if d.Status != "pending" {
		return errno.ErrDeliveryTaskNotBindable
	}
	// 更新包裹数量
	d.PackageCount = len(packageIDs)
	d.UpdatedAt = time.Now()
	return nil
}

// ReportAbnormal 上报派送异常（核心业务行为）
func (d *DeliveryTask) ReportAbnormal(abnormalType, reason string, handler string) {
	d.Status = "abnormal"
	d.Abnormal = DeliveryAbnormal{
		AbnormalType:   abnormalType,
		AbnormalReason: reason,
		Handler:        handler,
		HandleTime:     time.Now(),
	}
	d.UpdatedAt = time.Now()
}

// HandleAbnormal 处理派送异常（核心业务行为）
func (d *DeliveryTask) HandleAbnormal(result string, newStatus string) error {
	if d.Status != "abnormal" {
		return errno.ErrDeliveryTaskNotAbnormal
	}
	// 更新异常处理结果
	d.Abnormal.HandleResult = result
	// 恢复任务状态
	if err := d.ChangeStatus(newStatus); err != nil {
		return err
	}
	d.UpdatedAt = time.Now()
	return nil
}

// SignPackage 包裹签收（核心业务行为）
func (d *DeliveryTaskPackage) SignPackage(signerName, signerPhone, signType, remark string) {
	d.SignInfo = SignInfo{
		SignerName:  signerName,
		SignerPhone: signerPhone, // 实际需脱敏（如保留后4位）
		SignTime:    time.Now(),
		SignType:    signType,
		SignRemark:  remark,
	}
}

// TableName 表名映射
func (d *DeliveryTask) TableName() string {
	return "delivery_tasks"
}

func (d *DeliveryTaskPackage) TableName() string {
	return "delivery_task_packages"
}
