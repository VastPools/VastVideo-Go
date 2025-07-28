# 视频源配置迁移文档

## 概述

本项目已将视频源配置从 `config.ini` 文件中的 `[sources]` 部分迁移到独立的 `config/sources.json` 文件中，以提供更好的配置管理和扩展性。

## 变更内容

### 1. 新增文件

- **`config/sources.json`**: 新的视频源配置文件
- **`test_sources.html`**: 视频源配置测试页面

### 2. 修改文件

- **`config/config.ini`**: 删除了 `[sources]` 部分
- **`components/sources.go`**: 修改了配置加载逻辑，从JSON格式读取
- **`components/type_mapping.go`**: 更新了批量初始化功能
- **`main.go`**: 更新了配置加载和路由

### 3. 删除依赖

- 移除了对 `gopkg.in/ini.v1` 包的依赖（在sources.go中）

## 配置文件格式

### 新的 JSON 格式

```json
{
  "version": "1.0.0",
  "description": "视频源配置文件",
  "last_updated": "2025-07-28",
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

- **`version`**: 配置文件版本
- **`description`**: 配置文件描述
- **`last_updated`**: 最后更新时间
- **`sources`**: 视频源配置对象
  - **`name`**: 视频源显示名称
  - **`url`**: 视频源API地址
  - **`is_default`**: 是否为默认源
  - **`enabled`**: 是否启用该源

### 旧的 INI 格式（已废弃）

```ini
[sources]
bfzy.name = 暴风资源
bfzy.url = https://bfzyapi.com/api.php/provide/vod
bfzy.is_default = 1

dyttzy.name = 电影天堂资源
dyttzy.url = http://caiji.dyttzyapi.com/api.php/provide/vod
dyttzy.is_default = 1
```

## 功能特性

### 1. 增强的配置管理

- **结构化数据**: JSON格式提供更好的数据结构和验证
- **扩展性**: 易于添加新的配置字段
- **可读性**: 更清晰的配置格式
- **版本控制**: 支持配置文件版本管理

### 2. 源状态管理

- **启用/禁用**: 通过 `enabled` 字段控制源的状态
- **默认源**: 通过 `is_default` 字段标识默认源
- **批量操作**: 支持批量启用/禁用源

### 3. 向后兼容

- 保持了原有的API接口
- 保持了原有的功能特性
- 支持平滑迁移

## API 接口

### 视频源列表

```http
GET /api/sources
```

响应格式：
```json
{
  "success": true,
  "data": [
    {
      "code": "bfzy",
      "name": "暴风资源",
      "url": "https://bfzyapi.com/api.php/provide/vod",
      "is_default": true
    }
  ],
  "count": 16
}
```

### 批量初始化

```http
POST /api/type_mapping_manage/init_all_sources?config_path=config/sources.json
```

## 测试页面

访问 `http://localhost:8228/test_sources` 可以查看视频源配置测试页面，该页面提供：

- 视频源列表显示
- 类型映射状态
- 批量初始化功能
- 统计信息展示

## 迁移步骤

1. **备份配置**: 备份原有的 `config.ini` 文件
2. **创建新配置**: 创建 `config/sources.json` 文件
3. **迁移数据**: 将视频源配置从INI格式转换为JSON格式
4. **测试验证**: 使用测试页面验证配置是否正确
5. **部署更新**: 重新编译并部署应用

## 优势

### 1. 更好的维护性

- JSON格式更易于编辑和验证
- 支持注释和版本信息
- 更好的IDE支持

### 2. 更强的扩展性

- 易于添加新的配置字段
- 支持复杂的配置结构
- 支持配置验证

### 3. 更好的用户体验

- 提供测试页面进行配置验证
- 支持实时配置更新
- 提供详细的统计信息

## 注意事项

1. **配置文件路径**: 确保 `config/sources.json` 文件存在且格式正确
2. **权限设置**: 确保应用有读取配置文件的权限
3. **备份策略**: 建议定期备份配置文件
4. **版本管理**: 建议使用版本控制管理配置文件

## 故障排除

### 常见问题

1. **配置文件不存在**
   - 检查 `config/sources.json` 文件是否存在
   - 确认文件路径正确

2. **JSON格式错误**
   - 使用JSON验证工具检查格式
   - 检查字段名称和类型

3. **源无法加载**
   - 检查 `enabled` 字段是否为 `true`
   - 验证URL格式是否正确

### 调试方法

1. 查看应用日志
2. 使用测试页面验证配置
3. 检查API响应状态
4. 验证文件权限

## 总结

通过将视频源配置迁移到独立的JSON文件，我们实现了：

- 更好的配置管理
- 更强的扩展性
- 更好的用户体验
- 更清晰的代码结构

这一改进为项目的长期维护和发展奠定了良好的基础。 