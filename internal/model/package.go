package model

import (
	"time"

	"gorm.io/gorm"
)

// Package 包裹核心信息
type Package struct {
	PackageID        string         `gorm:"primaryKey;size:32;comment:运单号"`
	SenderName       string         `gorm:"size:64;not null;comment:寄件人姓名"`
	SenderPhone      string         `gorm:"size:20;not null;comment:寄件人电话"`
	SenderAddress    string         `gorm:"size:255;not null;comment:寄件人地址"`
	ReceiverName     string         `gorm:"size:64;not null;comment:收件人姓名"`
	ReceiverPhone    string         `gorm:"size:20;not null;comment:收件人电话"`
	ReceiverAddress  string         `gorm:"size:255;not null;comment:收件人地址"`
	ReceiverProvince string         `gorm:"size:32;not null;comment:收件人省份"`
	ReceiverCity     string         `gorm:"size:32;not null;comment:收件人城市"`
	ReceiverDistrict string         `gorm:"size:32;not null;comment:收件人区县"`
	Weight           float64        `gorm:"not null;comment:包裹重量(kg)"`
	Length           float64        `gorm:"comment:长度(cm)"`
	Width            float64        `gorm:"comment:宽度(cm)"`
	Height           float64        `gorm:"comment:高度(cm)"`
	Status           string         `gorm:"size:20;not null;default:pending;comment:包裹状态"`
	AbnormalReason   string         `gorm:"size:255;comment:异常原因"`
	AbnormalHandler  string         `gorm:"size:64;comment:异常处理人"`
	CreatedAt        time.Time      `gorm:"autoCreateTime;comment:创建时间"`
	UpdatedAt        time.Time      `gorm:"autoUpdateTime;comment:更新时间"`
	DeletedAt        gorm.DeletedAt `gorm:"index;comment:删除时间"`
}

// TableName 表名
func (p *Package) TableName() string {
	return "packages"
}

// PackageTrace 包裹轨迹
type PackageTrace struct {
	TraceID       string         `gorm:"primaryKey;size:32;comment:轨迹ID"`
	PackageID     string         `gorm:"size:32;not null;index;comment:运单号"`
	NodeType      string         `gorm:"size:20;not null;comment:节点类型"`
	NodeName      string         `gorm:"size:64;not null;comment:节点名称"`
	NodeAddress   string         `gorm:"size:255;comment:节点地址"`
	Longitude     float64        `gorm:"comment:经度"`
	Latitude      float64        `gorm:"comment:纬度"`
	OperationTime time.Time      `gorm:"not null;comment:操作时间"`
	Operator      string         `gorm:"size:64;comment:操作人"`
	Remark        string         `gorm:"size:255;comment:备注"`
	CreatedAt     time.Time      `gorm:"autoCreateTime;comment:创建时间"`
	DeletedAt     gorm.DeletedAt `gorm:"index;comment:删除时间"`
}

// TableName 表名
func (pt *PackageTrace) TableName() string {
	return "package_traces"
}

// AbnormalRecord 异常记录
type AbnormalRecord struct {
	RecordID         string         `gorm:"primaryKey;size:32;comment:异常记录ID"`
	PackageID        string         `gorm:"size:32;not null;index;comment:运单号"`
	AbnormalType     string         `gorm:"size:20;not null;comment:异常类型"`
	AbnormalReason   string         `gorm:"size:255;not null;comment:异常原因"`
	ProcessingMethod string         `gorm:"size:255;comment:处理方式"`
	Processor        string         `gorm:"size:64;comment:处理人"`
	ProcessingTime   time.Time      `gorm:"comment:处理时间"`
	Status           string         `gorm:"size:20;not null;default:pending;comment:处理状态"`
	CreatedAt        time.Time      `gorm:"autoCreateTime;comment:创建时间"`
	UpdatedAt        time.Time      `gorm:"autoUpdateTime;comment:更新时间"`
	DeletedAt        gorm.DeletedAt `gorm:"index;comment:删除时间"`
}

// TableName 表名
func (ar *AbnormalRecord) TableName() string {
	return "abnormal_records"
}
