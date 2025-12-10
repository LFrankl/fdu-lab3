package router

import (
	"github.com/LFrankl/fdu-lab3/config"
	"github.com/LFrankl/fdu-lab3/internal/api/handler"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// SetupRouter 配置路由
func SetupRouter() *gin.Engine {
	// 设置运行模式
	if config.Cfg.App.Env == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()

	// 中间件
	//r.Use(middleware.Logger())
	//r.Use(middleware.Cors())

	// Swagger文档
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// 初始化处理器
	pkgHandler := handler.NewPackageHandler()
	transportHandler := handler.NewTransportHandler()
	deliveryHandler := handler.NewDeliveryHandler()

	// API路由组
	api := r.Group("/api/v1")
	{
		// 包裹管理
		packages := api.Group("/packages")
		{
			packages.POST("", pkgHandler.CreatePackage)
			packages.GET("/:package_id", pkgHandler.GetPackageDetail)
			packages.POST("/sorting/:package_id", pkgHandler.Sorting)
			packages.POST("/:package_id/abnormal/sorting", pkgHandler.HandleSortingAbnormal)
			packages.POST("/:package_id/status", pkgHandler.ChangePackageStatus)
		}

		// 运输调度
		transport := api.Group("/transport")
		{
			// 创建运输任务
			transport.POST("/tasks", transportHandler.CreateTask)
			// 变更任务状态
			transport.PUT("/tasks/:task_id/status", transportHandler.ChangeTaskStatus)
			// 绑定包裹到任务
			transport.POST("/tasks/:task_id/packages/bind", transportHandler.BindPackages)
			// 司机查询任务包裹列表
			transport.GET("/tasks/:task_id/packages", transportHandler.GetDriverTaskPackages)
			// 上报运输异常
			transport.POST("/tasks/:task_id/abnormal", transportHandler.ReportAbnormal)
		}
		delivery := r.Group("/api/v1/delivery")
		{
			// 创建派送任务
			delivery.POST("/tasks", deliveryHandler.CreateTask)
			// 变更任务状态
			delivery.PUT("/tasks/:task_id/status", deliveryHandler.ChangeTaskStatus)
			// 绑定包裹到任务
			delivery.POST("/tasks/:task_id/packages/bind", deliveryHandler.BindPackages)
			// 上报派送异常
			delivery.POST("/tasks/:task_id/abnormal", deliveryHandler.ReportAbnormal)
			// 包裹签收
			delivery.POST("/tasks/:task_id/packages/:package_id/sign", deliveryHandler.SignPackage)
			// 派送员查询任务包裹列表
			delivery.GET("/tasks/:task_id/packages", deliveryHandler.GetCourierTaskPackages)
		}
		//
		//// 派送管理
		//delivery := api.Group("/delivery")
		//{
		//	delivery.POST("/tasks", deliveryHandler.CreateDeliveryTask)
		//	delivery.POST("/packages/:package_id/abnormal", deliveryHandler.HandleDeliveryAbnormal)
		//	delivery.GET("/couriers/:courier_id/tasks", deliveryHandler.GetCourierTasks)
		//}
	}

	return r
}
