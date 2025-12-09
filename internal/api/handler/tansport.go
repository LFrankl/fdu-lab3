package handler

import (
	"fmt"
	"net/http"

	"github.com/LFrankl/fdu-lab3/internal/service"
	"github.com/LFrankl/fdu-lab3/pkg/errno"
	"github.com/gin-gonic/gin"
)

// TransportHandler 运输领域API处理
type TransportHandler struct {
	transportSvc *service.TransportSvc
}

func NewTransportHandler() *TransportHandler {
	return &TransportHandler{
		transportSvc: service.NewTransportSvc(),
	}
}

// CreateTask 创建运输任务
// @Summary 创建运输任务
// @Description 运输调度员创建运输任务（关联车辆/司机/起止点）
// @Tags 运输任务管理
// @Accept json
// @Produce json
// @Param request body CreateTransportTaskRequest true "运输任务信息"
// @Param operator header string true "操作人（调度员ID）"
// @Success 200 {object} gin.H{"code":0,"msg":"任务创建成功","data":{"task_id":"TRAN20251208120000001","status":"pending"}}
// @Failure 400 {object} gin.H{"code":400,"msg":"参数错误","data":nil}
// @Failure 500 {object} gin.H{"code":500,"msg":"服务器错误","data":nil}
// @Router /transport/tasks [post]
func (h *TransportHandler) CreateTask(c *gin.Context) {
	var req service.CreateTransportTaskReq
	if err := c.ShouldBindJSON(&req); err != nil {
		fmt.Println(err)
		fmt.Printf("绑定后的req值：%+v\n", req)
		ResponseError(c, http.StatusBadRequest, errno.ErrParamInvalid)
		return
	}
	task, err := h.transportSvc.CreateTransportTask(&req)
	if err != nil {
		ResponseError(c, http.StatusInternalServerError, err)
		return
	}
	ResponseSuccess(c, gin.H{"task_id": task.TaskID, "status": task.Status})
}

// ChangeTaskStatus 变更运输任务状态
// @Summary 变更运输任务状态
// @Description 调度员/司机更新运输任务状态（仅支持pending→transporting→arrived→completed/abnormal的合法流转）
// @Tags 运输任务管理
// @Accept json
// @Produce json
// @Param task_id path string true "运输任务ID"
// @Param request body ChangeTransportTaskStatusRequest true "新状态信息"
// @Param operator header string true "操作人ID（司机/调度员）"
// @Success 200 {object} gin.H{"code":0,"msg":"运输任务状态已更新","data":{"task_id":"xxx","new_status":"transporting"}}
// @Failure 400 {object} gin.H{"code":400,"msg":"参数错误/状态流转不合法","data":nil}
// @Failure 404 {object} gin.H{"code":404,"msg":"运输任务不存在","data":nil}
// @Failure 500 {object} gin.H{"code":500,"msg":"服务器错误","data":nil}
// @Router /transport/tasks/{task_id}/status [put]
func (h *TransportHandler) ChangeTaskStatus(c *gin.Context) {
	taskID := c.Param("task_id")
	var req struct {
		NewStatus string `json:"new_status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		ResponseError(c, http.StatusBadRequest, errno.ErrParamInvalid)
		return
	}
	if err := h.transportSvc.ChangeTaskStatus(taskID, req.NewStatus); err != nil {
		ResponseError(c, http.StatusInternalServerError, err)
		return
	}
	ResponseSuccess(c, gin.H{"msg": "运输任务状态已更新"})
}

// BindPackages 绑定包裹到运输任务
// @Summary 绑定包裹到运输任务
// @Description 运输调度员将分拣完成的包裹绑定到指定运输任务（仅待执行/运输中任务可绑定）
// @Tags 运输任务管理
// @Accept json
// @Produce json
// @Param task_id path string true "运输任务ID"
// @Param request body BindPackagesRequest true "绑定的包裹ID列表"
// @Param operator header string true "操作人（调度员ID）"
// @Success 200 {object} gin.H{"code":0,"msg":"包裹绑定成功","data":{"task_id":"xxx","package_count":85}}
// @Failure 400 {object} gin.H{"code":400,"msg":"参数错误/任务状态不可绑定","data":nil}
// @Failure 500 {object} gin.H{"code":500,"msg":"服务器错误","data":nil}
// @Router /transport/tasks/{task_id}/packages/bind [post]
func (h *TransportHandler) BindPackages(c *gin.Context) {
	taskID := c.Param("task_id")
	var req struct {
		PackageIDs []string `json:"package_ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		ResponseError(c, http.StatusBadRequest, errno.ErrParamInvalid)
		return
	}
	if err := h.transportSvc.BindPackagesToTask(taskID, req.PackageIDs); err != nil {
		ResponseError(c, http.StatusInternalServerError, err)
		return
	}
	ResponseSuccess(c, gin.H{"msg": "包裹绑定成功", "package_count": len(req.PackageIDs)})
}

// GetDriverTaskPackages 司机查询任务包裹列表
// @Summary 司机查询运输任务包裹列表
// @Description 司机查询本人承接的运输任务下的所有包裹清单
// @Tags 司机端-运输任务
// @Accept json
// @Produce json
// @Param task_id path string true "运输任务ID"
// @Param driver_id header string true "司机ID（身份认证）"
// @Success 200 {object} gin.H{"code":0,"msg":"success","data":{"task_id":"xxx","package_count":85,"packages":[]}}
// @Failure 403 {object} gin.H{"code":403,"msg":"任务不属于该司机","data":nil}
// @Failure 500 {object} gin.H{"code":500,"msg":"服务器错误","data":nil}
// @Router /transport/tasks/{task_id}/packages [get]
func (h *TransportHandler) GetDriverTaskPackages(c *gin.Context) {
	driverID := c.GetHeader("driver_id") // 从请求头获取司机ID（身份认证后）
	taskID := c.Param("task_id")
	packages, err := h.transportSvc.GetDriverTaskPackages(driverID, taskID)
	if err != nil {
		ResponseError(c, http.StatusInternalServerError, err)
		return
	}
	ResponseSuccess(c, gin.H{
		"task_id":       taskID,
		"package_count": len(packages),
		"packages":      packages,
	})
}

// ReportAbnormal 上报运输异常
// @Summary 上报运输异常
// @Description 司机/调度员上报运输任务异常（如路线变更、车辆故障、包裹损坏等），标记任务为异常状态并同步包裹状态
// @Tags 运输任务管理
// @Accept json
// @Produce json
// @Param task_id path string true "运输任务ID"
// @Param request body ReportTransportAbnormalRequest true "运输异常信息"
// @Param operator header string true "操作人ID（司机/调度员）"
// @Success 200 {object} gin.H{"code":0,"msg":"运输异常已上报","data":{"task_id":"xxx","abnormal_type":"route_change"}}
// @Failure 400 {object} gin.H{"code":400,"msg":"参数错误","data":nil}
// @Failure 404 {object} gin.H{"code":404,"msg":"运输任务不存在","data":nil}
// @Failure 500 {object} gin.H{"code":500,"msg":"服务器错误","data":nil}
// @Router /transport/tasks/{task_id}/abnormal [post]
func (h *TransportHandler) ReportAbnormal(c *gin.Context) {
	taskID := c.Param("task_id")
	var req struct {
		AbnormalType string `json:"abnormal_type"`
		Reason       string `json:"reason"`
		Handler      string `json:"handler"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		ResponseError(c, http.StatusBadRequest, errno.ErrParamInvalid)
		return
	}
	if err := h.transportSvc.ReportTransportAbnormal(taskID, req.AbnormalType, req.Reason, req.Handler); err != nil {
		ResponseError(c, http.StatusInternalServerError, err)
		return
	}
	ResponseSuccess(c, gin.H{"msg": "运输异常已上报"})
}

func ResponseSuccess(c *gin.Context, data gin.H) {
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "success",
		"data": data,
	})
}

func ResponseError(c *gin.Context, code int, err error) {
	c.JSON(code, gin.H{
		"code": code,
		"msg":  err.Error(),
		"data": nil,
	})
}
