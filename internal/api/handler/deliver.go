package handler

import (
	"net/http"

	"github.com/LFrankl/fdu-lab3/internal/service"
	"github.com/LFrankl/fdu-lab3/pkg/errno"
	"github.com/gin-gonic/gin"
)

// DeliveryHandler 派送领域API处理
type DeliveryHandler struct {
	deliverySvc *service.DeliverySvc
}

func NewDeliveryHandler() *DeliveryHandler {
	return &DeliveryHandler{
		deliverySvc: service.NewDeliverySvc(),
	}
}

// CreateTask 创建派送任务
func (h *DeliveryHandler) CreateTask(c *gin.Context) {
	var req service.CreateDeliveryTaskReq
	if err := c.ShouldBindJSON(&req); err != nil {
		ResponseError(c, http.StatusBadRequest, errno.ErrParamInvalid)
		return
	}
	task, err := h.deliverySvc.CreateDeliveryTask(&req)
	if err != nil {
		ResponseError(c, http.StatusInternalServerError, err)
		return
	}
	ResponseSuccess(c, gin.H{"task_id": task.TaskID, "status": task.Status})
}

// ChangeTaskStatus 变更派送任务状态
func (h *DeliveryHandler) ChangeTaskStatus(c *gin.Context) {
	taskID := c.Param("task_id")
	var req struct {
		NewStatus string `json:"new_status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		ResponseError(c, http.StatusBadRequest, errno.ErrParamInvalid)
		return
	}
	if err := h.deliverySvc.ChangeTaskStatus(taskID, req.NewStatus); err != nil {
		ResponseError(c, http.StatusInternalServerError, err)
		return
	}
	ResponseSuccess(c, gin.H{"msg": "派送任务状态已更新"})
}

// BindPackages 绑定包裹到派送任务
func (h *DeliveryHandler) BindPackages(c *gin.Context) {
	taskID := c.Param("task_id")
	var req struct {
		PackageIDs []string `json:"package_ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		ResponseError(c, http.StatusBadRequest, errno.ErrParamInvalid)
		return
	}
	if err := h.deliverySvc.BindPackagesToTask(taskID, req.PackageIDs); err != nil {
		ResponseError(c, http.StatusInternalServerError, err)
		return
	}
	ResponseSuccess(c, gin.H{"msg": "包裹绑定成功", "package_count": len(req.PackageIDs)})
}

// ReportAbnormal 上报派送异常
func (h *DeliveryHandler) ReportAbnormal(c *gin.Context) {
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
	if err := h.deliverySvc.ReportDeliveryAbnormal(taskID, req.AbnormalType, req.Reason, req.Handler); err != nil {
		ResponseError(c, http.StatusInternalServerError, err)
		return
	}
	ResponseSuccess(c, gin.H{"msg": "派送异常已上报"})
}

// SignPackage 包裹签收
func (h *DeliveryHandler) SignPackage(c *gin.Context) {
	taskID := c.Param("task_id")
	packageID := c.Param("package_id")
	courierID := c.GetHeader("courier_id") // 从请求头获取派送员ID（身份认证后）
	var req struct {
		SignerName  string `json:"signer_name"`
		SignerPhone string `json:"signer_phone"`
		SignType    string `json:"sign_type"`
		Remark      string `json:"remark"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		ResponseError(c, http.StatusBadRequest, errno.ErrParamInvalid)
		return
	}
	if err := h.deliverySvc.SignPackage(taskID, packageID, courierID, req.SignerName, req.SignerPhone, req.SignType, req.Remark); err != nil {
		ResponseError(c, http.StatusInternalServerError, err)
		return
	}
	ResponseSuccess(c, gin.H{"msg": "包裹签收成功", "package_id": packageID})
}

// GetCourierTaskPackages 派送员查询任务包裹列表
func (h *DeliveryHandler) GetCourierTaskPackages(c *gin.Context) {
	courierID := c.GetHeader("courier_id")
	taskID := c.Param("task_id")
	packages, err := h.deliverySvc.GetCourierTaskPackages(courierID, taskID)
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
