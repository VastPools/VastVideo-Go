package components

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// HandleTypeMappingManageAPI 处理类型映射管理API
func (tmm *TypeMappingManager) HandleTypeMappingManageAPI(w http.ResponseWriter, r *http.Request) {
	// 设置CORS头
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// 处理OPTIONS请求
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// 解析URL路径
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 3 {
		http.Error(w, "Invalid API path", http.StatusBadRequest)
		return
	}

	// 路由到不同的处理函数
	switch r.Method {
	case "GET":
		tmm.handleGetMapping(w, r, pathParts)
	case "POST":
		tmm.handleCreateMapping(w, r, pathParts)
	case "PUT":
		tmm.handleUpdateMapping(w, r, pathParts)
	case "DELETE":
		tmm.handleDeleteMapping(w, r, pathParts)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleGetMapping 处理GET请求
func (tmm *TypeMappingManager) handleGetMapping(w http.ResponseWriter, r *http.Request, pathParts []string) {
	// 检查路径长度
	if len(pathParts) < 3 {
		http.Error(w, "Invalid API path", http.StatusBadRequest)
		return
	}

	// 根据路径获取特定信息
	switch pathParts[2] {
	case "global_types":
		response := map[string]interface{}{
			"success": true,
			"data":    tmm.GetGlobalTypes(),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)

	case "source_mappings":
		response := map[string]interface{}{
			"success": true,
			"data":    tmm.GetSourceMappings(),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)

	case "source":
		if len(pathParts) < 5 {
			http.Error(w, "Missing source code", http.StatusBadRequest)
			return
		}
		sourceCode := pathParts[4]
		if mapping, found := tmm.GetSourceMapping(sourceCode); found {
			response := map[string]interface{}{
				"success": true,
				"data":    mapping,
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		} else {
			response := map[string]interface{}{
				"success": false,
				"message": "Source not found",
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(response)
		}

	default:
		http.Error(w, "Invalid endpoint", http.StatusBadRequest)
	}
}

// handleCreateMapping 处理POST请求
func (tmm *TypeMappingManager) handleCreateMapping(w http.ResponseWriter, r *http.Request, pathParts []string) {
	if len(pathParts) < 3 {
		http.Error(w, "Invalid API path", http.StatusBadRequest)
		return
	}

	// 处理 auto_fetch 特殊情况
	if len(pathParts) == 3 && pathParts[2] == "auto_fetch" {
		tmm.handleAutoFetchSourceTypes(w, r)
		return
	}

	// 处理 init_all_sources 特殊情况
	if len(pathParts) == 3 && pathParts[2] == "init_all_sources" {
		tmm.handleInitAllSources(w, r)
		return
	}

	// 其他情况需要至少4个路径段
	if len(pathParts) < 4 {
		http.Error(w, "Missing endpoint", http.StatusBadRequest)
		return
	}

	switch pathParts[3] {
	case "global_type":
		tmm.handleCreateGlobalType(w, r)
	case "source_mapping":
		tmm.handleCreateSourceMapping(w, r)
	default:
		http.Error(w, "Invalid endpoint", http.StatusBadRequest)
	}
}

// handleUpdateMapping 处理PUT请求
func (tmm *TypeMappingManager) handleUpdateMapping(w http.ResponseWriter, r *http.Request, pathParts []string) {
	if len(pathParts) < 4 {
		http.Error(w, "Missing endpoint", http.StatusBadRequest)
		return
	}

	switch pathParts[3] {
	case "global_type":
		tmm.handleUpdateGlobalType(w, r)
	case "source_mapping":
		tmm.handleUpdateSourceMapping(w, r)
	default:
		http.Error(w, "Invalid endpoint", http.StatusBadRequest)
	}
}

// handleDeleteMapping 处理DELETE请求
func (tmm *TypeMappingManager) handleDeleteMapping(w http.ResponseWriter, r *http.Request, pathParts []string) {
	if len(pathParts) < 4 {
		http.Error(w, "Missing endpoint", http.StatusBadRequest)
		return
	}

	switch pathParts[3] {
	case "global_type":
		tmm.handleDeleteGlobalType(w, r)
	case "source_mapping":
		tmm.handleDeleteSourceMapping(w, r)
	default:
		http.Error(w, "Invalid endpoint", http.StatusBadRequest)
	}
}

// handleCreateGlobalType 创建全局类型
func (tmm *TypeMappingManager) handleCreateGlobalType(w http.ResponseWriter, r *http.Request) {
	var globalType GlobalType
	if err := json.NewDecoder(r.Body).Decode(&globalType); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// 验证必需字段
	if globalType.ID == "" || globalType.Name == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	// 检查是否已存在
	globalTypes := tmm.GetGlobalTypes()
	if _, exists := globalTypes[globalType.ID]; exists {
		response := map[string]interface{}{
			"success": false,
			"message": "Global type already exists",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(response)
		return
	}

	// 创建新配置
	newConfig := *tmm.config
	if newConfig.GlobalTypes == nil {
		newConfig.GlobalTypes = make(map[string]GlobalType)
	}
	newConfig.GlobalTypes[globalType.ID] = globalType

	// 更新配置
	if err := tmm.UpdateConfig(&newConfig); err != nil {
		response := map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("Failed to update config: %v", err),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"data":    globalType,
		"message": "Global type created successfully",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleUpdateGlobalType 更新全局类型
func (tmm *TypeMappingManager) handleUpdateGlobalType(w http.ResponseWriter, r *http.Request) {
	var globalType GlobalType
	if err := json.NewDecoder(r.Body).Decode(&globalType); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// 验证必需字段
	if globalType.ID == "" || globalType.Name == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	// 检查是否存在
	globalTypes := tmm.GetGlobalTypes()
	if _, exists := globalTypes[globalType.ID]; !exists {
		response := map[string]interface{}{
			"success": false,
			"message": "Global type not found",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(response)
		return
	}

	// 更新配置
	newConfig := *tmm.config
	newConfig.GlobalTypes[globalType.ID] = globalType

	// 更新配置
	if err := tmm.UpdateConfig(&newConfig); err != nil {
		response := map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("Failed to update config: %v", err),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"data":    globalType,
		"message": "Global type updated successfully",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleDeleteGlobalType 删除全局类型
func (tmm *TypeMappingManager) handleDeleteGlobalType(w http.ResponseWriter, r *http.Request) {
	// 从查询参数获取ID
	typeID := r.URL.Query().Get("id")
	if typeID == "" {
		http.Error(w, "Missing type ID", http.StatusBadRequest)
		return
	}

	// 检查是否存在
	globalTypes := tmm.GetGlobalTypes()
	if _, exists := globalTypes[typeID]; !exists {
		response := map[string]interface{}{
			"success": false,
			"message": "Global type not found",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(response)
		return
	}

	// 检查是否被使用
	sourceMappings := tmm.GetSourceMappings()
	for _, mapping := range sourceMappings {
		for _, sourceType := range mapping.TypeList {
			if sourceType.GlobalType == typeID {
				response := map[string]interface{}{
					"success": false,
					"message": "Cannot delete global type that is being used by source mappings",
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusConflict)
				json.NewEncoder(w).Encode(response)
				return
			}
		}
	}

	// 删除配置
	newConfig := *tmm.config
	delete(newConfig.GlobalTypes, typeID)

	// 更新配置
	if err := tmm.UpdateConfig(&newConfig); err != nil {
		response := map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("Failed to update config: %v", err),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Global type deleted successfully",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleCreateSourceMapping 创建源映射
func (tmm *TypeMappingManager) handleCreateSourceMapping(w http.ResponseWriter, r *http.Request) {
	var sourceMapping SourceMapping
	if err := json.NewDecoder(r.Body).Decode(&sourceMapping); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// 从查询参数获取源代码
	sourceCode := r.URL.Query().Get("source_code")
	if sourceCode == "" {
		http.Error(w, "Missing source code", http.StatusBadRequest)
		return
	}

	// 检查是否已存在
	sourceMappings := tmm.GetSourceMappings()
	if _, exists := sourceMappings[sourceCode]; exists {
		response := map[string]interface{}{
			"success": false,
			"message": "Source mapping already exists",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(response)
		return
	}

	// 验证全局类型
	globalTypes := tmm.GetGlobalTypes()
	for _, sourceType := range sourceMapping.TypeList {
		if sourceType.GlobalType != "" {
			if _, exists := globalTypes[sourceType.GlobalType]; !exists {
				response := map[string]interface{}{
					"success": false,
					"message": fmt.Sprintf("Global type '%s' not found", sourceType.GlobalType),
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(response)
				return
			}
		}
	}

	// 创建新配置
	newConfig := *tmm.config
	if newConfig.SourceMappings == nil {
		newConfig.SourceMappings = make(map[string]SourceMapping)
	}
	newConfig.SourceMappings[sourceCode] = sourceMapping

	// 更新配置
	if err := tmm.UpdateConfig(&newConfig); err != nil {
		response := map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("Failed to update config: %v", err),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"data":    sourceMapping,
		"message": "Source mapping created successfully",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleUpdateSourceMapping 更新源映射
func (tmm *TypeMappingManager) handleUpdateSourceMapping(w http.ResponseWriter, r *http.Request) {
	var sourceMapping SourceMapping
	if err := json.NewDecoder(r.Body).Decode(&sourceMapping); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// 从查询参数获取源代码
	sourceCode := r.URL.Query().Get("source_code")
	if sourceCode == "" {
		http.Error(w, "Missing source code", http.StatusBadRequest)
		return
	}

	// 检查是否存在
	sourceMappings := tmm.GetSourceMappings()
	if _, exists := sourceMappings[sourceCode]; !exists {
		response := map[string]interface{}{
			"success": false,
			"message": "Source mapping not found",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(response)
		return
	}

	// 验证全局类型
	globalTypes := tmm.GetGlobalTypes()
	for _, sourceType := range sourceMapping.TypeList {
		if sourceType.GlobalType != "" {
			if _, exists := globalTypes[sourceType.GlobalType]; !exists {
				response := map[string]interface{}{
					"success": false,
					"message": fmt.Sprintf("Global type '%s' not found", sourceType.GlobalType),
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(response)
				return
			}
		}
	}

	// 更新配置
	newConfig := *tmm.config
	newConfig.SourceMappings[sourceCode] = sourceMapping

	// 更新配置
	if err := tmm.UpdateConfig(&newConfig); err != nil {
		response := map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("Failed to update config: %v", err),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"data":    sourceMapping,
		"message": "Source mapping updated successfully",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleDeleteSourceMapping 删除源映射
func (tmm *TypeMappingManager) handleDeleteSourceMapping(w http.ResponseWriter, r *http.Request) {
	// 从查询参数获取源代码
	sourceCode := r.URL.Query().Get("source_code")
	if sourceCode == "" {
		http.Error(w, "Missing source code", http.StatusBadRequest)
		return
	}

	// 检查是否存在
	sourceMappings := tmm.GetSourceMappings()
	if _, exists := sourceMappings[sourceCode]; !exists {
		response := map[string]interface{}{
			"success": false,
			"message": "Source mapping not found",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(response)
		return
	}

	// 删除配置
	newConfig := *tmm.config
	delete(newConfig.SourceMappings, sourceCode)

	// 更新配置
	if err := tmm.UpdateConfig(&newConfig); err != nil {
		response := map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("Failed to update config: %v", err),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Source mapping deleted successfully",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleAutoFetchSourceTypes 自动获取源类型
func (tmm *TypeMappingManager) handleAutoFetchSourceTypes(w http.ResponseWriter, r *http.Request) {
	// 只允许POST请求
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 获取查询参数
	sourceCode := r.URL.Query().Get("source_code")
	sourceURL := r.URL.Query().Get("source_url")

	if sourceCode == "" || sourceURL == "" {
		response := map[string]interface{}{
			"success": false,
			"message": "Missing source_code or source_url",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	// 自动更新映射
	err := tmm.AutoUpdateMapping(sourceCode, sourceURL)
	if err != nil {
		response := map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("Failed to auto update mapping: %v", err),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	// 获取更新后的映射信息
	mapping, found := tmm.GetSourceMapping(sourceCode)
	if !found {
		response := map[string]interface{}{
			"success": false,
			"message": "Failed to get updated mapping",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	// 统计映射结果
	mappedCount := 0
	for _, sourceType := range mapping.TypeList {
		if sourceType.GlobalType != "" {
			mappedCount++
		}
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Auto fetch source types successfully",
		"data": map[string]interface{}{
			"source_code":  sourceCode,
			"type_count":   len(mapping.TypeList),
			"mapped_count": mappedCount,
			"mapping":      mapping,
		},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleInitAllSources 批量初始化所有源的类型映射
func (tmm *TypeMappingManager) handleInitAllSources(w http.ResponseWriter, r *http.Request) {
	// 只允许POST请求
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 获取配置文件路径
	configPath := r.URL.Query().Get("config_path")
	if configPath == "" {
		configPath = "config/config.ini" // 默认路径
	}

	// 执行批量初始化
	err := tmm.InitializeAllSourcesFromConfig(configPath)
	if err != nil {
		response := map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("Failed to initialize all sources: %v", err),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	// 获取初始化结果
	sourceMappings := tmm.GetSourceMappings()
	totalSources := len(sourceMappings)
	totalTypes := 0
	totalMapped := 0

	for _, mapping := range sourceMappings {
		totalTypes += len(mapping.TypeList)
		for _, sourceType := range mapping.TypeList {
			if sourceType.GlobalType != "" {
				totalMapped++
			}
		}
	}

	response := map[string]interface{}{
		"success": true,
		"message": "All sources initialized successfully",
		"data": map[string]interface{}{
			"total_sources":   totalSources,
			"total_types":     totalTypes,
			"total_mapped":    totalMapped,
			"source_mappings": sourceMappings,
		},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
