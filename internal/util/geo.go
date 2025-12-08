package util

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/LFrankl/fdu-lab3/config"
)

// GeoUtils 地理信息工具
type GeoUtils struct {
	amapKey string
}

// NewGeoUtils 创建地理信息工具实例
func NewGeoUtils() *GeoUtils {
	return &GeoUtils{
		amapKey: config.Cfg.Geo.AmapKey,
	}
}

// AmapGeoCodeResponse 高德地图地理编码响应
type AmapGeoCodeResponse struct {
	Status   string `json:"status"`
	Info     string `json:"info"`
	Geocodes []struct {
		FormattedAddress string `json:"formatted_address"`
		Location         string `json:"location"` // 经度,纬度
		Province         string `json:"province"`
		City             string `json:"city"`
		District         string `json:"district"`
	} `json:"geocodes"`
}

// GetCoordinates 通过地址获取经纬度
func (g *GeoUtils) GetCoordinates(address string) (lng, lat float64, err error) {
	apiURL := "https://restapi.amap.com/v3/geocode/geo"
	params := url.Values{}
	params.Set("address", address)
	params.Set("key", g.amapKey)
	params.Set("output", "json")

	resp, err := http.Get(fmt.Sprintf("%s?%s", apiURL, params.Encode()))
	if err != nil {
		return 0, 0, err
	}
	defer resp.Body.Close()

	var result AmapGeoCodeResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, 0, err
	}

	if result.Status != "1" || len(result.Geocodes) == 0 {
		return 0, 0, fmt.Errorf("地址解析失败: %s", result.Info)
	}

	// 解析经纬度
	location := result.Geocodes[0].Location
	_, err = fmt.Sscanf(location, "%f,%f", &lng, &lat)
	if err != nil {
		return 0, 0, err
	}

	return lng, lat, nil
}

// RoutePlan 路径规划（简化版）
func (g *GeoUtils) RoutePlan(startAddr, endAddr string) ([]map[string]interface{}, error) {
	// 实际场景调用高德路径规划API
	startLng, startLat, err := g.GetCoordinates(startAddr)
	if err != nil {
		return nil, err
	}

	endLng, endLat, err := g.GetCoordinates(endAddr)
	if err != nil {
		return nil, err
	}

	// 模拟路线
	return []map[string]interface{}{
		{
			"address":   startAddr,
			"longitude": startLng,
			"latitude":  startLat,
			"type":      "start",
		},
		{
			"address":   fmt.Sprintf("%s中转站", startAddr[:2]),
			"longitude": (startLng + endLng) / 2,
			"latitude":  (startLat + endLat) / 2,
			"type":      "transit",
		},
		{
			"address":   endAddr,
			"longitude": endLng,
			"latitude":  endLat,
			"type":      "end",
		},
	}, nil
}
