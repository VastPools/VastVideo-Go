package components

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// GlobalType 全局类型定义
type GlobalType struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Priority    int    `json:"priority"`
	Enabled     bool   `json:"enabled"`
}

// SourceType 源类型定义
type SourceType struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	GlobalType string `json:"global_type"`
}

// SourceMapping 源映射配置
type SourceMapping struct {
	Name     string       `json:"name"`
	Enabled  bool         `json:"enabled"`
	TypeList []SourceType `json:"type_list"`
}

// TypeMappingConfig 类型映射配置
type TypeMappingConfig struct {
	Version        string                   `json:"version"`
	Description    string                   `json:"description"`
	LastUpdated    string                   `json:"last_updated"`
	GlobalTypes    map[string]GlobalType    `json:"global_types"`
	SourceMappings map[string]SourceMapping `json:"source_mappings"`
}

// TypeMappingManager 类型映射管理器
type TypeMappingManager struct {
	config     *TypeMappingConfig
	configPath string
	mutex      sync.RWMutex

	// 内存中的映射缓存
	// 格式: map[sourceCode]map[sourceTypeID]globalTypeCode
	typeCache map[string]map[int]string

	// 反向映射缓存
	// 格式: map[sourceCode]map[globalTypeCode][]int
	reverseCache map[string]map[string][]int

	// 类型名称缓存
	// 格式: map[sourceCode]map[sourceTypeID]sourceTypeName
	nameCache map[string]map[int]string
}

// NewTypeMappingManager 创建新的类型映射管理器
func NewTypeMappingManager(configPath string) *TypeMappingManager {
	return &TypeMappingManager{
		configPath:   configPath,
		typeCache:    make(map[string]map[int]string),
		reverseCache: make(map[string]map[string][]int),
		nameCache:    make(map[string]map[int]string),
	}
}

// LoadConfig 加载配置文件
func (tmm *TypeMappingManager) LoadConfig() error {
	tmm.mutex.Lock()
	defer tmm.mutex.Unlock()

	// 读取配置文件
	data, err := os.ReadFile(tmm.configPath)
	if err != nil {
		return fmt.Errorf("读取类型映射配置文件失败: %v", err)
	}

	// 解析JSON配置
	var config TypeMappingConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("解析类型映射配置文件失败: %v", err)
	}

	tmm.config = &config

	// 构建内存映射缓存
	tmm.buildCache()

	log.Printf("✅ 类型映射配置加载成功，共 %d 个全局类型，%d 个源映射",
		len(config.GlobalTypes), len(config.SourceMappings))

	return nil
}

// buildCache 构建内存映射缓存
func (tmm *TypeMappingManager) buildCache() {
	tmm.typeCache = make(map[string]map[int]string)
	tmm.reverseCache = make(map[string]map[string][]int)
	tmm.nameCache = make(map[string]map[int]string)

	for sourceCode, sourceMapping := range tmm.config.SourceMappings {
		// 初始化缓存
		tmm.typeCache[sourceCode] = make(map[int]string)
		tmm.reverseCache[sourceCode] = make(map[string][]int)
		tmm.nameCache[sourceCode] = make(map[int]string)

		// 构建映射缓存
		for _, sourceType := range sourceMapping.TypeList {
			// 正向映射: sourceTypeID -> globalTypeCode
			tmm.typeCache[sourceCode][sourceType.ID] = sourceType.GlobalType

			// 类型名称缓存
			tmm.nameCache[sourceCode][sourceType.ID] = sourceType.Name

			// 反向映射: globalTypeCode -> []sourceTypeID
			if tmm.reverseCache[sourceCode][sourceType.GlobalType] == nil {
				tmm.reverseCache[sourceCode][sourceType.GlobalType] = make([]int, 0)
			}
			tmm.reverseCache[sourceCode][sourceType.GlobalType] = append(
				tmm.reverseCache[sourceCode][sourceType.GlobalType], sourceType.ID)
		}
	}
}

// GetGlobalType 根据源类型ID获取全局类型
func (tmm *TypeMappingManager) GetGlobalType(sourceCode string, sourceTypeID int) (string, bool) {
	tmm.mutex.RLock()
	defer tmm.mutex.RUnlock()

	if sourceCache, exists := tmm.typeCache[sourceCode]; exists {
		if globalType, found := sourceCache[sourceTypeID]; found {
			return globalType, true
		}
	}
	return "", false
}

// GetSourceTypeIDs 根据全局类型获取源类型ID列表
func (tmm *TypeMappingManager) GetSourceTypeIDs(sourceCode, globalTypeCode string) ([]int, bool) {
	tmm.mutex.RLock()
	defer tmm.mutex.RUnlock()

	if sourceCache, exists := tmm.reverseCache[sourceCode]; exists {
		if sourceTypeIDs, found := sourceCache[globalTypeCode]; found {
			return sourceTypeIDs, true
		}
	}
	return nil, false
}

// GetSourceTypeName 根据源类型ID获取类型名称
func (tmm *TypeMappingManager) GetSourceTypeName(sourceCode string, sourceTypeID int) (string, bool) {
	tmm.mutex.RLock()
	defer tmm.mutex.RUnlock()

	if sourceCache, exists := tmm.nameCache[sourceCode]; exists {
		if typeName, found := sourceCache[sourceTypeID]; found {
			return typeName, true
		}
	}
	return "", false
}

// GetGlobalTypes 获取所有全局类型
func (tmm *TypeMappingManager) GetGlobalTypes() map[string]GlobalType {
	tmm.mutex.RLock()
	defer tmm.mutex.RUnlock()

	if tmm.config == nil {
		return nil
	}
	return tmm.config.GlobalTypes
}

// GetSourceMappings 获取所有源映射
func (tmm *TypeMappingManager) GetSourceMappings() map[string]SourceMapping {
	tmm.mutex.RLock()
	defer tmm.mutex.RUnlock()

	if tmm.config == nil {
		return nil
	}
	return tmm.config.SourceMappings
}

// GetSourceMapping 获取指定源的映射
func (tmm *TypeMappingManager) GetSourceMapping(sourceCode string) (*SourceMapping, bool) {
	tmm.mutex.RLock()
	defer tmm.mutex.RUnlock()

	if tmm.config == nil {
		return nil, false
	}

	if mapping, exists := tmm.config.SourceMappings[sourceCode]; exists {
		return &mapping, true
	}
	return nil, false
}

// UpdateConfig 更新配置文件
func (tmm *TypeMappingManager) UpdateConfig(newConfig *TypeMappingConfig) error {
	tmm.mutex.Lock()
	defer tmm.mutex.Unlock()

	// 更新配置
	tmm.config = newConfig
	tmm.config.LastUpdated = time.Now().Format("2006-01-02T15:04:05Z")

	// 重新构建缓存
	tmm.buildCache()

	// 保存到文件
	return tmm.saveConfig()
}

// saveConfig 保存配置到文件
func (tmm *TypeMappingManager) saveConfig() error {
	if tmm.config == nil {
		return fmt.Errorf("配置为空")
	}

	data, err := json.MarshalIndent(tmm.config, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化配置失败: %v", err)
	}

	if err := os.WriteFile(tmm.configPath, data, 0644); err != nil {
		return fmt.Errorf("保存配置文件失败: %v", err)
	}

	return nil
}

// FetchSourceTypes 从源API获取类型列表
func (tmm *TypeMappingManager) FetchSourceTypes(sourceCode, sourceURL string) ([]SourceType, error) {
	// 构建请求URL - 直接使用源URL获取类型列表
	requestURL := strings.TrimSuffix(sourceURL, "/")

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

	// 提取类型列表
	var sourceTypes []SourceType

	// 从class字段获取类型列表
	if class, ok := result["class"].([]interface{}); ok {
		for _, item := range class {
			if typeMap, ok := item.(map[string]interface{}); ok {
				typeID := getInt(typeMap, "type_id")
				typeName := getStringFromTypeMapping(typeMap, "type_name")
				typePID := getInt(typeMap, "type_pid")

				if typeID > 0 && typeName != "" {
					sourceTypes = append(sourceTypes, SourceType{
						ID:         typeID,
						Name:       typeName,
						GlobalType: "", // 需要后续映射
					})

					// 记录层级关系（可选）
					if typePID > 0 {
						log.Printf("📋 类型层级: %s (ID:%d) -> 父类型ID:%d", typeName, typeID, typePID)
					}
				}
			}
		}
	}

	if len(sourceTypes) == 0 {
		return nil, fmt.Errorf("未找到类型数据，响应结构: %v", getMapKeysFromTypeMapping(result))
	}

	return sourceTypes, nil
}

// AutoUpdateMapping 自动更新源的类型映射
func (tmm *TypeMappingManager) AutoUpdateMapping(sourceCode, sourceURL string) error {
	// 获取源的类型列表
	sourceTypes, err := tmm.FetchSourceTypes(sourceCode, sourceURL)
	if err != nil {
		return fmt.Errorf("获取源类型失败: %v", err)
	}

	log.Printf("📋 源 %s 发现 %d 个类型", sourceCode, len(sourceTypes))

	// 获取当前映射
	currentMapping, exists := tmm.GetSourceMapping(sourceCode)
	if !exists {
		// 创建新的映射
		currentMapping = &SourceMapping{
			Name:     sourceCode,
			Enabled:  true,
			TypeList: make([]SourceType, 0),
		}
	}

	// 创建新的映射
	newTypeList := make([]SourceType, 0)

	// 复制现有的映射
	for _, existingType := range currentMapping.TypeList {
		newTypeList = append(newTypeList, existingType)
	}

	// 为未映射的类型创建映射
	unmappedTypes := make([]SourceType, 0)
	for _, sourceType := range sourceTypes {
		mapped := false
		for _, existingType := range newTypeList {
			if existingType.ID == sourceType.ID {
				mapped = true
				break
			}
		}
		if !mapped {
			unmappedTypes = append(unmappedTypes, sourceType)
		}
	}

	// 智能映射未映射的类型
	if len(unmappedTypes) > 0 {
		log.Printf("🔍 发现 %d 个未映射的类型，尝试智能映射...", len(unmappedTypes))

		for i := range unmappedTypes {
			globalType := tmm.smartMapType(unmappedTypes[i].Name)
			if globalType != "" {
				unmappedTypes[i].GlobalType = globalType
				log.Printf("🔗 智能映射: %s (ID:%d) -> %s",
					unmappedTypes[i].Name, unmappedTypes[i].ID, globalType)
			} else {
				log.Printf("⚠️ 无法映射类型: %s (ID:%d)",
					unmappedTypes[i].Name, unmappedTypes[i].ID)
			}
		}
	}

	// 添加新映射的类型
	newTypeList = append(newTypeList, unmappedTypes...)

	// 更新配置
	newMapping := SourceMapping{
		Name:     currentMapping.Name,
		Enabled:  currentMapping.Enabled,
		TypeList: newTypeList,
	}

	// 创建新的配置
	newConfig := *tmm.config
	newConfig.SourceMappings[sourceCode] = newMapping

	return tmm.UpdateConfig(&newConfig)
}

// smartMapType 智能映射类型
func (tmm *TypeMappingManager) smartMapType(sourceTypeName string) string {
	sourceTypeName = strings.ToLower(sourceTypeName)

	// 电影类型映射
	if strings.Contains(sourceTypeName, "电影") || strings.Contains(sourceTypeName, "动作") ||
		strings.Contains(sourceTypeName, "喜剧") || strings.Contains(sourceTypeName, "爱情") ||
		strings.Contains(sourceTypeName, "科幻") || strings.Contains(sourceTypeName, "恐怖") ||
		strings.Contains(sourceTypeName, "剧情") || strings.Contains(sourceTypeName, "战争") ||
		strings.Contains(sourceTypeName, "理论") || strings.Contains(sourceTypeName, "动画片") {
		return "movie"
	}

	// 电视剧类型映射
	if strings.Contains(sourceTypeName, "连续剧") || strings.Contains(sourceTypeName, "国产剧") ||
		strings.Contains(sourceTypeName, "香港剧") || strings.Contains(sourceTypeName, "韩国剧") ||
		strings.Contains(sourceTypeName, "台湾剧") || strings.Contains(sourceTypeName, "日本剧") ||
		strings.Contains(sourceTypeName, "海外剧") || strings.Contains(sourceTypeName, "泰国剧") ||
		strings.Contains(sourceTypeName, "欧美剧") || strings.Contains(sourceTypeName, "短剧") {
		return "tv"
	}

	// 综艺类型映射
	if strings.Contains(sourceTypeName, "综艺") || strings.Contains(sourceTypeName, "娱乐") {
		return "variety"
	}

	// 动漫类型映射
	if strings.Contains(sourceTypeName, "动漫") || strings.Contains(sourceTypeName, "动画") ||
		strings.Contains(sourceTypeName, "日韩动漫") || strings.Contains(sourceTypeName, "欧美动漫") ||
		strings.Contains(sourceTypeName, "港台动漫") || strings.Contains(sourceTypeName, "海外动漫") {
		return "anime"
	}

	// 纪录片类型映射
	if strings.Contains(sourceTypeName, "纪录片") || strings.Contains(sourceTypeName, "记录片") {
		return "documentary"
	}

	// 体育类型映射
	if strings.Contains(sourceTypeName, "体育") || strings.Contains(sourceTypeName, "足球") ||
		strings.Contains(sourceTypeName, "篮球") || strings.Contains(sourceTypeName, "网球") ||
		strings.Contains(sourceTypeName, "斯诺克") {
		return "sport"
	}

	// 成人类型映射
	if strings.Contains(sourceTypeName, "福利") || strings.Contains(sourceTypeName, "伦理") {
		return "adult"
	}

	return ""
}

// getStringFromTypeMapping 安全地从map中获取字符串值
func getStringFromTypeMapping(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

// getInt 安全地从map中获取整数值
func getInt(m map[string]interface{}, key string) int {
	if val, ok := m[key]; ok {
		switch v := val.(type) {
		case int:
			return v
		case float64:
			return int(v)
		case string:
			if i, err := strconv.Atoi(v); err == nil {
				return i
			}
		}
	}
	return 0
}

// getMapKeysFromTypeMapping 获取map的所有键
func getMapKeysFromTypeMapping(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// HandleTypeMappingAPI 处理类型映射API
func (tmm *TypeMappingManager) HandleTypeMappingAPI(w http.ResponseWriter, r *http.Request) {
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
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 获取查询参数
	sourceCode := r.URL.Query().Get("source")
	globalType := r.URL.Query().Get("global_type")
	sourceTypeIDStr := r.URL.Query().Get("source_type_id")

	var response interface{}

	if sourceCode != "" {
		if globalType != "" {
			// 获取指定源的指定全局类型对应的源类型ID列表
			if sourceTypeIDs, found := tmm.GetSourceTypeIDs(sourceCode, globalType); found {
				// 构建包含名称的响应
				typeInfoList := make([]map[string]interface{}, 0)
				for _, typeID := range sourceTypeIDs {
					if typeName, found := tmm.GetSourceTypeName(sourceCode, typeID); found {
						typeInfoList = append(typeInfoList, map[string]interface{}{
							"id":   typeID,
							"name": typeName,
						})
					}
				}

				response = map[string]interface{}{
					"success": true,
					"data": map[string]interface{}{
						"source_code":  sourceCode,
						"global_type":  globalType,
						"source_types": typeInfoList,
					},
				}
			} else {
				response = map[string]interface{}{
					"success": false,
					"message": "未找到映射关系",
				}
			}
		} else if sourceTypeIDStr != "" {
			// 获取指定源的指定源类型ID对应的全局类型
			if sourceTypeID, err := strconv.Atoi(sourceTypeIDStr); err == nil {
				if globalTypeCode, found := tmm.GetGlobalType(sourceCode, sourceTypeID); found {
					if typeName, found := tmm.GetSourceTypeName(sourceCode, sourceTypeID); found {
						response = map[string]interface{}{
							"success": true,
							"data": map[string]interface{}{
								"source_code":      sourceCode,
								"source_type_id":   sourceTypeID,
								"source_type_name": typeName,
								"global_type":      globalTypeCode,
							},
						}
					} else {
						response = map[string]interface{}{
							"success": false,
							"message": "未找到类型名称",
						}
					}
				} else {
					response = map[string]interface{}{
						"success": false,
						"message": "未找到映射关系",
					}
				}
			} else {
				response = map[string]interface{}{
					"success": false,
					"message": "无效的源类型ID",
				}
			}
		} else {
			// 获取指定源的所有映射
			if mapping, found := tmm.GetSourceMapping(sourceCode); found {
				response = map[string]interface{}{
					"success": true,
					"data":    mapping,
				}
			} else {
				response = map[string]interface{}{
					"success": false,
					"message": "源不存在",
				}
			}
		}
	} else {
		// 获取所有配置信息
		response = map[string]interface{}{
			"success": true,
			"data": map[string]interface{}{
				"global_types":    tmm.GetGlobalTypes(),
				"source_mappings": tmm.GetSourceMappings(),
			},
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// InitializeAllSourcesFromConfig 从sources.json初始化所有源的类型映射
func (tmm *TypeMappingManager) InitializeAllSourcesFromConfig(configPath string) error {
	// 读取sources.json文件
	configData, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("读取配置文件失败: %v", err)
	}

	// 解析JSON格式
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

	log.Printf("📋 从配置文件发现 %d 个源", len(config.Sources))

	// 为每个源初始化类型映射
	for sourceCode, sourceConfig := range config.Sources {
		// 只处理启用的源
		if !sourceConfig.Enabled {
			log.Printf("ℹ️ 源 %s 已禁用，跳过", sourceCode)
			continue
		}

		log.Printf("🔄 正在初始化源: %s (%s)", sourceCode, sourceConfig.Name)

		// 检查是否已经存在映射
		if _, exists := tmm.config.SourceMappings[sourceCode]; exists {
			log.Printf("ℹ️ 源 %s 已存在映射，跳过", sourceCode)
			continue
		}

		// 自动更新映射
		err := tmm.AutoUpdateMapping(sourceCode, sourceConfig.URL)
		if err != nil {
			log.Printf("❌ 源 %s 初始化失败: %v", sourceCode, err)
			continue
		}

		// 更新源名称
		tmm.mutex.Lock()
		if mapping, exists := tmm.config.SourceMappings[sourceCode]; exists {
			mapping.Name = sourceConfig.Name
			tmm.config.SourceMappings[sourceCode] = mapping
		}
		tmm.mutex.Unlock()

		log.Printf("✅ 源 %s 初始化成功", sourceCode)
	}

	// 保存所有更改
	return tmm.saveConfig()
}
