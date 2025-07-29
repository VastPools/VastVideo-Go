package components

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"vastproxy-go/utils"
)

// SourcesManageHandler 处理视频源管理相关的API请求
type SourcesManageHandler struct {
	sourcesConfig *SourcesConfig
	configPath    string
}

// NewSourcesManageHandler 创建新的视频源管理处理器
func NewSourcesManageHandler(sourcesConfig *SourcesConfig, configPath string) *SourcesManageHandler {
	return &SourcesManageHandler{
		sourcesConfig: sourcesConfig,
		configPath:    configPath,
	}
}

// HandleSourcesManageAPI 处理视频源管理API请求
func (smh *SourcesManageHandler) HandleSourcesManageAPI(w http.ResponseWriter, r *http.Request) {
	// 设置CORS头
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// 处理OPTIONS请求
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// 解析路径参数
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	log.Printf("🔍 路径解析: %v, 长度: %d, 方法: %s", pathParts, len(pathParts), r.Method)
	log.Printf("🔍 原始路径: %s", r.URL.Path)

	// 检查基本路径格式
	if len(pathParts) < 2 || pathParts[0] != "api" || pathParts[1] != "sources_manage" {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	// 根据路径和方法分发到不同的处理函数
	switch {
	case r.Method == "GET" && len(pathParts) == 2:
		// GET /api/sources_manage - 获取所有视频源
		smh.handleGetSources(w, r)
	case r.Method == "POST" && len(pathParts) == 2:
		// POST /api/sources_manage - 添加新视频源
		smh.handleAddSource(w, r)
	case r.Method == "GET" && len(pathParts) == 3 && pathParts[2] == "test_remote":
		// GET /api/sources_manage/test_remote - 测试远程URL
		smh.handleTestRemoteURL(w, r)
	case r.Method == "GET" && len(pathParts) == 3 && pathParts[2] == "types":
		// GET /api/sources_manage/types - 获取指定源的类型列表
		smh.handleGetSourceTypes(w, r)
	case r.Method == "POST" && len(pathParts) == 3 && pathParts[2] == "upload":
		// POST /api/sources_manage/upload - 上传配置文件
		smh.handleUploadConfig(w, r)
	case r.Method == "POST" && len(pathParts) == 3 && pathParts[2] == "update_from_url":
		// POST /api/sources_manage/update_from_url - 从远程URL更新配置
		smh.handleUpdateFromURL(w, r)
	case r.Method == "GET" && len(pathParts) == 3 && pathParts[2] != "test_remote":
		// GET /api/sources_manage/{code} - 获取指定视频源
		smh.handleGetSource(w, r, pathParts[2])
	case r.Method == "PUT" && len(pathParts) == 3 && pathParts[2] != "upload" && pathParts[2] != "update_from_url":
		// PUT /api/sources_manage/{code} - 更新指定视频源
		smh.handleUpdateSource(w, r, pathParts[2])
	case r.Method == "DELETE" && len(pathParts) == 3 && pathParts[2] != "upload" && pathParts[2] != "update_from_url":
		// DELETE /api/sources_manage/{code} - 删除指定视频源
		smh.handleDeleteSource(w, r, pathParts[2])
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleGetSources 获取所有视频源
func (smh *SourcesManageHandler) handleGetSources(w http.ResponseWriter, r *http.Request) {
	sources := smh.sourcesConfig.GetSources()

	response := map[string]interface{}{
		"success": true,
		"data":    sources,
		"count":   len(sources),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
	log.Printf("✅ /api/sources_manage GET 请求 [IP:%s]", utils.GetRequestIP(r))
}

// handleAddSource 添加新视频源
func (smh *SourcesManageHandler) handleAddSource(w http.ResponseWriter, r *http.Request) {
	var sourceData map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&sourceData); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// 验证必需字段
	code, ok := sourceData["code"].(string)
	if !ok || code == "" {
		http.Error(w, "Missing or invalid code field", http.StatusBadRequest)
		return
	}

	name, ok := sourceData["name"].(string)
	if !ok || name == "" {
		http.Error(w, "Missing or invalid name field", http.StatusBadRequest)
		return
	}

	url, ok := sourceData["url"].(string)
	if !ok || url == "" {
		http.Error(w, "Missing or invalid url field", http.StatusBadRequest)
		return
	}

	// 检查源代码是否已存在
	existingSource := smh.sourcesConfig.GetSourceByCode(code)
	if existingSource != nil {
		http.Error(w, "Source code already exists", http.StatusConflict)
		return
	}

	// 创建新视频源
	newSource := VideoSource{
		Code:      code,
		Name:      name,
		URL:       url,
		IsDefault: getBoolValue(sourceData, "is_default"),
		Enabled:   getBoolValue(sourceData, "enabled"),
	}

	// 保存到配置文件
	if err := smh.saveSourceToConfig(newSource); err != nil {
		log.Printf("❌ 保存视频源配置失败: %v", err)
		http.Error(w, "Failed to save configuration", http.StatusInternalServerError)
		return
	}

	// 重新加载配置
	smh.reloadConfig()

	response := map[string]interface{}{
		"success": true,
		"message": "视频源添加成功",
		"data":    newSource,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
	log.Printf("✅ /api/sources_manage POST 添加视频源: %s [IP:%s]", code, utils.GetRequestIP(r))
}

// handleGetSource 获取指定视频源
func (smh *SourcesManageHandler) handleGetSource(w http.ResponseWriter, r *http.Request, code string) {
	source := smh.sourcesConfig.GetSourceByCode(code)
	if source == nil {
		http.Error(w, "Source not found", http.StatusNotFound)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"data":    source,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleUpdateSource 更新指定视频源
func (smh *SourcesManageHandler) handleUpdateSource(w http.ResponseWriter, r *http.Request, code string) {
	var sourceData map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&sourceData); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	log.Printf("🔍 更新源请求: code=%s, data=%+v", code, sourceData)

	// 检查视频源是否存在
	existingSource := smh.sourcesConfig.GetSourceByCode(code)
	if existingSource == nil {
		log.Printf("❌ 源未找到: code=%s", code)
		log.Printf("🔍 当前所有源: %+v", smh.sourcesConfig.GetSources())
		http.Error(w, "Source not found", http.StatusNotFound)
		return
	}

	log.Printf("✅ 找到源: %+v", existingSource)

	// 更新字段
	if name, ok := sourceData["name"].(string); ok && name != "" {
		existingSource.Name = name
	}
	if url, ok := sourceData["url"].(string); ok && url != "" {
		existingSource.URL = url
	}
	if _, exists := sourceData["enabled"]; exists {
		existingSource.Enabled = getBoolValue(sourceData, "enabled")
	}
	if _, exists := sourceData["is_default"]; exists {
		existingSource.IsDefault = getBoolValue(sourceData, "is_default")
	}

	// 保存到配置文件
	if err := smh.saveSourceToConfig(*existingSource); err != nil {
		log.Printf("❌ 更新视频源配置失败: %v", err)
		http.Error(w, "Failed to save configuration", http.StatusInternalServerError)
		return
	}

	// 重新加载配置
	smh.reloadConfig()

	response := map[string]interface{}{
		"success": true,
		"message": "视频源更新成功",
		"data":    existingSource,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
	log.Printf("✅ /api/sources_manage PUT 更新视频源: %s [IP:%s]", code, utils.GetRequestIP(r))
}

// handleDeleteSource 删除指定视频源
func (smh *SourcesManageHandler) handleDeleteSource(w http.ResponseWriter, r *http.Request, code string) {
	// 检查视频源是否存在
	existingSource := smh.sourcesConfig.GetSourceByCode(code)
	if existingSource == nil {
		http.Error(w, "Source not found", http.StatusNotFound)
		return
	}

	// 从配置文件中删除
	if err := smh.removeSourceFromConfig(code); err != nil {
		log.Printf("❌ 删除视频源配置失败: %v", err)
		http.Error(w, "Failed to delete configuration", http.StatusInternalServerError)
		return
	}

	// 重新加载配置
	smh.reloadConfig()

	response := map[string]interface{}{
		"success": true,
		"message": "视频源删除成功",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
	log.Printf("✅ /api/sources_manage DELETE 删除视频源: %s [IP:%s]", code, utils.GetRequestIP(r))
}

// handleTestRemoteURL 测试远程URL
func (smh *SourcesManageHandler) handleTestRemoteURL(w http.ResponseWriter, r *http.Request) {
	url := r.URL.Query().Get("url")
	if url == "" {
		http.Error(w, "Missing url parameter", http.StatusBadRequest)
		return
	}

	// 测试远程URL
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		response := map[string]interface{}{
			"success": false,
			"message": "连接失败: " + err.Error(),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		response := map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("HTTP错误: %d", resp.StatusCode),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	// 尝试解析JSON
	var config interface{}
	if err := json.NewDecoder(resp.Body).Decode(&config); err != nil {
		response := map[string]interface{}{
			"success": false,
			"message": "无效的JSON格式: " + err.Error(),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "远程URL测试成功",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleUploadConfig 上传配置文件
func (smh *SourcesManageHandler) handleUploadConfig(w http.ResponseWriter, r *http.Request) {
	// 解析multipart表单
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Failed to get uploaded file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// 检查文件类型
	if !strings.HasSuffix(header.Filename, ".json") {
		http.Error(w, "Only JSON files are allowed", http.StatusBadRequest)
		return
	}

	// 读取文件内容
	content, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Failed to read file", http.StatusInternalServerError)
		return
	}

	// 验证JSON格式
	var config map[string]interface{}
	if err := json.Unmarshal(content, &config); err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	// 备份原配置文件
	backupPath := smh.configPath + ".backup." + fmt.Sprintf("%d", time.Now().Unix())
	if err := copyFile(smh.configPath, backupPath); err != nil {
		log.Printf("⚠️ 备份配置文件失败: %v", err)
	}

	// 写入新配置文件
	if err := os.WriteFile(smh.configPath, content, 0644); err != nil {
		http.Error(w, "Failed to write configuration file", http.StatusInternalServerError)
		return
	}

	// 重新加载配置
	smh.reloadConfig()

	response := map[string]interface{}{
		"success": true,
		"message": "配置文件上传成功",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
	log.Printf("✅ /api/sources_manage/upload 配置文件上传成功 [IP:%s]", utils.GetRequestIP(r))
}

// handleUpdateFromURL 从远程URL更新配置
func (smh *SourcesManageHandler) handleUpdateFromURL(w http.ResponseWriter, r *http.Request) {
	var requestData map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	url, ok := requestData["url"].(string)
	if !ok || url == "" {
		http.Error(w, "Missing url field", http.StatusBadRequest)
		return
	}

	// 下载配置文件
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		http.Error(w, "Failed to download configuration: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		http.Error(w, fmt.Sprintf("HTTP error: %d", resp.StatusCode), http.StatusInternalServerError)
		return
	}

	// 读取内容
	content, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Failed to read response", http.StatusInternalServerError)
		return
	}

	// 验证JSON格式
	var config map[string]interface{}
	if err := json.Unmarshal(content, &config); err != nil {
		http.Error(w, "Invalid JSON format in remote file", http.StatusBadRequest)
		return
	}

	// 备份原配置文件
	backupPath := smh.configPath + ".backup." + fmt.Sprintf("%d", time.Now().Unix())
	if err := copyFile(smh.configPath, backupPath); err != nil {
		log.Printf("⚠️ 备份配置文件失败: %v", err)
	}

	// 写入新配置文件
	if err := os.WriteFile(smh.configPath, content, 0644); err != nil {
		http.Error(w, "Failed to write configuration file", http.StatusInternalServerError)
		return
	}

	// 重新加载配置
	smh.reloadConfig()

	response := map[string]interface{}{
		"success": true,
		"message": "远程配置更新成功",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
	log.Printf("✅ /api/sources_manage/update_from_url 远程配置更新成功 [IP:%s]", utils.GetRequestIP(r))
}

// saveSourceToConfig 保存视频源到配置文件
func (smh *SourcesManageHandler) saveSourceToConfig(source VideoSource) error {
	// 读取现有配置
	configData, err := os.ReadFile(smh.configPath)
	if err != nil {
		return fmt.Errorf("读取配置文件失败: %v", err)
	}

	var config map[string]interface{}
	if err := json.Unmarshal(configData, &config); err != nil {
		return fmt.Errorf("解析配置文件失败: %v", err)
	}

	// 确保sources字段存在
	if config["sources"] == nil {
		config["sources"] = make(map[string]interface{})
	}

	sources := config["sources"].(map[string]interface{})

	// 添加或更新视频源
	sources[source.Code] = map[string]interface{}{
		"name":       source.Name,
		"url":        source.URL,
		"enabled":    source.Enabled,
		"is_default": source.IsDefault,
	}

	// 更新最后修改时间
	config["last_updated"] = time.Now().Format("2006-01-02")

	// 写回配置文件
	newConfigData, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化配置失败: %v", err)
	}

	return os.WriteFile(smh.configPath, newConfigData, 0644)
}

// removeSourceFromConfig 从配置文件中删除视频源
func (smh *SourcesManageHandler) removeSourceFromConfig(code string) error {
	// 读取现有配置
	configData, err := os.ReadFile(smh.configPath)
	if err != nil {
		return fmt.Errorf("读取配置文件失败: %v", err)
	}

	var config map[string]interface{}
	if err := json.Unmarshal(configData, &config); err != nil {
		return fmt.Errorf("解析配置文件失败: %v", err)
	}

	sources := config["sources"].(map[string]interface{})
	delete(sources, code)

	// 更新最后修改时间
	config["last_updated"] = time.Now().Format("2006-01-02")

	// 写回配置文件
	newConfigData, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化配置失败: %v", err)
	}

	return os.WriteFile(smh.configPath, newConfigData, 0644)
}

// reloadConfig 重新加载配置
func (smh *SourcesManageHandler) reloadConfig() {
	log.Printf("🔄 开始重新加载配置...")

	configData, err := os.ReadFile(smh.configPath)
	if err != nil {
		log.Printf("❌ 重新加载配置文件失败: %v", err)
		return
	}

	// 记录重新加载前的源数量
	beforeCount := len(smh.sourcesConfig.GetSources())
	log.Printf("📊 重新加载前源数量: %d", beforeCount)

	if err := smh.sourcesConfig.LoadFromConfigFile(configData); err != nil {
		log.Printf("❌ 重新加载视频源配置失败: %v", err)
		return
	}

	// 记录重新加载后的源数量
	afterCount := len(smh.sourcesConfig.GetSources())
	log.Printf("📊 重新加载后源数量: %d", afterCount)

	// 记录所有源的信息
	sources := smh.sourcesConfig.GetSources()
	log.Printf("📋 重新加载后的源列表:")
	for _, source := range sources {
		log.Printf("   - %s: %s (启用: %v, 默认: %v)", source.Code, source.Name, source.Enabled, source.IsDefault)
	}

	log.Printf("✅ 视频源配置重新加载成功")
}

// 辅助函数

// getBoolValue 从map中获取布尔值
func getBoolValue(data map[string]interface{}, key string) bool {
	if value, exists := data[key]; exists {
		if boolValue, ok := value.(bool); ok {
			return boolValue
		}
	}
	return false
}

// copyFile 复制文件
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

// handleGetSourceTypes 获取指定源的类型列表
func (smh *SourcesManageHandler) handleGetSourceTypes(w http.ResponseWriter, r *http.Request) {
	// 获取源代码参数
	sourceCode := r.URL.Query().Get("source")
	if sourceCode == "" {
		http.Error(w, "Missing source parameter", http.StatusBadRequest)
		return
	}

	// 查找指定的源
	source := smh.sourcesConfig.GetSourceByCode(sourceCode)
	if source == nil {
		http.Error(w, "Source not found", http.StatusNotFound)
		return
	}

	// 构建API URL（不使用ac参数）
	apiURL := source.URL
	if !strings.HasSuffix(apiURL, "/") {
		apiURL += "/"
	}
	apiURL += "api.php/provide/vod"

	log.Printf("🔍 获取类型列表，源: %s, URL: %s", sourceCode, apiURL)

	// 发送HTTP请求获取类型列表
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(apiURL)
	if err != nil {
		log.Printf("❌ 请求类型列表失败: %v", err)
		http.Error(w, fmt.Sprintf("Failed to fetch types: %v", err), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// 读取响应内容
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("❌ 读取响应失败: %v", err)
		http.Error(w, fmt.Sprintf("Failed to read response: %v", err), http.StatusInternalServerError)
		return
	}

	// 解析JSON响应
	var apiResponse map[string]interface{}
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		log.Printf("❌ 解析JSON失败: %v", err)
		http.Error(w, fmt.Sprintf("Failed to parse JSON: %v", err), http.StatusInternalServerError)
		return
	}

	// 提取类型列表
	classList, ok := apiResponse["class"].([]interface{})
	if !ok {
		log.Printf("❌ 未找到类型列表")
		http.Error(w, "No class list found in response", http.StatusInternalServerError)
		return
	}

	// 构建响应
	response := map[string]interface{}{
		"success": true,
		"source":  sourceCode,
		"data":    classList,
		"count":   len(classList),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
	log.Printf("✅ 获取类型列表成功，源: %s, 类型数量: %d", sourceCode, len(classList))
}
