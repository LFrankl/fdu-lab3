package service

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/LFrankl/fdu-lab3/internal/model"
	"github.com/LFrankl/fdu-lab3/internal/repository"
	"github.com/LFrankl/fdu-lab3/pkg/errno"
)

// TransportSvc 运输领域核心业务服务
type TransportSvc struct {
	transportRepo repository.TransportRepo
	packageRepo   repository.PackageRepository // 依赖包裹领域Repo（交互用）
}

func NewTransportSvc() *TransportSvc {
	return &TransportSvc{
		transportRepo: repository.NewTransportRepo(),
		packageRepo:   repository.NewPackageRepository(),
	}
}

// CreateTransportTask 创建运输任务（含基础校验）
func (s *TransportSvc) CreateTransportTask(req *CreateTransportTaskReq) (*model.TransportTask, error) {
	// 1. 校验必填参数
	if req.StartNode == "" || req.EndNode == "" || req.VehicleID == "" {
		return nil, errno.ErrParamInvalid
	}

	// 2. 构建运输任务模型
	task := &model.TransportTask{
		TaskID:        genTaskID(), // 生成唯一任务ID（需实现ID生成逻辑）
		StartNode:     req.StartNode,
		EndNode:       req.EndNode,
		Status:        "pending",
		VehicleID:     req.VehicleID,
		DriverID:      req.DriverID,
		DriverName:    req.DriverName,
		EstimatedTime: req.EstimatedTime,
		Route: model.TransportRoute{
			RouteJSON: req.RouteJSON,
			Distance:  req.Distance,
		},
	}
	// 3. 入库
	if err := s.transportRepo.CreateTask(task); err != nil {
		return nil, err
	}
	return task, nil
}

// ChangeTaskStatus 变更运输任务状态（含包裹状态同步）
func (s *TransportSvc) ChangeTaskStatus(taskID, newStatus string) error {
	// 1. 查询任务
	task, err := s.transportRepo.GetTaskByID(taskID)
	if err != nil {
		return err
	}
	// 2. 执行领域行为：状态变更
	if err := task.ChangeStatus(newStatus); err != nil {
		return err
	}
	// 3. 同步包裹状态（核心交互逻辑）
	if newStatus == "transporting" {
		// 运输中：同步包裹状态为transporting
		pkgIDs, err := s.transportRepo.GetPackageIDsByTaskID(taskID)
		if err != nil {
			return err
		}
		for _, pkgID := range pkgIDs {
			_ = s.packageRepo.UpdateStatus(pkgID, "transporting", "", "") // 调用包裹领域更新状态
		}
	} else if newStatus == "arrived" {
		// 已到站：同步包裹状态为arrived
		pkgIDs, err := s.transportRepo.GetPackageIDsByTaskID(taskID)
		if err != nil {
			return err
		}
		for _, pkgID := range pkgIDs {
			_ = s.packageRepo.UpdateStatus(pkgID, "arrived", "", "")
		}
	}
	// 4. 更新任务
	return s.transportRepo.UpdateTask(task)
}

// BindPackagesToTask 绑定包裹到运输任务（含包裹状态校验）
func (s *TransportSvc) BindPackagesToTask(taskID string, packageIDs []string) error {
	// 1. 查询任务
	task, err := s.transportRepo.GetTaskByID(taskID)
	if err != nil {
		return err
	}
	// 2. 校验包裹状态：仅已分拣（sorted）的包裹可绑定
	for _, pkgID := range packageIDs {
		pkg, err := s.packageRepo.GetByID(pkgID)
		if err != nil {
			return err
		}
		if pkg.Status != "sorted" {
			return fmt.Errorf("包裹%s状态为%s，仅已分拣包裹可绑定运输任务", pkgID, pkg.Status)
		}
	}
	// 3. 执行领域行为：绑定包裹
	if err := task.BindPackage(packageIDs); err != nil {
		fmt.Println(err)
		return err
	}
	// 4. 保存任务（更新包裹数量）
	//保存的时候，不小心把为空的一些时间也算做更新范围了
	if err := s.transportRepo.UpdateTask(task); err != nil {
		fmt.Println("更新任务报错：", err)
		fmt.Printf("更新的task对象：%+v\n", task)
		return err
	}
	// 5. 保存关联关系
	return s.transportRepo.BindPackages(taskID, packageIDs)
}

// ReportTransportAbnormal 上报运输异常
func (s *TransportSvc) ReportTransportAbnormal(taskID, abnormalType, reason, handler string) error {
	// 1. 查询任务
	task, err := s.transportRepo.GetTaskByID(taskID)
	if err != nil {
		return err
	}
	// 2. 执行领域行为：上报异常
	task.ReportAbnormal(abnormalType, reason, handler)
	// 3. 同步包裹状态为transport_abnormal
	pkgIDs, err := s.transportRepo.GetPackageIDsByTaskID(taskID)
	if err != nil {
		return err
	}
	for _, pkgID := range pkgIDs {
		_ = s.packageRepo.UpdateStatus(pkgID, "transport_abnormal", reason, handler)
	}
	// 4. 更新任务
	return s.transportRepo.UpdateTask(task)
}

// GetDriverTaskPackages 司机查询本人任务的包裹列表（核心场景）
func (s *TransportSvc) GetDriverTaskPackages(driverID, taskID string) ([]*model.Package, error) {
	// 1. 校验任务归属：确保任务属于该司机
	task, err := s.transportRepo.GetTaskByID(taskID)
	if err != nil {
		return nil, err
	}
	if task.DriverID != driverID {
		return nil, errno.ErrTransportTaskNotBelongToDriver
	}
	// 2. 查询任务绑定的包裹ID
	pkgIDs, err := s.transportRepo.GetPackageIDsByTaskID(taskID)
	if err != nil {
		return nil, err
	}
	// 3. 查询包裹详情（调用包裹领域）
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

// CreateTransportTaskReq 创建运输任务请求参数
type CreateTransportTaskReq struct {
	StartNode     string    `json:"start_node"`
	EndNode       string    `json:"end_node"`
	VehicleID     string    `json:"vehicle_id"`
	DriverID      string    `json:"driver_id"`
	DriverName    string    `json:"driver_name"`
	EstimatedTime time.Time `json:"estimated_time"`
	RouteJSON     string    `json:"route_json"`
	Distance      float64   `json:"distance"`
}

// genTaskID 生成唯一运输任务ID（示例实现）
func genTaskID() string {
	// 时间部分（精确到秒）+ 4位随机大写字母/数字
	return "TRAN" + time.Now().Format("20060102150405") + randString(4)
}

// randString 生成指定长度随机字符串（仅用math/rand，简单高效）
func randString(n int) string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	// 初始化随机源（只需要执行一次，放函数里也不影响课设使用）
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, n)
	for i := range b {
		// 随机取字符池中的字符
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}
