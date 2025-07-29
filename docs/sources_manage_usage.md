# 视频源管理功能使用说明

## 概述

VastVideo-Go 提供了完整的视频源管理功能，支持通过Web界面管理 `sources.json` 配置文件，包括增删改查、单点测试、文件上传和远程URL更新等功能。

## 功能特性

### 🎯 核心功能
- **增删改查**: 完整的视频源CRUD操作
- **单点测试**: 实时测试视频源API可用性
- **文件上传**: 支持JSON配置文件上传更新
- **远程更新**: 支持从远程URL自动更新配置
- **实时预览**: 即时查看配置变更效果
- **备份恢复**: 自动备份原配置文件

### 🛡️ 安全特性
- **配置验证**: 自动验证JSON格式和必需字段
- **备份机制**: 操作前自动备份原配置
- **错误处理**: 完善的错误提示和回滚机制
- **权限控制**: 支持密码保护（可选）

## 访问方式

### 主管理界面
```
http://localhost:8228/sources_manage
```

### 测试界面
```
http://localhost:8228/test_sources_manage
```

## API接口说明

### 1. 获取视频源列表
```http
GET /api/sources_manage
```

**响应示例**:
```json
{
  "success": true,
  "data": [
    {
      "code": "bfzy",
      "name": "暴风资源",
      "url": "https://bfzyapi.com/api.php/provide/vod",
      "is_default": true,
      "enabled": true
    }
  ],
  "count": 16
}
```

### 2. 添加新视频源
```http
POST /api/sources_manage
Content-Type: application/json

{
  "code": "new_source",
  "name": "新视频源",
  "url": "https://example.com/api.php/provide/vod",
  "enabled": true,
  "is_default": false
}
```

### 3. 更新视频源
```http
PUT /api/sources_manage/{code}
Content-Type: application/json

{
  "name": "更新后的名称",
  "url": "https://new-url.com/api.php/provide/vod",
  "enabled": true,
  "is_default": false
}
```

### 4. 删除视频源
```http
DELETE /api/sources_manage/{code}
```

### 5. 测试视频源
```http
GET /api/source_search?source={code}&latest=true&page=1
```

### 6. 测试远程URL
```http
GET /api/sources_manage/test_remote?url={remote_url}
```

### 7. 上传配置文件
```http
POST /api/sources_manage/upload
Content-Type: multipart/form-data

file: [JSON文件]
```

### 8. 从远程URL更新配置
```http
POST /api/sources_manage/update_from_url
Content-Type: application/json

{
  "url": "https://example.com/sources.json"
}
```

## 配置文件格式

### sources.json 标准格式
```json
{
  "version": "1.0.0",
  "description": "视频源配置文件",
  "last_updated": "2025-01-20",
  "sources": {
    "bfzy": {
      "name": "暴风资源",
      "url": "https://bfzyapi.com/api.php/provide/vod",
      "is_default": true,
      "enabled": true
    },
    "dyttzy": {
      "name": "电影天堂资源",
      "url": "http://caiji.dyttzyapi.com/api.php/provide/vod",
      "is_default": true,
      "enabled": true
    }
  }
}
```

### 字段说明
- `version`: 配置文件版本
- `description`: 配置描述
- `last_updated`: 最后更新时间
- `sources`: 视频源配置对象
  - `{code}`: 视频源代码（唯一标识）
    - `name`: 视频源名称
    - `url`: API接口地址
    - `is_default`: 是否设为默认源
    - `enabled`: 是否启用

## 使用步骤

### 1. 访问管理界面
打开浏览器访问 `http://localhost:8228/sources_manage`

### 2. 查看现有视频源
页面加载后会自动显示所有已配置的视频源列表

### 3. 添加新视频源
1. 点击"添加源"按钮
2. 填写源代码、名称、URL等信息
3. 选择是否启用和设为默认
4. 点击"保存"完成添加

### 4. 编辑视频源
1. 在视频源列表中点击"编辑"按钮
2. 修改相关信息
3. 点击"保存"完成更新

### 5. 测试视频源
1. 在视频源列表中点击"测试"按钮
2. 系统会自动调用API测试可用性
3. 查看测试结果

### 6. 删除视频源
1. 在视频源列表中点击"删除"按钮
2. 确认删除操作
3. 视频源将被从配置中移除

### 7. 上传配置文件
1. 点击"上传配置"按钮
2. 选择JSON格式的配置文件
3. 点击"上传并更新"
4. 系统会自动验证并应用新配置

### 8. 远程URL更新
1. 点击"远程更新"按钮
2. 输入远程配置文件的URL
3. 点击"测试连接"验证URL可用性
4. 点击"更新配置"下载并应用新配置

## 注意事项

### ⚠️ 重要提醒
1. **备份重要**: 系统会自动备份原配置文件，但建议手动备份重要配置
2. **格式验证**: 上传的JSON文件必须符合标准格式
3. **源代码唯一**: 每个视频源的代码必须唯一
4. **URL格式**: API URL必须是有效的HTTP/HTTPS地址
5. **权限设置**: 确保程序有读写配置文件的权限

### 🔧 故障排除
1. **配置文件损坏**: 删除损坏的配置文件，使用备份文件恢复
2. **权限问题**: 检查程序对配置文件的读写权限
3. **网络问题**: 确保远程URL可访问且返回有效JSON
4. **格式错误**: 使用JSON验证工具检查配置文件格式

### 📝 最佳实践
1. **定期备份**: 定期备份 `sources.json` 文件
2. **测试验证**: 添加新视频源后及时测试可用性
3. **版本管理**: 使用版本控制管理配置文件变更
4. **监控日志**: 关注程序日志了解配置变更情况

## 高级功能

### 批量操作
- 支持通过上传配置文件批量更新多个视频源
- 支持从远程URL批量同步配置

### 配置验证
- 自动验证JSON格式
- 检查必需字段完整性
- 验证URL格式和可访问性

### 错误处理
- 详细的错误提示信息
- 自动回滚失败的配置变更
- 保留操作日志便于排查问题

## 技术支持

如遇到问题，请：
1. 查看程序运行日志
2. 检查配置文件格式
3. 验证网络连接状态
4. 联系技术支持团队

---

**VastVideo-Go 视频源管理功能** - 让视频源配置管理变得简单高效 