package util

import (
	"math/rand"
	"strconv"
	"time"
)

// IDGenerator ID生成器
type IDGenerator struct{}

// NewIDGenerator 创建ID生成器实例
func NewIDGenerator() *IDGenerator {
	return &IDGenerator{}
}

// GeneratePackageID 生成运单号：KD + 时间戳(14位) + 随机字符串(6位)
func (g *IDGenerator) GeneratePackageID() string {
	prefix := "KD"
	timestamp := time.Now().Format("20060102150405")
	randomStr := g.generateRandomString(6)
	return prefix + timestamp + randomStr
}

// GenerateTraceID 生成轨迹ID
func (g *IDGenerator) GenerateTraceID() string {
	prefix := "TR"
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	randomStr := g.generateRandomString(4)
	return prefix + timestamp + randomStr
}

// GenerateAbnormalRecordID 生成异常记录ID
func (g *IDGenerator) GenerateAbnormalRecordID() string {
	prefix := "AB"
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	randomStr := g.generateRandomString(4)
	return prefix + timestamp + randomStr
}

// GenerateTransportTaskID 生成运输任务ID
func (g *IDGenerator) GenerateTransportTaskID() string {
	prefix := "TRAN"
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	randomStr := g.generateRandomString(4)
	return prefix + timestamp + randomStr
}

// GenerateDeliveryTaskID 生成派送任务ID
func (g *IDGenerator) GenerateDeliveryTaskID() string {
	prefix := "DELIV"
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	randomStr := g.generateRandomString(4)
	return prefix + timestamp + randomStr
}

// 生成随机字符串
func (g *IDGenerator) generateRandomString(length int) string {
	chars := "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		result[i] = chars[r.Intn(len(chars))]
	}
	return string(result)
}
