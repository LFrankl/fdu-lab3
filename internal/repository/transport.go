package repository

import (
	"errors"
	"fmt"
	"time"

	"github.com/LFrankl/fdu-lab3/internal/model"
	"github.com/LFrankl/fdu-lab3/pkg/db"
	"github.com/LFrankl/fdu-lab3/pkg/errno"
	"gorm.io/gorm"
)

// TransportRepo 运输领域数据访问接口
type TransportRepo interface {
	// CreateTask 创建运输任务
	CreateTask(task *model.TransportTask) error
	// GetTaskByID 根据ID查询运输任务
	GetTaskByID(taskID string) (*model.TransportTask, error)
	// GetTasksByDriverID 根据司机ID查询任务
	GetTasksByDriverID(driverID string, status string) ([]*model.TransportTask, error)
	// UpdateTask 更新运输任务
	UpdateTask(task *model.TransportTask) error
	// BindPackages 绑定包裹到运输任务
	BindPackages(taskID string, packageIDs []string) error
	// GetPackageIDsByTaskID 查询运输任务绑定的包裹列表
	GetPackageIDsByTaskID(taskID string) ([]string, error)
	// CountPackagesByTaskID 统计运输任务包裹数量
	CountPackagesByTaskID(taskID string) (int, error)
}

// transportRepo 实现TransportRepo接口
type transportRepo struct{}

func NewTransportRepo() TransportRepo {
	return &transportRepo{}
}

// CreateTask 创建运输任务
func (r *transportRepo) CreateTask(task *model.TransportTask) error {
	return db.DB.Create(task).Error
}

// GetTaskByID 根据ID查询运输任务
func (r *transportRepo) GetTaskByID(taskID string) (*model.TransportTask, error) {
	var task model.TransportTask
	if err := db.DB.Where("task_id = ?", taskID).First(&task).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errno.ErrTransportTaskNotFound
		}
		return nil, err
	}
	return &task, nil
}

// GetTasksByDriverID 根据司机ID查询任务
func (r *transportRepo) GetTasksByDriverID(driverID string, status string) ([]*model.TransportTask, error) {
	var tasks []*model.TransportTask
	query := db.DB.Where("driver_id = ?", driverID)
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if err := query.Find(&tasks).Error; err != nil {
		return nil, err
	}
	return tasks, nil
}

// UpdateTask 更新运输任务
func (r *transportRepo) UpdateTask(task *model.TransportTask) error {
	return db.DB.Updates(task).Error
}

// GetBoundPackageCount 查询任务已绑定的包裹总数
func (r *transportRepo) GetBoundPackageCount(taskID string) (int, error) {
	if taskID == "" {
		return 0, errors.New("任务ID不能为空")
	}
	var count int64
	err := db.DB.Model(&model.TransportTaskPackage{}).
		Where("transport_task_id = ?", taskID).
		Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("查询包裹数量失败：%w", err)
	}
	return int(count), nil
}

// UpdatePackageCount 单独更新任务的包裹数量
func (r *transportRepo) UpdatePackageCount(taskID string, count int) error {
	if taskID == "" {
		return errors.New("任务ID不能为空")
	}
	return db.DB.Model(&model.TransportTask{}).
		Where("task_id = ?", taskID).
		Update("package_count", count).Error
}

// BindPackages 绑定包裹到运输任务
func (r *transportRepo) BindPackages(taskID string, packageIDs []string) error {
	// 1. 入参基础校验
	if taskID == "" {
		return errors.New("运输任务ID不能为空")
	}
	if len(packageIDs) == 0 {
		return errors.New("包裹ID列表不能为空")
	}

	// 2. 查询该任务已绑定的包裹ID（用于过滤重复）
	var existPkgIDs []string
	err := db.DB.Model(&model.TransportTaskPackage{}).
		Where("transport_task_id = ?", taskID).
		Pluck("package_id", &existPkgIDs).Error
	if err != nil {
		return fmt.Errorf("查询已绑定包裹失败：%w", err)
	}

	// 3. 构建已绑定包裹的映射（快速去重）
	existPkgMap := make(map[string]struct{}, len(existPkgIDs))
	for _, pkgID := range existPkgIDs {
		existPkgMap[pkgID] = struct{}{}
	}

	// 4. 过滤重复/空包裹ID，仅保留未绑定的新包裹
	var newTaskPackages []*model.TransportTaskPackage
	for _, pkgID := range packageIDs {
		if pkgID == "" { // 跳过空ID
			continue
		}
		if _, exists := existPkgMap[pkgID]; exists { // 跳过已绑定的ID
			continue
		}
		newTaskPackages = append(newTaskPackages, &model.TransportTaskPackage{
			TransportTaskID: taskID,
			PackageID:       pkgID,
			AddedTime:       time.Now(),
		})
	}

	// 5. 无新包裹需绑定，直接返回（避免空插入）
	if len(newTaskPackages) == 0 {
		return nil
	}

	// 6. 批量新增新包裹关联
	batchSize := 100
	if len(newTaskPackages) <= batchSize {
		err = db.DB.Create(newTaskPackages).Error
	} else {
		err = db.DB.CreateInBatches(newTaskPackages, batchSize).Error
	}
	if err != nil {
		return fmt.Errorf("新增包裹关联失败：%w", err)
	}

	// 7. 核心：重新计算并更新任务的包裹数量
	totalCount, err := r.CountPackagesByTaskID(taskID)
	if err != nil {
		return fmt.Errorf("获取包裹总数失败：%w", err)
	}
	if err = r.UpdatePackageCount(taskID, totalCount); err != nil {
		return fmt.Errorf("更新包裹数量失败：%w", err)
	}

	return nil
}

// GetPackageIDsByTaskID 查询运输任务绑定的包裹列表
func (r *transportRepo) GetPackageIDsByTaskID(taskID string) ([]string, error) {
	var pkgIDs []string
	err := db.DB.Model(&model.TransportTaskPackage{}).
		Where("transport_task_id = ?", taskID).
		Pluck("package_id", &pkgIDs).Error
	return pkgIDs, err
}

// CountPackagesByTaskID 统计运输任务包裹数量
func (r *transportRepo) CountPackagesByTaskID(taskID string) (int, error) {
	var count int64
	err := db.DB.Model(&model.TransportTaskPackage{}).
		Where("transport_task_id = ?", taskID).
		Count(&count).Error
	return int(count), err
}
