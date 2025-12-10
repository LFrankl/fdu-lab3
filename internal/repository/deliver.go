package repository

import (
	"errors"
	"time"

	"github.com/LFrankl/fdu-lab3/internal/model"
	"github.com/LFrankl/fdu-lab3/pkg/db"
	"github.com/LFrankl/fdu-lab3/pkg/errno"
	"gorm.io/gorm"
)

// DeliveryRepo 派送领域数据访问接口
type DeliveryRepo interface {
	// CreateTask 创建派送任务
	CreateTask(task *model.DeliveryTask) error
	// GetTaskByID 根据ID查询派送任务
	GetTaskByID(taskID string) (*model.DeliveryTask, error)
	// GetTasksByCourierID 根据派送员ID查询任务
	GetTasksByCourierID(courierID string, status string) ([]*model.DeliveryTask, error)
	// UpdateTask 更新派送任务
	UpdateTask(task *model.DeliveryTask) error
	// BindPackages 绑定包裹到派送任务
	BindPackages(taskID string, packageIDs []string) error
	// GetPackageIDsByTaskID 查询派送任务绑定的包裹列表
	GetPackageIDsByTaskID(taskID string) ([]string, error)
	// CountPackagesByTaskID 统计派送任务包裹数量
	CountPackagesByTaskID(taskID string) (int, error)
	// SignPackage 包裹签收
	SignPackage(deliveryTaskID, packageID, signerName, signerPhone, signType, remark string) error
	// GetDeliveryTaskPackage 查询派送任务-包裹关联记录
	GetDeliveryTaskPackage(deliveryTaskID, packageID string) (*model.DeliveryTaskPackage, error)
}

// deliveryRepo 实现DeliveryRepo接口
type deliveryRepo struct{}

func NewDeliveryRepo() DeliveryRepo {
	return &deliveryRepo{}
}

// CreateTask 创建派送任务
func (r *deliveryRepo) CreateTask(task *model.DeliveryTask) error {
	return db.DB.Create(task).Error
}

// GetTaskByID 根据ID查询派送任务
func (r *deliveryRepo) GetTaskByID(taskID string) (*model.DeliveryTask, error) {
	var task model.DeliveryTask
	if err := db.DB.Where("task_id = ?", taskID).First(&task).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errno.ErrDeliveryTaskNotFound
		}
		return nil, err
	}
	return &task, nil
}

// GetTasksByCourierID 根据派送员ID查询任务
func (r *deliveryRepo) GetTasksByCourierID(courierID string, status string) ([]*model.DeliveryTask, error) {
	var tasks []*model.DeliveryTask
	query := db.DB.Where("courier_id = ?", courierID)
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if err := query.Find(&tasks).Error; err != nil {
		return nil, err
	}
	return tasks, nil
}

// UpdateTask 更新派送任务
func (r *deliveryRepo) UpdateTask(task *model.DeliveryTask) error {
	return db.DB.Save(task).Error
}

// BindPackages 绑定包裹到派送任务
func (r *deliveryRepo) BindPackages(taskID string, packageIDs []string) error {
	// 批量创建关联记录
	var taskPackages []*model.DeliveryTaskPackage
	for i, pkgID := range packageIDs {
		taskPackages = append(taskPackages, &model.DeliveryTaskPackage{
			DeliveryTaskID: taskID,
			PackageID:      pkgID,
			DeliveryOrder:  i + 1, // 按绑定顺序生成派送顺序
			AddedTime:      time.Now(),
		})
	}
	// 先删除旧关联，再新增（避免重复绑定）
	if err := db.DB.Where("delivery_task_id = ?", taskID).Delete(&model.DeliveryTaskPackage{}).Error; err != nil {
		return err
	}
	return db.DB.CreateInBatches(taskPackages, len(taskPackages)).Error
}

// GetPackageIDsByTaskID 查询派送任务绑定的包裹列表
func (r *deliveryRepo) GetPackageIDsByTaskID(taskID string) ([]string, error) {
	var pkgIDs []string
	err := db.DB.Model(&model.DeliveryTaskPackage{}).
		Where("delivery_task_id = ?", taskID).
		Order("delivery_order ASC"). // 按派送顺序排序
		Pluck("package_id", &pkgIDs).Error
	return pkgIDs, err
}

// CountPackagesByTaskID 统计派送任务包裹数量
func (r *deliveryRepo) CountPackagesByTaskID(taskID string) (int, error) {
	var count int64
	err := db.DB.Model(&model.DeliveryTaskPackage{}).
		Where("delivery_task_id = ?", taskID).
		Count(&count).Error
	return int(count), err
}

// SignPackage 包裹签收
func (r *deliveryRepo) SignPackage(deliveryTaskID, packageID, signerName, signerPhone, signType, remark string) error {
	var dtp model.DeliveryTaskPackage
	if err := db.DB.Where("delivery_task_id = ? AND package_id = ?", deliveryTaskID, packageID).First(&dtp).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errno.ErrPackageNotBindToDeliveryTask
		}
		return err
	}
	// 执行签收行为
	dtp.SignPackage(signerName, signerPhone, signType, remark)
	return db.DB.Save(&dtp).Error
}

// GetDeliveryTaskPackage 查询派送任务-包裹关联记录
func (r *deliveryRepo) GetDeliveryTaskPackage(deliveryTaskID, packageID string) (*model.DeliveryTaskPackage, error) {
	var dtp model.DeliveryTaskPackage
	if err := db.DB.Where("delivery_task_id = ? AND package_id = ?", deliveryTaskID, packageID).First(&dtp).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errno.ErrPackageNotBindToDeliveryTask
		}
		return nil, err
	}
	return &dtp, nil
}
