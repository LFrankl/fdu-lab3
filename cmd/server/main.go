package main

import (
	"fmt"
	"log"

	"github.com/LFrankl/fdu-lab3/config"
	"github.com/LFrankl/fdu-lab3/internal/api/router"
	"github.com/LFrankl/fdu-lab3/internal/model"
	"github.com/LFrankl/fdu-lab3/pkg/db"
)

// @title 快递物流包裹处理与运输调度系统API
// @version 1.0
// @description 基于Golang实现的快递物流系统API
// @host localhost:8080
// @BasePath /api/v1
func main() {
	// 加载配置
	if err := config.Load("config/app.yaml"); err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 初始化日志
	//logger.Init(config.Cfg.App.Env)

	// 初始化数据库
	if err := db.InitMySQL(); err != nil {
		log.Fatalf("初始化MySQL失败: %v", err)
	}

	// 初始化Redis（可选）
	//if err := db.InitRedis(); err != nil {
	//	log.Fatalf("初始化Redis失败: %v", err)
	//}

	// 自动迁移表结构
	if err := db.DB.AutoMigrate(
		&model.Package{},
		&model.PackageTrace{},
		&model.AbnormalRecord{},
		//&model.TransportTask{},
		//&model.TransportTaskPackage{},
		//&model.DeliveryTask{},
		//&model.DeliveryTaskPackage{},
	); err != nil {
		log.Fatalf("表结构迁移失败: %v", err)
	}

	// 配置路由
	r := router.SetupRouter()

	// 启动服务
	port := config.Cfg.App.Port
	log.Printf("服务器启动成功，监听端口：%d", port)
	if err := r.Run(fmt.Sprintf(":%d", port)); err != nil {
		log.Fatalf("服务器启动失败: %v", err)
	}
}
