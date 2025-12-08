package main

import (
	"fmt"
	"time"

	"github.com/LFrankl/fdu-lab3/config"
	"github.com/LFrankl/fdu-lab3/internal/model"
	"github.com/LFrankl/fdu-lab3/pkg/db"
	"gorm.io/gorm/clause"
)

// 初始化测试数据
func initTestData() error {
	// 1. 加载配置（需确保config/app.yaml配置正确）
	if err := config.Load("config/app.yaml"); err != nil {
		return fmt.Errorf("加载配置失败: %v", err)
	}

	// 2. 初始化MySQL连接
	if err := db.InitMySQL(); err != nil {
		return fmt.Errorf("初始化MySQL失败: %v", err)
	}

	// 3. 自动迁移表结构（确保表存在）
	if err := db.DB.AutoMigrate(
		&model.Package{},
		&model.PackageTrace{},
		&model.AbnormalRecord{},
	); err != nil {
		return fmt.Errorf("表结构迁移失败: %v", err)
	}

	packageID := "KD20251208120000FULL001" // 全链路测试运单号
	pkg := &model.Package{
		PackageID:        packageID,
		SenderName:       "王小明",
		SenderPhone:      "13812345678",
		SenderAddress:    "浙江省杭州市西湖区文三路478号",
		ReceiverName:     "李小华",
		ReceiverPhone:    "13987654321",
		ReceiverAddress:  "广东省深圳市南山区科技园10栋",
		ReceiverProvince: "广东省",
		ReceiverCity:     "深圳市",
		ReceiverDistrict: "南山区",
		Weight:           1.8,
		Length:           25.0,
		Width:            15.0,
		Height:           8.0,
		Status:           "delivered", // 最终状态：已签收
		AbnormalReason:   "",          // 异常已处理，清空原因
		AbnormalHandler:  "",
		CreatedAt:        time.Date(2025, 12, 8, 9, 0, 0, 0, time.Local),
		UpdatedAt:        time.Date(2025, 12, 8, 18, 0, 0, 0, time.Local),
	}
	if err := db.DB.Clauses(clause.OnConflict{DoNothing: true}).Create(pkg).Error; err != nil {
		return fmt.Errorf("插入包裹数据失败: %v", err)
	}

	// ========== 2. 包裹全生命周期轨迹数据（7个核心节点） ==========
	traces := []*model.PackageTrace{
		// 节点1：揽收
		{
			TraceID:       "TR20251208090000001",
			PackageID:     packageID,
			NodeType:      "collection",
			NodeName:      "杭州文三路网点",
			NodeAddress:   "浙江省杭州市西湖区文三路478号",
			Longitude:     120.1307,
			Latitude:      30.2798,
			OperationTime: time.Date(2025, 12, 8, 9, 0, 0, 0, time.Local),
			Operator:      "揽收员-张芳",
			Remark:        "包裹已揽收，等待分拣",
			CreatedAt:     time.Date(2025, 12, 8, 9, 0, 0, 0, time.Local),
		},
		// 节点2：分拣（异常节点）
		{
			TraceID:       "TR20251208100000002",
			PackageID:     packageID,
			NodeType:      "sorting",
			NodeName:      "杭州分拣中心",
			NodeAddress:   "浙江省杭州市萧山区分拣产业园",
			Longitude:     120.2599,
			Latitude:      30.1798,
			OperationTime: time.Date(2025, 12, 8, 10, 0, 0, 0, time.Local),
			Operator:      "分拣员-李军",
			Remark:        "分拣时发现收件人电话缺失，标记异常",
			CreatedAt:     time.Date(2025, 12, 8, 10, 0, 0, 0, time.Local),
		},
		// 节点3：异常处理完成
		{
			TraceID:       "TR20251208110000003",
			PackageID:     packageID,
			NodeType:      "abnormal_process",
			NodeName:      "杭州分拣中心",
			NodeAddress:   "浙江省杭州市萧山区分拣产业园",
			Longitude:     120.2599,
			Latitude:      30.1798,
			OperationTime: time.Date(2025, 12, 8, 11, 0, 0, 0, time.Local),
			Operator:      "客服-王丽",
			Remark:        "联系寄件人补充收件人电话：13987654321，异常处理完成",
			CreatedAt:     time.Date(2025, 12, 8, 11, 0, 0, 0, time.Local),
		},
		// 节点4：运输中
		{
			TraceID:       "TR20251208120000004",
			PackageID:     packageID,
			NodeType:      "transporting",
			NodeName:      "浙A12345运输车辆",
			NodeAddress:   "杭甬高速-杭州段",
			Longitude:     120.3898,
			Latitude:      30.1298,
			OperationTime: time.Date(2025, 12, 8, 12, 0, 0, 0, time.Local),
			Operator:      "司机-赵刚",
			Remark:        "包裹已装车，从杭州分拣中心发往深圳分拣中心",
			CreatedAt:     time.Date(2025, 12, 8, 12, 0, 0, 0, time.Local),
		},
		// 节点5：到站
		{
			TraceID:       "TR20251209080000005",
			PackageID:     packageID,
			NodeType:      "arrived",
			NodeName:      "深圳分拣中心",
			NodeAddress:   "广东省深圳市宝安区分拣产业园",
			Longitude:     113.8898,
			Latitude:      22.5598,
			OperationTime: time.Date(2025, 12, 9, 8, 0, 0, 0, time.Local),
			Operator:      "仓管-陈亮",
			Remark:        "包裹到达深圳分拣中心，等待派送分配",
			CreatedAt:     time.Date(2025, 12, 9, 8, 0, 0, 0, time.Local),
		},
		// 节点6：派送中
		{
			TraceID:       "TR20251209140000006",
			PackageID:     packageID,
			NodeType:      "delivering",
			NodeName:      "深圳南山派送站",
			NodeAddress:   "广东省深圳市南山区科技园8栋",
			Longitude:     113.9898,
			Latitude:      22.5398,
			OperationTime: time.Date(2025, 12, 9, 14, 0, 0, 0, time.Local),
			Operator:      "派送员-刘阳",
			Remark:        "包裹已出库，正在派送中",
			CreatedAt:     time.Date(2025, 12, 9, 14, 0, 0, 0, time.Local),
		},
		// 节点7：签收
		{
			TraceID:       "TR20251209160000007",
			PackageID:     packageID,
			NodeType:      "delivered",
			NodeName:      "深圳南山科技园10栋",
			NodeAddress:   "广东省深圳市南山区科技园10栋",
			Longitude:     113.9998,
			Latitude:      22.5298,
			OperationTime: time.Date(2025, 12, 9, 16, 0, 0, 0, time.Local),
			Operator:      "派送员-刘阳",
			Remark:        "包裹已签收，收件人：李小华",
			CreatedAt:     time.Date(2025, 12, 9, 16, 0, 0, 0, time.Local),
		},
	}
	// 批量插入轨迹数据
	if err := db.DB.Clauses(clause.OnConflict{DoNothing: true}).CreateInBatches(traces, len(traces)).Error; err != nil {
		return fmt.Errorf("批量插入轨迹数据失败: %v", err)
	}

	// ========== 初始化分拣异常记录（可选） ==========
	abnormal := &model.AbnormalRecord{
		RecordID:       "AB1733656800ABC1", // 自定义异常ID
		PackageID:      pkg.PackageID,
		AbnormalType:   "sorting",
		AbnormalReason: "地址标签模糊，需核对收件人信息",
		Processor:      "测试分拣员",
		Status:         "pending",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		ProcessingTime: time.Now(),
	}
	if err := db.DB.Clauses(clause.OnConflict{DoNothing: true}).Create(abnormal).Error; err != nil {
		return fmt.Errorf("插入异常数据失败: %v", err)
	}

	fmt.Println("测试数据初始化成功！")
	return nil
}

func main() {
	if err := initTestData(); err != nil {
		fmt.Printf("数据初始化失败: %v\n", err)
		return
	}
}
