package service

import (
	"fmt"
	"time"

	"github.com/LFrankl/fdu-lab3/internal/model"
	"github.com/LFrankl/fdu-lab3/internal/repository"
	"github.com/LFrankl/fdu-lab3/pkg/errno"
)

// DeliverySvc 派送领域核心业务服务
type DeliverySvc struct {
	deliveryRepo repository.DeliveryRepo
	packageRepo  repository.PackageRepository // 依赖包裹领域Repo
	//networkRepo  repository.NetworkRepo // 依赖网点领域Repo（校验派送区域）
}

func NewDeliverySvc() *DeliverySvc {
	return &DeliverySvc{
		deliveryRepo: repository.NewDeliveryRepo(),
		packageRepo:  repository.NewPackageRepository(),
		//networkRepo:  networkRepo,
	}
}

// CreateDeliveryTask 创建派送任务（含基础校验）
func (s *DeliverySvc) CreateDeliveryTask(req *CreateDeliveryTaskReq) (*model.DeliveryTask, error) {
	// 1. 校验必填参数
	if req.DeliveryArea == "" || req.CourierID == "" || req.StartNode == "" {
		return nil, errno.ErrParamInvalid
	}
	// 2. 校验派送员与派送区域匹配（联动网点领域）
	//areaCourier, err := s.networkRepo.CheckCourierArea(req.CourierID, req.DeliveryArea)
	//if err != nil || !areaCourier {
	//	return nil, fmt.Errorf("派送员%s不负责%s区域派送", req.CourierID, req.DeliveryArea)
	//}
	// 3. 构建派送任务模型
	task := &model.DeliveryTask{
		TaskID:       genDeliveryTaskID(), // 生成唯一任务ID
		DeliveryArea: req.DeliveryArea,
		CourierID:    req.CourierID,
		CourierName:  req.CourierName,
		Status:       "pending",
		StartNode:    req.StartNode,
	}
	// 4. 入库
	if err := s.deliveryRepo.CreateTask(task); err != nil {
		return nil, err
	}
	return task, nil
}

// ChangeTaskStatus 变更派送任务状态（含包裹状态同步）
func (s *DeliverySvc) ChangeTaskStatus(taskID, newStatus string) error {
	// 1. 查询任务
	task, err := s.deliveryRepo.GetTaskByID(taskID)
	if err != nil {
		return err
	}
	// 2. 执行领域行为：状态变更
	if err := task.ChangeStatus(newStatus); err != nil {
		return err
	}
	// 3. 同步包裹状态（核心交互逻辑）
	if newStatus == "delivering" {
		// 派送中：同步包裹状态为delivering
		pkgIDs, err := s.deliveryRepo.GetPackageIDsByTaskID(taskID)
		if err != nil {
			return err
		}
		for _, pkgID := range pkgIDs {
			_ = s.packageRepo.UpdateStatus(pkgID, "delivering", "", "")
		}
	} else if newStatus == "completed" {
		// 已完成：同步所有包裹状态为delivered（需先确认全部签收）
		pkgIDs, err := s.deliveryRepo.GetPackageIDsByTaskID(taskID)
		if err != nil {
			return err
		}
		for _, pkgID := range pkgIDs {
			dtp, err := s.deliveryRepo.GetDeliveryTaskPackage(taskID, pkgID)
			if err != nil {
				return err
			}
			if dtp.SignInfo.SignTime.IsZero() {
				return fmt.Errorf("包裹%s未签收，无法完成派送任务", pkgID)
			}
			_ = s.packageRepo.UpdateStatus(pkgID, "delivered", "", "")
		}
	}
	// 4. 更新任务
	return s.deliveryRepo.UpdateTask(task)
}

// BindPackagesToTask 绑定包裹到派送任务（含包裹状态校验）
func (s *DeliverySvc) BindPackagesToTask(taskID string, packageIDs []string) error {
	// 1. 查询任务
	task, err := s.deliveryRepo.GetTaskByID(taskID)
	if err != nil {
		return err
	}
	// 2. 校验包裹状态：仅已到站（arrived）的包裹可绑定
	for _, pkgID := range packageIDs {
		pkg, err := s.packageRepo.GetByID(pkgID)
		if err != nil {
			return err
		}
		if pkg.Status != "arrived" {
			return fmt.Errorf("包裹%s状态为%s，仅已到站包裹可绑定派送任务", pkgID, pkg.Status)
		}
	}
	// 3. 执行领域行为：绑定包裹
	if err := task.BindPackage(packageIDs); err != nil {
		return err
	}
	// 4. 保存任务（更新包裹数量）
	if err := s.deliveryRepo.UpdateTask(task); err != nil {
		return err
	}
	// 5. 保存关联关系
	return s.deliveryRepo.BindPackages(taskID, packageIDs)
}

// ReportDeliveryAbnormal 上报派送异常
func (s *DeliverySvc) ReportDeliveryAbnormal(taskID, abnormalType, reason, handler string) error {
	// 1. 查询任务
	task, err := s.deliveryRepo.GetTaskByID(taskID)
	if err != nil {
		return err
	}
	// 2. 执行领域行为：上报异常
	task.ReportAbnormal(abnormalType, reason, handler)
	// 3. 同步包裹状态为delivery_abnormal
	pkgIDs, err := s.deliveryRepo.GetPackageIDsByTaskID(taskID)
	if err != nil {
		return err
	}
	for _, pkgID := range pkgIDs {
		_ = s.packageRepo.UpdateStatus(pkgID, "delivery_abnormal", reason, handler)
	}
	// 4. 更新任务
	return s.deliveryRepo.UpdateTask(task)
}

// SignPackage 包裹签收（核心场景）
func (s *DeliverySvc) SignPackage(taskID, packageID, courierID, signerName, signerPhone, signType, remark string) error {
	// 1. 校验任务归属：确保任务属于该派送员
	task, err := s.deliveryRepo.GetTaskByID(taskID)
	if err != nil {
		return err
	}
	if task.CourierID != courierID {
		return errno.ErrDeliveryTaskNotBelongToCourier
	}
	// 2. 脱敏处理手机号（仅保留后4位）
	desensitizedPhone := desensitizePhone(signerPhone)
	// 3. 执行签收行为
	if err := s.deliveryRepo.SignPackage(taskID, packageID, signerName, desensitizedPhone, signType, remark); err != nil {
		return err
	}
	// 4. 同步包裹状态为delivered
	return s.packageRepo.UpdateStatus(packageID, "delivered", "", "")
}

// GetCourierTaskPackages 派送员查询本人任务的包裹列表
func (s *DeliverySvc) GetCourierTaskPackages(courierID, taskID string) ([]*model.Package, error) {
	// 1. 校验任务归属
	task, err := s.deliveryRepo.GetTaskByID(taskID)
	if err != nil {
		return nil, err
	}
	if task.CourierID != courierID {
		return nil, errno.ErrDeliveryTaskNotBelongToCourier
	}
	// 2. 查询任务绑定的包裹ID
	pkgIDs, err := s.deliveryRepo.GetPackageIDsByTaskID(taskID)
	if err != nil {
		return nil, err
	}
	// 3. 查询包裹详情（联动包裹领域）
	var packages []*model.Package
	for _, pkgID := range pkgIDs {
		pkg, err := s.packageRepo.GetByID(pkgID)
		if err != nil {
			return nil, err
		}
		packages = append(packages, pkg)
	}
	return packages, nil
}

// CreateDeliveryTaskReq ========== 辅助结构体/函数 ==========
// CreateDeliveryTaskReq 创建派送任务请求参数
type CreateDeliveryTaskReq struct {
	DeliveryArea string `json:"delivery_area"`
	CourierID    string `json:"courier_id"`
	CourierName  string `json:"courier_name"`
	StartNode    string `json:"start_node"`
}

// genDeliveryTaskID 生成唯一派送任务ID
func genDeliveryTaskID() string {
	return "DELI" + time.Now().Format("20060102150405") + randString(4)
}

// desensitizePhone 手机号脱敏（保留后4位）
func desensitizePhone(phone string) string {
	if len(phone) != 11 {
		return phone
	}
	return "*******" + phone[7:]
}
