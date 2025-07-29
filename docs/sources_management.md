# VastVideo-Go 源管理功能

## 🗄️ 概述

VastVideo-Go 统一管理控制台中的源管理模块提供了完整的视频源管理功能，采用现代化的卡片式布局设计，支持增删改查、批量操作和实时测试。

## 🎨 界面设计

### 卡片式布局
- **响应式网格**: 自适应 2-6 列布局，紧凑显示
- **悬停效果**: 卡片悬停时的阴影和变换动画
- **状态标识**: 通过背景色区分不同状态
- **操作按钮**: 测试、编辑、删除的图标按钮
- **双开关控制**: 启用开关(绿色)和默认开关(橙色)分别控制

### 视觉层次
- **默认源**: 金色渐变背景 (yellow-100 to orange-200)
- **启用状态**: 绿色开关按钮控制，实时切换
- **禁用状态**: 灰色渐变背景 (gray-100 to gray-200)
- **普通启用**: 根据源代码自动分配8种不同渐变背景色
- **源代码**: 小字体等宽显示，位于标题下方
- **双开关**: 启用和默认状态分别用不同颜色的开关控制

## 🛠️ 核心功能

### 1. 视频源列表展示
```javascript
// 卡片排序规则
const sortedSources = sources.sort((a, b) => {
    if (a.is_default && !b.is_default) return -1;  // 默认源优先
    if (!a.is_default && b.is_default) return 1;
    return a.name.localeCompare(b.name);           // 按名称排序
});
```

**特性**:
- 默认源自动排在前面
- 支持实时搜索过滤
- 加载状态和空状态处理
- 自动刷新统计数据

### 2. 增删改查操作

#### 添加视频源
- **模态框表单**: 包含所有必要字段
- **实时验证**: 必填字段和URL格式验证
- **状态设置**: 启用/禁用和默认源设置

#### 编辑视频源
- **预填充表单**: 自动加载现有数据
- **代码锁定**: 源代码不可修改
- **状态保持**: 保持原有配置

#### 删除视频源
- **确认对话框**: 防止误操作
- **软删除**: 支持恢复机制
- **关联检查**: 检查依赖关系

#### 查询视频源
- **实时搜索**: 支持代码、名称、URL搜索
- **状态过滤**: 按启用状态过滤
- **排序选项**: 多种排序方式
- **状态切换**: 双开关实时控制启用和默认状态

### 3. 批量操作

#### 文件上传更新
```javascript
// 支持标准的 sources.json 格式
{
    "sources": [
        {
            "code": "bfzy",
            "name": "暴风资源",
            "url": "https://bfzyapi.com/api.php/provide/vod",
            "enabled": true,
            "is_default": false
        }
    ]
}
```

**功能**:
- 拖拽上传支持
- 文件格式验证
- 备份现有配置
- 错误回滚机制

#### 远程URL更新
- **URL验证**: 格式和可访问性检查
- **测试功能**: 预检查远程配置
- **自动下载**: 获取并解析配置
- **错误处理**: 网络异常处理

### 4. 实时测试功能

#### 单源测试
```javascript
// 测试流程
1. 发送测试请求到 /api/source_search
2. 记录响应时间和状态
3. 显示详细的测试结果
4. 提供JSON数据查看
```

**测试内容**:
- API连接性测试
- 响应时间测量
- 数据格式验证
- 错误信息收集

#### 测试结果展示
- **状态指示**: 成功/失败的可视化标识
- **性能指标**: 响应时间显示
- **详细数据**: 完整的JSON响应
- **复制功能**: 一键复制测试结果

## 🔧 API 接口

### 获取视频源列表
```http
GET /api/sources
```

**响应格式**:
```json
{
    "success": true,
    "data": [
        {
            "code": "bfzy",
            "name": "暴风资源",
            "url": "https://bfzyapi.com/api.php/provide/vod",
            "enabled": true,
            "is_default": false
        }
    ]
}
```

### 添加视频源
```http
POST /api/sources_manage
Content-Type: application/json

{
    "code": "new_source",
    "name": "新视频源",
    "url": "https://example.com/api",
    "enabled": true,
    "is_default": false
}
```

### 更新视频源
```http
PUT /api/sources_manage/{code}
Content-Type: application/json

{
    "name": "更新的名称",
    "url": "https://new-url.com/api",
    "enabled": true,
    "is_default": false
}
```

**注意**: 路径格式为 `/api/sources_manage/{code}`，其中 `{code}` 为视频源的唯一标识符。

### 删除视频源
```http
DELETE /api/sources_manage/{code}
```

### 文件上传
```http
POST /api/sources_manage/upload
Content-Type: multipart/form-data

file: sources.json
```

### 远程更新
```http
POST /api/sources_manage/update_from_url
Content-Type: application/json

{
    "url": "https://example.com/sources.json"
}
```

### 测试远程URL
```http
GET /api/sources_manage/test_remote?url={url}
```

## 🎯 用户体验

### 交互设计
- **即时反馈**: 操作结果的Toast通知
- **加载状态**: 优雅的加载动画
- **错误处理**: 友好的错误提示
- **键盘支持**: 快捷键操作

### 响应式设计
- **移动端适配**: 触摸友好的界面
- **平板优化**: 中等屏幕的布局调整
- **桌面端**: 完整功能展示

### 性能优化
- **懒加载**: 按需加载数据
- **缓存机制**: 减少重复请求
- **防抖搜索**: 优化搜索性能
- **异步操作**: 非阻塞的用户界面

## 🔒 安全特性

### 数据验证
- **输入验证**: 前端和后端双重验证
- **URL安全**: 防止恶意URL注入
- **文件类型**: 限制上传文件类型
- **大小限制**: 防止过大文件上传

### 错误处理
- **异常捕获**: 全面的错误处理
- **用户提示**: 清晰的错误信息
- **日志记录**: 操作日志追踪
- **回滚机制**: 失败时的数据恢复

## 📊 监控和统计

### 实时统计
- **源数量**: 总源数和活跃源数
- **状态分布**: 启用/禁用源的比例
- **性能指标**: 平均响应时间
- **错误率**: 测试失败率统计

### 操作日志
- **操作记录**: 所有增删改查操作
- **时间戳**: 精确的操作时间
- **用户标识**: 操作者信息
- **结果状态**: 操作成功/失败状态

## 🚀 未来规划

### 短期目标
- [ ] 批量操作功能
- [ ] 导入/导出功能
- [ ] 源分组管理
- [ ] 性能监控面板

### 长期目标
- [ ] 自动化测试
- [ ] 智能推荐
- [ ] 多用户权限
- [ ] API版本管理

## 📝 使用指南

### 基本操作流程
1. **访问管理页面**: 打开 `/manage` 页面
2. **切换到源管理**: 点击顶部菜单的"源管理"
3. **查看源列表**: 浏览所有视频源卡片
4. **执行操作**: 使用工具栏或卡片操作按钮

### 快捷键
- `Ctrl/Cmd + 2`: 切换到源管理页面
- `Ctrl/Cmd + F`: 聚焦搜索框
- `Enter`: 执行搜索
- `ESC`: 关闭模态框

### 最佳实践
- 定期测试视频源状态
- 备份重要配置
- 使用有意义的源代码
- 及时更新失效的URL

## 🔧 技术修复记录

### API路径解析修复 (2025-07-29)

#### 修复前的问题
- PUT请求 `/api/sources_manage/{code}` 返回 405 Method Not Allowed
- GET请求 `/api/sources_manage` 返回 400 Bad Request  
- 路径解析逻辑存在缺陷，导致路由匹配失败

#### 修复内容
1. **路径解析逻辑优化**
   ```go
   // 修复前：检查 len(pathParts) < 3
   // 修复后：检查基本路径格式
   if len(pathParts) < 2 || pathParts[0] != "api" || pathParts[1] != "sources_manage" {
       http.Error(w, "Invalid path", http.StatusBadRequest)
       return
   }
   ```

2. **路由匹配条件修正**
   ```go
   // 正确的路径匹配逻辑
   case r.Method == "GET" && len(pathParts) == 2:                    // /api/sources_manage
   case r.Method == "PUT" && len(pathParts) == 3:                    // /api/sources_manage/{code}
   case r.Method == "GET" && len(pathParts) == 3 && pathParts[2] == "test_remote": // /api/sources_manage/test_remote
   ```

3. **路由注册顺序优化**
   ```go
   // 先注册带斜杠的路由，再注册不带斜杠的
   http.HandleFunc("/api/sources_manage/", sourcesManageAPIHandler.HandleSourcesManageAPI)
   http.HandleFunc("/api/sources_manage", sourcesManageAPIHandler.HandleSourcesManageAPI)
   ```

4. **禁用状态源删除问题修复**
   ```go
   // 修复前：只加载启用的源
   if !sourceConfig.Enabled {
       continue
   }
   
   // 修复后：加载所有源，包括禁用的
   // 移除了启用状态检查，确保所有源都能被管理
   ```

#### 修复效果
- ✅ PUT请求更新视频源状态正常工作
- ✅ DELETE请求删除视频源正常工作  
- ✅ GET请求获取视频源列表正常工作
- ✅ 所有API路径匹配正确
- ✅ 添加源时正确设置Enabled字段
- ✅ 修复禁用状态源的删除问题

#### 测试验证
使用以下脚本验证所有API功能：
```bash
# 完整功能测试
node test_final_fix.js

# 添加-更新-删除流程测试
node test_add_fix.js

# 删除功能修复测试
node test_delete_fix.js
```

---

**VastVideo-Go 源管理** - 让视频源管理变得简单高效 🚀 