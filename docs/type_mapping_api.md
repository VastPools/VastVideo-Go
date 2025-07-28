# 类型映射API文档

## 概述

类型映射系统用于统一管理不同视频源的视频类型，将各个源的本地类型映射到全局统一类型，便于前端展示和用户理解。

## 数据结构

### 全局类型 (GlobalType)
```json
{
  "id": "movie",
  "name": "电影",
  "description": "电影类型",
  "priority": 1,
  "enabled": true
}
```

### 源类型 (SourceType)
```json
{
  "id": 1,
  "name": "电影",
  "global_type": "movie"
}
```

### 源映射 (SourceMapping)
```json
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

## API接口

### 1. 查询类型映射

#### 获取所有映射信息
```
GET /api/type_mapping
```

响应示例：
```json
{
  "success": true,
  "data": {
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
        "type_list": [...]
      }
    }
  }
}
```

#### 根据源类型ID获取全局类型
```
GET /api/type_mapping?source=bfzy&source_type_id=1
```

响应示例：
```json
{
  "success": true,
  "data": {
    "source_code": "bfzy",
    "source_type_id": 1,
    "source_type_name": "电影",
    "global_type": "movie"
  }
}
```

#### 根据全局类型获取源类型列表
```
GET /api/type_mapping?source=bfzy&global_type=movie
```

响应示例：
```json
{
  "success": true,
  "data": {
    "source_code": "bfzy",
    "global_type": "movie",
    "source_types": [
      {
        "id": 1,
        "name": "电影"
      }
    ]
  }
}
```

### 2. 类型映射管理API

#### 获取所有映射信息
```
GET /api/type_mapping_manage
```

#### 获取全局类型列表
```
GET /api/type_mapping_manage/global_types
```

#### 获取源映射列表
```
GET /api/type_mapping_manage/source_mappings
```

#### 获取指定源的映射
```
GET /api/type_mapping_manage/source/{source_code}
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

## 使用示例

### 1. 获取源的类型列表
```javascript
// 获取暴风资源的所有类型
fetch('/api/type_mapping?source=bfzy')
  .then(response => response.json())
  .then(data => {
    console.log('源类型列表:', data.data.type_list);
  });
```

### 2. 根据类型ID获取全局类型
```javascript
// 获取类型ID为1对应的全局类型
fetch('/api/type_mapping?source=bfzy&source_type_id=1')
  .then(response => response.json())
  .then(data => {
    if (data.success) {
      console.log('全局类型:', data.data.global_type);
      console.log('类型名称:', data.data.source_type_name);
    }
  });
```

### 3. 根据全局类型获取源类型列表
```javascript
// 获取电影类型对应的所有源类型
fetch('/api/type_mapping?source=bfzy&global_type=movie')
  .then(response => response.json())
  .then(data => {
    if (data.success) {
      console.log('电影类型对应的源类型:', data.data.source_types);
    }
  });
```

### 4. 创建新的全局类型
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

## 错误处理

所有API都会返回统一的错误格式：

```json
{
  "success": false,
  "message": "错误描述"
}
```

常见错误：
- `Missing required fields`: 缺少必需字段
- `Global type not found`: 全局类型不存在
- `Source mapping not found`: 源映射不存在
- `Global type already exists`: 全局类型已存在
- `Source mapping already exists`: 源映射已存在
- `Cannot delete global type that is being used`: 无法删除正在使用的全局类型

## 注意事项

1. 全局类型的ID必须唯一
2. 源类型ID在同一个源内必须唯一
3. 删除全局类型前需要确保没有源映射在使用它
4. 所有API都支持CORS，可以在前端直接调用
5. 配置更改会立即生效并保存到文件 