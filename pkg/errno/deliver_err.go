package errno

import "fmt"

// 派送领域专属错误码
var (
	// ErrDeliveryStatusInvalid 状态相关
	ErrDeliveryStatusInvalid = fmt.Errorf("派送任务状态流转不合法")
	// ErrDeliveryTaskNotBindable 包裹绑定相关
	ErrDeliveryTaskNotBindable = fmt.Errorf("派送任务当前状态不可绑定包裹")
	// ErrDeliveryTaskNotAbnormal 异常相关
	ErrDeliveryTaskNotAbnormal = fmt.Errorf("派送任务非异常状态，无法处理异常")
	// ErrDeliveryTaskNotFound 数据操作相关
	ErrDeliveryTaskNotFound           = fmt.Errorf("派送任务不存在")
	ErrPackageNotBindToDeliveryTask   = fmt.Errorf("包裹未绑定到该派送任务")
	ErrDeliveryTaskNotBelongToCourier = fmt.Errorf("派送任务不属于该派送员")
	ErrPackageNotSigned               = fmt.Errorf("包裹未完成签收")
)
