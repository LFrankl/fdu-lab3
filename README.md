# fdu-lab3
复旦大学2025fall高级软开作业repo（小组14

test

1. 功能验证要点

功能模块	验证方式	预期结果
包裹创建	调用 POST /api/v1/packages	成功创建包裹并生成运单号，添加揽收轨迹
包裹轨迹查询	调用 GET /api/v1/packages/{pkg_id}	返回包裹详情及完整轨迹
分拣异常处理	调用 POST /packages/{id}/abnormal	包裹状态更新为异常，生成异常记录和轨迹
运输任务创建	调用 POST /transport/tasks	成功创建运输任务并关联包裹
派送任务查询	调用 GET /delivery/couriers/{id}/tasks	返回派送员的派送任务列表
