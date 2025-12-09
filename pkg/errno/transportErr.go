package errno

import "fmt"

// 运输领域专属错误码
var (
	// ErrTransportStatusInvalid 状态相关
	ErrTransportStatusInvalid = fmt.Errorf("运输任务状态流转不合法")
	// ErrTransportTaskNotBindable 包裹绑定相关
	ErrTransportTaskNotBindable = fmt.Errorf("运输任务当前状态不可绑定包裹")
	// ErrTransportTaskNotAbnormal 异常相关
	ErrTransportTaskNotAbnormal = fmt.Errorf("运输任务非异常状态，无法处理异常")
	// ErrTransportTaskNotFound 数据操作相关
	ErrTransportTaskNotFound = fmt.Errorf("运输任务不存在")
	ErrPackageNotBindToTask  = fmt.Errorf("包裹未绑定到该运输任务")

	ErrParamInvalid = fmt.Errorf("参数无效")

	ErrTransportTaskNotBelongToDriver = fmt.Errorf("运输任务不属于该司机")
)
