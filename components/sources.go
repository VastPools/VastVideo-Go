package components

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"vastproxy-go/utils"
)

// VideoSource 视频源结构
type VideoSource struct {
	Code      string `json:"code"`
	Name      string `json:"name"`
	URL       string `json:"url"`
	IsDefault bool   `json:"is_default"`
	Enabled   bool   `json:"enabled"`
}

// VideoItem 视频项目结构
type VideoItem struct {
	VodName     string `json:"vod_name"`
	VodPic      string `json:"vod_pic"`
	VodYear     string `json:"vod_year"`
	TypeName    string `json:"type_name"`
	VodScore    string `json:"vod_score"`
	VodContent  string `json:"vod_content"`
	VodActor    string `json:"vod_actor"`
	VodDirector string `json:"vod_director"`
	VodArea     string `json:"vod_area"`
	VodLang     string `json:"vod_lang"`
	VodTime     string `json:"vod_time"`
	VodRemarks  string `json:"vod_remarks"`
	VodPlayUrl  string `json:"vod_play_url"`
}

// SearchResponse 搜索响应结构
type SearchResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    []VideoItem `json:"data"`
	Count   int         `json:"count"`
	// 原始API响应数据
	RawData map[string]interface{} `json:"raw_data,omitempty"`
}

// VideoItemWithGlobalType 带全局类型的视频项目结构
type VideoItemWithGlobalType struct {
	VideoItem
	GlobalType     string `json:"global_type"`
	GlobalTypeName string `json:"global_type_name"`
}

// SourcesConfig 视频源配置管理器
type SourcesConfig struct {
	sources            []VideoSource
	typeMappingManager *TypeMappingManager
}

// NewSourcesConfig 创建新的视频源配置管理器
func NewSourcesConfig() *SourcesConfig {
	return &SourcesConfig{
		sources:            []VideoSource{},
		typeMappingManager: nil,
	}
}

// SetTypeMappingManager 设置类型映射管理器
func (sc *SourcesConfig) SetTypeMappingManager(tmm *TypeMappingManager) {
	sc.typeMappingManager = tmm
}

// LoadFromConfigFile 从配置文件加载视频源
func (sc *SourcesConfig) LoadFromConfigFile(configData []byte) error {
	sc.sources = []VideoSource{}

	// 解析JSON配置文件
	var config struct {
		Sources map[string]struct {
			Name      string `json:"name"`
			URL       string `json:"url"`
			IsDefault bool   `json:"is_default"`
			Enabled   bool   `json:"enabled"`
		} `json:"sources"`
	}

	if err := json.Unmarshal(configData, &config); err != nil {
		return fmt.Errorf("解析JSON配置文件失败: %v", err)
	}

	if config.Sources == nil {
		return fmt.Errorf("配置文件中未找到 sources 部分")
	}

	// 从JSON配置构建VideoSource对象
	for code, sourceConfig := range config.Sources {
		source := VideoSource{
			Code:      code,
			Name:      sourceConfig.Name,
			URL:       sourceConfig.URL,
			IsDefault: sourceConfig.IsDefault,
			Enabled:   sourceConfig.Enabled,
		}

		sc.sources = append(sc.sources, source)
	}

	return nil
}

// GetSources 获取所有视频源
func (sc *SourcesConfig) GetSources() []VideoSource {
	return sc.sources
}

// GetSourceByCode 根据代码获取视频源
func (sc *SourcesConfig) GetSourceByCode(code string) *VideoSource {
	for _, source := range sc.sources {
		if source.Code == code {
			return &source
		}
	}
	return nil
}

// HandleSourcesAPI 处理 /api/sources 接口
func (sc *SourcesConfig) HandleSourcesAPI(w http.ResponseWriter, r *http.Request) {
	// 设置CORS头
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// 处理OPTIONS请求
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// 只允许GET请求
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 返回JSON格式的视频源列表
	response := map[string]interface{}{
		"success": true,
		"data":    sc.sources,
		"count":   len(sc.sources),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
	log.Printf("✅ /api/sources 请求 [IP:%s]", utils.GetRequestIP(r))
}

// HandleSourceSearchAPI 处理 /api/source_search 接口
func (sc *SourcesConfig) HandleSourceSearchAPI(w http.ResponseWriter, r *http.Request) {
	// 设置CORS头
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// 处理OPTIONS请求
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// 只允许GET请求
	if r.Method != "GET" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Method not allowed",
			"data":    []VideoItem{},
		})
		return
	}

	// 获取查询参数
	sourceCode := r.URL.Query().Get("source")
	keyword := r.URL.Query().Get("keyword")
	page := r.URL.Query().Get("page")
	typeId := r.URL.Query().Get("t") // 添加类型ID参数
	isLatest := r.URL.Query().Get("latest") == "true"

	if sourceCode == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Missing source parameter",
			"data":    []VideoItem{},
		})
		return
	}

	// 如果不是获取最新推荐，则keyword是必需的（除非指定了类型ID）
	if !isLatest && keyword == "" && typeId == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Missing keyword parameter or type ID",
			"data":    []VideoItem{},
		})
		return
	}

	// 获取指定的视频源
	source := sc.GetSourceByCode(sourceCode)
	if source == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Source not found",
			"data":    []VideoItem{},
		})
		return
	}

	// 执行搜索
	rawResults, err := sc.searchSource(source, keyword, page, typeId)
	if err != nil {
		log.Printf("❌ 搜索失败: %v [IP:%s]", err, utils.GetRequestIP(r))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Search failed: " + err.Error(),
			"data":    []VideoItem{},
		})
		return
	}

	// 从原始数据中提取视频列表
	var videos []VideoItem
	if list, ok := rawResults["list"].([]interface{}); ok {
		log.Printf("✅ 找到list字段，包含 %d 个视频", len(list))
		for _, item := range list {
			if videoMap, ok := item.(map[string]interface{}); ok {
				typeName := getString(videoMap, "type_name")

				video := VideoItem{
					VodName:     getString(videoMap, "vod_name"),
					VodPic:      getString(videoMap, "vod_pic"),
					VodYear:     getString(videoMap, "vod_year"),
					TypeName:    typeName,
					VodScore:    getString(videoMap, "vod_score"),
					VodContent:  getString(videoMap, "vod_content"),
					VodActor:    getString(videoMap, "vod_actor"),
					VodDirector: getString(videoMap, "vod_director"),
					VodArea:     getString(videoMap, "vod_area"),
					VodLang:     getString(videoMap, "vod_lang"),
					VodTime:     getString(videoMap, "vod_time"),
					VodRemarks:  getString(videoMap, "vod_remarks"),
					VodPlayUrl:  getString(videoMap, "vod_play_url"),
				}
				videos = append(videos, video)
			}
		}
	} else {
		log.Printf("⚠️ 未找到list字段，尝试其他字段名")
		// 尝试其他可能的字段名
		for _, fieldName := range []string{"data", "videos", "results", "items"} {
			if list, ok := rawResults[fieldName].([]interface{}); ok {
				log.Printf("✅ 找到%s字段，包含 %d 个视频", fieldName, len(list))
				for _, item := range list {
					if videoMap, ok := item.(map[string]interface{}); ok {
						typeName := getString(videoMap, "type_name")

						video := VideoItem{
							VodName:     getString(videoMap, "vod_name"),
							VodPic:      getString(videoMap, "vod_pic"),
							VodYear:     getString(videoMap, "vod_year"),
							TypeName:    typeName,
							VodScore:    getString(videoMap, "vod_score"),
							VodContent:  getString(videoMap, "vod_content"),
							VodActor:    getString(videoMap, "vod_actor"),
							VodDirector: getString(videoMap, "vod_director"),
							VodArea:     getString(videoMap, "vod_area"),
							VodLang:     getString(videoMap, "vod_lang"),
							VodTime:     getString(videoMap, "vod_time"),
							VodRemarks:  getString(videoMap, "vod_remarks"),
							VodPlayUrl:  getString(videoMap, "vod_play_url"),
						}
						videos = append(videos, video)
					}
				}
				break
			}
		}
	}

	// 构建响应
	response := SearchResponse{
		Success: true,
		Message: "搜索成功",
		Data:    videos,
		Count:   len(videos),
		RawData: rawResults, // 包含原始API响应数据
	}

	// 设置响应头
	w.Header().Set("Content-Type", "application/json")

	// 编码并发送响应
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("❌ 编码响应失败: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	log.Printf("✅ /api/source_search 请求 [IP:%s]", utils.GetRequestIP(r))
}

// HandleScorpioSourcesAPI 处理 /api/scorpio_sources 接口，返回 scorpio.json 中的全部内容
func HandleScorpioSourcesAPI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	f, err := os.Open("config/scorpio.json")
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "无法读取scorpio.json: " + err.Error(),
			"data":    nil,
		})
		return
	}
	defer f.Close()
	var sources []map[string]interface{}
	dec := json.NewDecoder(f)
	if err := dec.Decode(&sources); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "scorpio.json 解析失败: " + err.Error(),
			"data":    nil,
		})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    sources,
		"count":   len(sources),
	})
}

// searchSource 搜索指定源
func (sc *SourcesConfig) searchSource(source *VideoSource, keyword, page, typeId string) (map[string]interface{}, error) {
	// 构建请求URL
	baseURL := source.URL
	if !strings.HasSuffix(baseURL, "/") {
		baseURL += "/"
	}

	// 构建查询参数
	params := url.Values{}

	// 统一使用 videolist 接口
	params.Set("ac", "videolist")

	// 判断是搜索、获取最新推荐还是按类型筛选
	if keyword == "" && typeId == "" {
		// 获取最新推荐 - 使用默认参数
		params.Set("pg", "1") // 第一页
	} else if typeId != "" {
		// 按类型筛选
		params.Set("t", typeId)
	} else {
		// 搜索 - 添加关键词
		params.Set("wd", keyword)
	}

	if page != "" {
		params.Set("pg", page)
	}

	requestURL := baseURL + "?" + params.Encode()

	// 创建HTTP客户端
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// 创建请求
	req, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	// 设置请求头
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
	req.Header.Set("Cache-Control", "no-cache")

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP错误: %d", resp.StatusCode)
	}

	// 读取响应内容
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	// 解析JSON响应
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("解析JSON失败: %v", err)
	}

	// 添加调试日志
	log.Printf("🔍 API响应状态: 成功")
	if typeId != "" {
		log.Printf("🔍 类型筛选: t=%s", typeId)
	}

	// 记录原始响应数据
	log.Printf("🔍 原始API响应包含字段: %v", getMapKeys(result))

	// 记录分页信息
	if total, ok := result["total"].(float64); ok {
		log.Printf("📊 总数据量: %.0f", total)
	}
	if pageCount, ok := result["pagecount"].(float64); ok {
		log.Printf("📄 总页数: %.0f", pageCount)
	}
	if currentPage, ok := result["page"].(float64); ok {
		log.Printf("📖 当前页: %.0f", currentPage)
	}

	return result, nil
}

// getString 安全地从map中获取字符串值
func getString(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

// getMapKeys 获取map的所有键
func getMapKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
