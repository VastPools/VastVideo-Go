# 类型映射管理功能使用说明

## 概述

类型映射管理功能允许你统一管理不同视频源的视频类型，将各个源的本地类型映射到全局统一类型，便于前端展示和用户理解。

## 功能特点

- 🎯 **基于ID的映射**：使用源类型ID而不是字符串进行映射
- 🔄 **智能自动映射**：自动识别和映射常见类型
- 📱 **响应式管理界面**：支持移动端和桌面端
- ⚡ **高性能缓存**：内存中的高效映射缓存
- 🔧 **完整API接口**：提供查询和管理两套API

## 快速开始

### 1. 启动程序

```bash
./vastvideo-go
```

程序启动后，你会看到类似以下的日志：

```
🎯 类型映射管理: http://0.0.0.0:8228/type_mapping
```

### 2. 访问管理界面

在浏览器中打开：`http://localhost:8228/type_mapping`

## 管理界面功能

### 全局类型管理

#### 查看全局类型列表
- 在"全局类型管理"标签页中查看所有已定义的全局类型
- 每个类型显示：ID、名称、描述、优先级、启用状态

#### 添加全局类型
1. 点击"添加类型"标签页
2. 填写以下信息：
   - **类型ID**：唯一标识符（如：movie、tv、anime）
   - **类型名称**：显示名称（如：电影、电视剧、动漫）
   - **描述**：类型说明（可选）
   - **优先级**：排序优先级（数字越小优先级越高）
   - **启用状态**：是否启用该类型

#### 编辑全局类型
1. 在类型列表中点击"编辑"按钮
2. 修改相关信息
3. 点击"保存"按钮

#### 删除全局类型
1. 在类型列表中点击"删除"按钮
2. 确认删除操作

**注意**：如果全局类型正在被源映射使用，则无法删除。

### 源映射管理

#### 查看源映射列表
- 在"源映射管理"标签页中查看所有源的映射配置
- 每个映射显示：源名称、源代码、类型数量、启用状态

#### 添加源映射
1. 点击"添加映射"标签页
2. 填写以下信息：
   - **源代码**：源的唯一标识符（如：bfzy、dyttzy）
   - **源名称**：源的显示名称（如：暴风资源、电影天堂资源）
   - **启用状态**：是否启用该源映射
   - **类型映射列表**：添加该源的各个类型映射

#### 类型映射配置
对于每个源类型，需要配置：
- **类型ID**：源中的类型ID（如：1、2、3）
- **类型名称**：源中的类型名称（如：电影、电视剧）
- **全局类型**：映射到的全局类型（从下拉列表选择）

#### 编辑源映射
1. 在映射列表中点击"编辑"按钮
2. 修改相关信息
3. 点击"保存"按钮

#### 删除源映射
1. 在映射列表中点击"删除"按钮
2. 确认删除操作

## API接口

### 查询API

#### 获取所有映射信息
```
GET /api/type_mapping
```

#### 根据源类型ID获取全局类型
```
GET /api/type_mapping?source=bfzy&source_type_id=1
```

#### 根据全局类型获取源类型列表
```
GET /api/type_mapping?source=bfzy&global_type=movie
```

### 管理API

#### 获取全局类型列表
```
GET /api/type_mapping_manage/global_types
```

#### 创建全局类型
```
POST /api/type_mapping_manage/global_type
Content-Type: application/json

{
  "id": "new_type",
  "name": "新类型",
  "description": "新类型描述",
  "priority": 8,
  "enabled": true
}
```

#### 更新全局类型
```
PUT /api/type_mapping_manage/global_type
Content-Type: application/json

{
  "id": "movie",
  "name": "电影",
  "description": "电影类型",
  "priority": 1,
  "enabled": true
}
```

#### 删除全局类型
```
DELETE /api/type_mapping_manage/global_type?id=movie
```

#### 获取源映射列表
```
GET /api/type_mapping_manage/source_mappings
```

#### 创建源映射
```
POST /api/type_mapping_manage/source_mapping?source_code=new_source
Content-Type: application/json

{
  "name": "新源",
  "enabled": true,
  "type_list": [
    {
      "id": 1,
      "name": "电影",
      "global_type": "movie"
    }
  ]
}
```

#### 更新源映射
```
PUT /api/type_mapping_manage/source_mapping?source_code=bfzy
Content-Type: application/json

{
  "name": "暴风资源",
  "enabled": true,
  "type_list": [
    {
      "id": 1,
      "name": "电影",
      "global_type": "movie"
    }
  ]
}
```

#### 删除源映射
```
DELETE /api/type_mapping_manage/source_mapping?source_code=bfzy
```

## 配置文件

类型映射配置保存在 `config/type_mapping.json` 文件中，格式如下：

```json
{
  "version": "1.0.0",
  "description": "视频类型映射配置",
  "last_updated": "2024-01-01T00:00:00Z",
  "global_types": {
    "movie": {
      "id": "movie",
      "name": "电影",
      "description": "电影类型",
      "priority": 1,
      "enabled": true
    }
  },
  "source_mappings": {
    "bfzy": {
      "name": "暴风资源",
      "enabled": true,
      "type_list": [
        {
          "id": 1,
          "name": "电影",
          "global_type": "movie"
        }
      ]
    }
  }
}
```

## 使用示例

### 1. 获取源的类型列表
```javascript
fetch('/api/type_mapping?source=bfzy')
  .then(response => response.json())
  .then(data => {
    console.log('源类型列表:', data.data.type_list);
  });
```

### 2. 根据类型ID获取全局类型
```javascript
fetch('/api/type_mapping?source=bfzy&source_type_id=1')
  .then(response => response.json())
  .then(data => {
    if (data.success) {
      console.log('全局类型:', data.data.global_type);
      console.log('类型名称:', data.data.source_type_name);
    }
  });
```

### 3. 创建新的全局类型
```javascript
fetch('/api/type_mapping_manage/global_type', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({
    id: 'documentary',
    name: '纪录片',
    description: '纪录片类型',
    priority: 5,
    enabled: true
  })
})
.then(response => response.json())
.then(data => {
  console.log('创建结果:', data);
});
```

## 注意事项

1. **全局类型ID唯一性**：全局类型的ID必须唯一，不能重复
2. **源类型ID唯一性**：源类型ID在同一个源内必须唯一
3. **依赖关系**：删除全局类型前需要确保没有源映射在使用它
4. **配置持久化**：所有配置更改会立即保存到文件
5. **缓存机制**：配置加载到内存中，查询速度快
6. **CORS支持**：所有API都支持跨域请求

## 故障排除

### 常见问题

1. **配置文件不存在**
   - 程序会自动创建默认配置文件
   - 检查 `config/type_mapping.json` 文件是否存在

2. **API请求失败**
   - 检查程序是否正常运行
   - 确认端口号是否正确（默认8228）
   - 检查网络连接

3. **类型映射不生效**
   - 确认源映射已启用
   - 检查类型ID是否正确
   - 验证全局类型是否存在且已启用

### 日志查看

程序运行时会输出详细的日志信息，包括：
- 配置加载状态
- API请求处理
- 错误信息

查看日志可以帮助诊断问题。

## 扩展功能

### 自动映射更新

程序支持从源API自动获取类型列表并更新映射：

```go
err := typeMappingManager.AutoUpdateMapping("bfzy", "https://bfzyapi.com/api.php/provide/vod")
```

### 智能类型识别

系统内置智能类型识别功能，可以自动将常见的中文类型名称映射到全局类型。

### 批量操作

通过API可以实现批量操作，如批量更新多个源的映射配置。 