package handler

import (
	"net/http"

	"github.com/LFrankl/fdu-lab3/internal/model"
	"github.com/LFrankl/fdu-lab3/internal/service"
	"github.com/gin-gonic/gin"
)

// PackageHandler 包裹API处理器
type PackageHandler struct {
	pkgService service.PackageService
}

// NewPackageHandler 创建包裹处理器
func NewPackageHandler() *PackageHandler {
	return &PackageHandler{
		pkgService: service.NewPackageService(),
	}
}

// CreatePackageRequest 创建包裹请求体
type CreatePackageRequest struct {
	SenderName       string  `json:"sender_name" binding:"required"`
	SenderPhone      string  `json:"sender_phone" binding:"required"`
	SenderAddress    string  `json:"sender_address" binding:"required"`
	ReceiverName     string  `json:"receiver_name" binding:"required"`
	ReceiverPhone    string  `json:"receiver_phone" binding:"required"`
	ReceiverAddress  string  `json:"receiver_address" binding:"required"`
	ReceiverProvince string  `json:"receiver_province" binding:"required"`
	ReceiverCity     string  `json:"receiver_city" binding:"required"`
	ReceiverDistrict string  `json:"receiver_district" binding:"required"`
	Weight           float64 `json:"weight" binding:"required"`
	Length           float64 `json:"length"`
	Width            float64 `json:"width"`
	Height           float64 `json:"height"`
}

// CreatePackage 创建包裹（揽收）
// @Summary 创建包裹
// @Description 网点工作人员创建包裹记录（揽收）
// @Tags 包裹管理
// @Accept json
// @Produce json
// @Param request body CreatePackageRequest true "包裹信息"
// @Param operator header string true "操作人"
// @Param node_name header string true "节点名称"
// @Param node_address header string true "节点地址"
// @Success 200 {object} gin.H{"code":0,"msg":"success","data":{}}
// @Failure 400 {object} gin.H{"code":400,"msg":"参数错误","data":nil}
// @Failure 500 {object} gin.H{"code":500,"msg":"服务器错误","data":nil}
// @Router /packages [post]
func (h *PackageHandler) CreatePackage(c *gin.Context) {
	var req CreatePackageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "参数错误：" + err.Error(),
			"data": nil,
		})
		return
	}

	// 获取请求头信息
	operator := c.GetHeader("operator")
	nodeName := c.GetHeader("node_name")
	nodeAddr := c.GetHeader("node_address")

	if operator == "" || nodeName == "" || nodeAddr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "操作人/节点信息不能为空",
			"data": nil,
		})
		return
	}

	// 构建包裹模型
	pkg := &model.Package{
		SenderName:       req.SenderName,
		SenderPhone:      req.SenderPhone,
		SenderAddress:    req.SenderAddress,
		ReceiverName:     req.ReceiverName,
		ReceiverPhone:    req.ReceiverPhone,
		ReceiverAddress:  req.ReceiverAddress,
		ReceiverProvince: req.ReceiverProvince,
		ReceiverCity:     req.ReceiverCity,
		ReceiverDistrict: req.ReceiverDistrict,
		Weight:           req.Weight,
		Length:           req.Length,
		Width:            req.Width,
		Height:           req.Height,
	}

	// 创建包裹
	result, err := h.pkgService.CreatePackage(pkg, operator, nodeName, nodeAddr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "创建包裹失败：" + err.Error(),
			"data": nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "success",
		"data": result,
	})
}

// GetPackageDetail 获取包裹详情
// @Summary 获取包裹详情
// @Description 根据运单号查询包裹详情及轨迹
// @Tags 包裹管理
// @Accept json
// @Produce json
// @Param package_id path string true "运单号"
// @Success 200 {object} gin.H{"code":0,"msg":"success","data":{}}
// @Failure 404 {object} gin.H{"code":404,"msg":"包裹不存在","data":nil}
// @Failure 500 {object} gin.H{"code":500,"msg":"服务器错误","data":nil}
// @Router /packages/{package_id} [get]
func (h *PackageHandler) GetPackageDetail(c *gin.Context) {
	packageID := c.Param("package_id")
	if packageID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "运单号不能为空",
			"data": nil,
		})
		return
	}

	detail, err := h.pkgService.GetPackageDetail(packageID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code": 404,
			"msg":  err.Error(),
			"data": nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "success",
		"data": detail,
	})
}

// HandleSortingAbnormalRequest 分拣异常处理请求
type HandleSortingAbnormalRequest struct {
	Reason string `json:"reason" binding:"required"`
}

// HandleSortingAbnormal 处理分拣异常
// @Summary 处理分拣异常
// @Description 记录分拣异常信息并更新包裹状态
// @Tags 包裹管理
// @Accept json
// @Produce json
// @Param package_id path string true "运单号"
// @Param request body HandleSortingAbnormalRequest true "异常信息"
// @Param handler header string true "处理人"
// @Success 200 {object} gin.H{"code":0,"msg":"success","data":nil}
// @Failure 400 {object} gin.H{"code":400,"msg":"参数错误","data":nil}
// @Failure 500 {object} gin.H{"code":500,"msg":"服务器错误","data":nil}
// @Router /packages/{package_id}/abnormal/sorting [post]
func (h *PackageHandler) HandleSortingAbnormal(c *gin.Context) {
	packageID := c.Param("package_id")
	if packageID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "运单号不能为空",
			"data": nil,
		})
		return
	}

	var req HandleSortingAbnormalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "参数错误：" + err.Error(),
			"data": nil,
		})
		return
	}

	handler := c.GetHeader("handler")
	if handler == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "处理人不能为空",
			"data": nil,
		})
		return
	}

	if err := h.pkgService.HandleSortingAbnormal(packageID, req.Reason, handler); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "处理异常失败：" + err.Error(),
			"data": nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "分拣异常已记录",
		"data": nil,
	})
}
