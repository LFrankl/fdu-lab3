package service

import (
	"fmt"
	"time"

	"github.com/LFrankl/fdu-lab3/internal/model"
	"github.com/LFrankl/fdu-lab3/internal/repository"
	"github.com/LFrankl/fdu-lab3/internal/util"
)

// PackageService 包裹业务接口
type PackageService interface {
	CreatePackage(pkg *model.Package, operator, nodeName, nodeAddr string) (*model.Package, error)
	GetPackageDetail(packageID string) (map[string]interface{}, error)
	HandleSortingAbnormal(packageID, reason, handler string) error
	ChangeStatus(packageID string, status string) error
}

// packageService 实现
type packageService struct {
	pkgRepo  repository.PackageRepository
	geoUtils *util.GeoUtils
	idGen    *util.IDGenerator
}

// NewPackageService 创建包裹服务实例
func NewPackageService() PackageService {
	return &packageService{
		pkgRepo:  repository.NewPackageRepository(),
		geoUtils: util.NewGeoUtils(),
		idGen:    util.NewIDGenerator(),
	}
}

// ChangeStatus 更改状态
// 这里我们认为是提供一般性状态变更，异常不走这里
func (s *packageService) ChangeStatus(packageID string, status string) error {
	//调用底层的updateStatus
	//id和status已在上层做过检验
	return s.pkgRepo.UpdateStatus(packageID, status, "", "")
}

// CreatePackage 创建包裹并添加揽收轨迹
func (s *packageService) CreatePackage(pkg *model.Package, operator, nodeName, nodeAddr string) (*model.Package, error) {
	// 生成运单号
	if pkg.PackageID == "" {
		pkg.PackageID = s.idGen.GeneratePackageID()
	}
	// 这里初始化为collected只有后续检查后，才会变成sorted
	pkg.Status = "collected"

	// 创建包裹
	if err := s.pkgRepo.Create(pkg); err != nil {
		return nil, err
	}

	// 获取节点经纬度
	lng, lat, err := s.geoUtils.GetCoordinates(nodeAddr)
	if err != nil {
		lng, lat = 0, 0 // 解析失败使用默认值
	}

	// 创建揽收轨迹
	trace := &model.PackageTrace{
		PackageID:     pkg.PackageID,
		NodeType:      "collection",
		NodeName:      nodeName,
		NodeAddress:   nodeAddr,
		Longitude:     lng,
		Latitude:      lat,
		OperationTime: time.Now(),
		Operator:      operator,
		Remark:        "包裹已揽收",
	}

	if err := s.pkgRepo.CreateTrace(trace); err != nil {
		return nil, err
	}

	return pkg, nil
}

// GetPackageDetail 获取包裹详情（含轨迹）
func (s *packageService) GetPackageDetail(packageID string) (map[string]interface{}, error) {
	// 获取包裹基本信息
	pkg, err := s.pkgRepo.GetByID(packageID)
	if err != nil {
		return nil, fmt.Errorf("包裹不存在: %v", err)
	}

	// 获取轨迹
	traces, err := s.pkgRepo.GetTracesByPackageID(packageID)
	if err != nil {
		return nil, err
	}

	// 解析轨迹
	//traceList := make([]map[string]interface{}, 0, len(traces))
	//var currentNode *model.PackageTrace
	//var nextNode *model.PackageTrace
	//
	//for i, t := range traces {
	//	traceList = append(traceList, map[string]interface{}{
	//		"node_type":      t.NodeType,
	//		"node_name":      t.NodeName,
	//		"operation_time": t.OperationTime.Format("2006-01-02 15:04:05"),
	//		"remark":         t.Remark,
	//	})
	//	currentNode = &traces[i]
	//	if i < len(traces)-1 {
	//		nextNode = &traces[i+1]
	//	} else {
	//		nextNode = nil
	//	}
	//}
	//
	//// 构建返回结果
	//result := map[string]interface{}{
	//	"package_id": pkg.PackageID,
	//	"sender_info": map[string]interface{}{
	//		"name":    pkg.SenderName,
	//		"phone":   pkg.SenderPhone,
	//		"address": pkg.SenderAddress,
	//	},
	//	"receiver_info": map[string]interface{}{
	//		"name":    pkg.ReceiverName,
	//		"phone":   pkg.ReceiverPhone,
	//		"address": pkg.ReceiverAddress,
	//	},
	//	"current_status":         pkg.Status,
	//	"current_position":       currentNode.NodeName,
	//	"next_node":              nextNode.NodeName,
	//	"estimated_arrival_time": "2025-12-01 10:00", // 实际场景需计算
	//	"trace_history":          traceList,
	//}
	// 解析轨迹
	traceList := make([]map[string]interface{}, 0, len(traces))
	var currentNode *model.PackageTrace
	var nextNode *model.PackageTrace

	// 第一步：先判断traces是否为空，避免空列表导致currentNode/nextNode为nil
	if len(traces) > 0 {
		for i, t := range traces {
			traceList = append(traceList, map[string]interface{}{
				"node_type":      t.NodeType,
				"node_name":      t.NodeName,
				"operation_time": t.OperationTime.Format("2006-01-02 15:04:05"),
				"remark":         t.Remark,
			})
			// 当前节点：取最后一条轨迹作为当前节点（更符合业务逻辑）
			currentNode = &traces[i]
			// 下一个节点：只有i不是最后一个时才赋值
			if i < len(traces)-1 {
				nextNode = &traces[i+1]
			} else {
				// 最后一个节点：下一个节点置nil
				nextNode = nil
			}
		}
	}

	// 第二步：安全获取当前节点/下一个节点名称（核心修复）
	var currentNodeName string
	if currentNode != nil {
		currentNodeName = currentNode.NodeName
	} else {
		currentNodeName = "暂无轨迹信息" // 兜底值
	}

	var nextNodeName string
	if nextNode != nil {
		nextNodeName = nextNode.NodeName
	} else {
		nextNodeName = "暂无后续节点" // 兜底值
	}

	// 构建返回结果（使用安全的节点名称）
	result := map[string]interface{}{
		"package_id": pkg.PackageID,
		"sender_info": map[string]interface{}{
			"name":    pkg.SenderName,
			"phone":   pkg.SenderPhone,
			"address": pkg.SenderAddress,
		},
		"receiver_info": map[string]interface{}{
			"name":    pkg.ReceiverName,
			"phone":   pkg.ReceiverPhone,
			"address": pkg.ReceiverAddress,
		},
		"current_status":         pkg.Status,
		"current_position":       currentNodeName,    // 替换为安全值
		"next_node":              nextNodeName,       // 替换为安全值
		"estimated_arrival_time": "2025-12-01 10:00", // 实际场景需计算
		"trace_history":          traceList,
	}

	return result, nil

}

// HandleSortingAbnormal 处理分拣异常
// 这里直接给上层调用，功能是 更新对应包裹的状态为不正常，然后，把异常传到db上传
func (s *packageService) HandleSortingAbnormal(packageID, reason, handler string) error {
	// 更新包裹状态
	if err := s.pkgRepo.UpdateStatus(packageID, "abnormal", reason, handler); err != nil {
		return err
	}

	// 创建异常记录
	abnormalRecord := &model.AbnormalRecord{
		PackageID:      packageID,
		AbnormalType:   "sorting",
		AbnormalReason: reason,
		Processor:      handler,
		Status:         "pending",
	}

	if err := s.pkgRepo.CreateAbnormalRecord(abnormalRecord); err != nil {
		return err
	}

	// 创建异常轨迹
	trace := &model.PackageTrace{
		PackageID:     packageID,
		NodeType:      "abnormal",
		NodeName:      "分拣中心",
		OperationTime: time.Now(),
		Operator:      handler,
		Remark:        fmt.Sprintf("分拣异常：%s", reason),
	}

	return s.pkgRepo.CreateTrace(trace)
}
