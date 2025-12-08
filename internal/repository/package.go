package repository

import (
	"time"

	"github.com/LFrankl/fdu-lab3/internal/model"
	"github.com/LFrankl/fdu-lab3/internal/util"
	"github.com/LFrankl/fdu-lab3/pkg/db"
	"gorm.io/gorm"
)

// PackageRepository 包裹数据访问接口
type PackageRepository interface {
	Create(pkg *model.Package) error
	GetByID(packageID string) (*model.Package, error)
	UpdateStatus(packageID, status, reason, handler string) error
	CreateTrace(trace *model.PackageTrace) error
	GetTracesByPackageID(packageID string) ([]model.PackageTrace, error)
	CreateAbnormalRecord(record *model.AbnormalRecord) error
}

// packageRepository 实现
type packageRepository struct {
	db    *gorm.DB
	idGen *util.IDGenerator
}

// NewPackageRepository 创建包裹仓库实例
func NewPackageRepository() PackageRepository {
	return &packageRepository{
		db:    db.DB,
		idGen: util.NewIDGenerator(),
	}
}

// Create 创建包裹
func (r *packageRepository) Create(pkg *model.Package) error {
	return r.db.Create(pkg).Error
}

// GetByID 根据运单号获取包裹
func (r *packageRepository) GetByID(packageID string) (*model.Package, error) {
	var pkg model.Package
	if err := r.db.Where("package_id = ?", packageID).First(&pkg).Error; err != nil {
		return nil, err
	}
	return &pkg, nil
}

// UpdateStatus 更新包裹状态
func (r *packageRepository) UpdateStatus(packageID, status, reason, handler string) error {
	updateData := map[string]interface{}{
		"status":     status,
		"updated_at": time.Now(),
	}
	if reason != "" {
		updateData["abnormal_reason"] = reason
	}
	if handler != "" {
		updateData["abnormal_handler"] = handler
	}

	return r.db.Model(&model.Package{}).
		Where("package_id = ?", packageID).
		Updates(updateData).Error
}

// CreateTrace 创建包裹轨迹
func (r *packageRepository) CreateTrace(trace *model.PackageTrace) error {
	if trace.TraceID == "" {
		trace.TraceID = r.idGen.GenerateTraceID()
	}
	if trace.OperationTime.IsZero() {
		trace.OperationTime = time.Now()
	}
	return r.db.Create(trace).Error
}

// GetTracesByPackageID 获取包裹轨迹
func (r *packageRepository) GetTracesByPackageID(packageID string) ([]model.PackageTrace, error) {
	var traces []model.PackageTrace
	if err := r.db.Where("package_id = ?", packageID).
		Order("operation_time ASC").
		Find(&traces).Error; err != nil {
		return nil, err
	}
	return traces, nil
}

// CreateAbnormalRecord 创建异常记录
func (r *packageRepository) CreateAbnormalRecord(record *model.AbnormalRecord) error {
	if record.RecordID == "" {
		record.RecordID = r.idGen.GenerateAbnormalRecordID()
	}
	return r.db.Create(record).Error
}
