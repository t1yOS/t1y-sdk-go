# t1yOS SDK for Go

[English](./README.md)

[t1yOS](https://www.t1y.net) Serverless 平台 Go SDK — 云数据库、元数据和云函数客户端。

## 安装

```bash
go get github.com/t1yOS/t1y-sdk-go
```

## 快速开始

```go
package main

import (
    "context"
    "fmt"
    "log"

    t1y "github.com/t1yOS/t1y-sdk-go"
)

func main() {
    // 1. 创建客户端
    client, err := t1y.NewClient(&t1y.Config{
        AppID:     1001,                                    // 必填：应用 ID（>= 1001）
        APIKey:    "4fd7448cdc684431a62d8a0111dc6973",     // 必填：32 位 API Key
        SecretKey: "17b784e359c946ffa65eebbf9ce29752",     // 必填：32 位 Secret Key
        // 以下参数可选（均有默认值）：
        // BaseURL: "https://myapp.t1y.net",
        // Version: 0,
        // IsSafeMode: false,
        // TimeFormat: "YYYY-MM-DD HH:mm:ss",
        // Offset: 0,
    })
    if err != nil {
        log.Fatal(err)
    }

    // 2. 初始化（与服务器同步时间偏移和安全模式）
    ctx := context.Background()
    if err := client.Init(ctx); err != nil {
        log.Printf("Warning: %v", err)
    }

    // 3. 开始使用数据库！
    resp, err := client.DB.Collection("users").InsertOne(ctx, map[string]any{
        "name":   "张三",
        "age":    25,
        "active": true,
    })
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(resp.Data.ObjectID)
}
```

## 数据库操作

### 单条操作

```go
db := client.DB.Collection("users")
ctx := context.Background()

// 插入一条
resp, err := db.InsertOne(ctx, map[string]any{"name": "张三", "age": 25})
fmt.Println(resp.Data.ObjectID) // "507f1f77bcf86cd799439011"

// 通过 ObjectID 查询
resp, err := db.FindByID(ctx, "507f1f77bcf86cd799439011")
fmt.Println(resp.Data.Result) // { "_id": "507f1f77...", "name": "张三", ... }

// 通过 ObjectID 更新
_, err = db.UpdateByID(ctx, "507f1f77bcf86cd799439011", map[string]any{
    "$set": map[string]any{"age": 26},
})

// 通过 ObjectID 删除
_, err = db.DeleteByID(ctx, "507f1f77bcf86cd799439011")
```

### 条件操作

```go
// 条件查询一条
resp, err := db.FindOne(ctx, map[string]any{"name": "张三"})

// 条件更新一条
_, err = db.UpdateOne(
    ctx,
    map[string]any{"name": "张三"},                       // 查询条件
    map[string]any{"$set": map[string]any{"age": 27}},    // 更新内容
)

// 条件删除一条
_, err = db.DeleteOne(ctx, map[string]any{"name": "张三"})
```

### 批量操作

```go
// 插入多条
resp, err := db.InsertMany(ctx, []map[string]any{
    {"name": "张三", "age": 25},
    {"name": "李四", "age": 30},
})
fmt.Println(resp.Data.InsertedCount) // 2

// 删除多条
_, err = db.DeleteMany(ctx, map[string]any{"age": map[string]any{"$lt": 18}})

// 更新多条
_, err = db.UpdateMany(
    ctx,
    map[string]any{"status": "inactive"},
    map[string]any{"$set": map[string]any{"status": "archived"}},
)
```

### 高级查询

```go
// 分页查询
resp, err := db.FindSimple(ctx, 1, 20,
    map[string]int{"createdAt": -1},               // 排序（按创建时间倒序）
    map[string]any{"age": map[string]any{"$gte": 18}}, // 查询条件
)
fmt.Println(resp.Data.Results)    // 文档数组
fmt.Println(resp.Data.Pagination) // { totalItems: 42, totalPages: 3 }

// 或使用 FindParams
resp, err = db.Find(ctx, t1y.FindParams{
    Page:   1,
    Size:   20,
    Sort:   map[string]int{"createdAt": -1},
    Filter: map[string]any{"age": map[string]any{"$gte": 18}},
})

// 聚合查询
resp, err := db.Aggregate(ctx, []map[string]any{
    {"$match": map[string]any{"status": "completed"}},
    {"$group": map[string]any{
        "_id": "$category",
        "total": map[string]any{"$sum": "$amount"},
    }},
    {"$sort": map[string]any{"total": -1}},
})

// 计数
countResp, err := db.Count(ctx, map[string]any{"status": "active"})
fmt.Println(countResp.Data.Count)

// 去重查询
distinctResp, err := db.Distinct(ctx, "city", nil)
// 带条件过滤
distinctResp, err = db.Distinct(ctx, "city", map[string]any{"country": "China"})
```

### 表管理

```go
// 获取所有表
resp, err := client.DB.GetCollections(ctx)
fmt.Println(resp.Data.Results) // ["users", "orders", "products"]

// 创建表
_, err = client.DB.Collection("posts").Create(ctx)

// 清空表
clearResp, err := client.DB.Collection("posts").Clear(ctx)
fmt.Println(clearResp.Data.DeletedCount)

// 删除表
_, err = client.DB.Collection("posts").Drop(ctx)
```

## 特殊类型

SDK 提供了一系列辅助函数，用于生成服务端可识别的类型标记：

```go
import t1y "github.com/t1yOS/t1y-sdk-go"

client.DB.Collection("users").InsertOne(ctx, map[string]any{
    // ObjectID 引用
    "userId": t1y.ObjectID("507f1f77bcf86cd799439011"),

    // 日期类型
    "birthday":  t1y.Date("2000-01-01T00:00:00Z"),
    "eventTime": t1y.DateTime("2024-06-15T14:30:00Z"),
    "loginAt":   t1y.Timestamp(1705312200),

    // 数值类型
    "active":        t1y.Boolean(true),
    "quantity":      t1y.Integer(42),
    "bigNumber":     t1y.Bigint(9007199254740991),
    "rating":        t1y.Float(4.5),
    "preciseValue":  t1y.Double(3.141592653589793),

    // 结构化类型
    "tags":     t1y.Array([]any{"javascript", "typescript"}),
    "metadata": t1y.Map_(map[string]any{"theme": "dark", "lang": "zh"}),
    "history":  t1y.MapArray([]map[string]any{
        {"action": "login"},
        {"action": "logout"},
    }),

    // 空值
    "deletedAt":  t1y.Null,  // 服务端转为 nil
    "middleName": t1y.None,  // 服务端转为 nil

    // 服务端时间辅助
    "customTimeAt":      t1y.TimeNow.Now(),              // 服务端的 time.Now()
    "unixCreatedAt":     t1y.TimeNow.NowUnix(),          // 服务端的 time.Now().Unix()
})
```

## 元数据

```go
// 获取全部元数据
resp, err := client.GetMeta(ctx, "")
fmt.Println(resp.Data.Results) // { "version": 1, "collections": [...], ... }

// 获取指定字段
fieldResp, err := client.GetMetaField(ctx, "version")
fmt.Println(fieldResp.Data.Result) // 1

// 检查更新
hasUpdate, err := client.CheckUpdate(ctx)
```

## 云函数

```go
// 调用 .jsc 云函数
resp, err := client.CallFunc(ctx, "hello", map[string]any{"name": "World"}, nil)

// 为此调用单独启用安全模式
safeMode := true
resp, err = client.CallFunc(ctx, "secureFunc", params, &safeMode)
```

## 安全机制

### 认证请求头

每个请求都会携带以下请求头：

- `X-T1Y-Application-ID` — 应用 ID
- `X-T1Y-API-Key` — 32 位 API Key
- `X-T1Y-Safe-Timestamp` — Unix 时间戳（UTC + 初始化时获取的时间偏移）
- `X-T1Y-Safe-Sign` — HMAC-SHA256 签名（64 位十六进制）

### 签名算法

```
message = METHOD + "\n" + URL_PATH + "\n" + SHA256(body) + "\n" + appId + "\n" + timestamp
signature = HMAC-SHA256(secretKey, message)
```

### 安全模式（AES-256-GCM）

当启用安全模式时（通过 `IsSafeMode: true` 或初始化时自动检测），请求体将使用 AES-256-GCM 加密，密钥为应用的 SecretKey，服务端响应也会自动解密。

## API 参考

### T1YOS

| 方法                                             | 说明                                        |
| ------------------------------------------------ | ------------------------------------------- |
| `NewClient(config)`                              | 创建客户端（校验 AppID、APIKey、SecretKey） |
| `Init(ctx)`                                      | 与服务端同步时间偏移和安全模式              |
| `GetMeta(ctx, field)`                            | 获取应用元数据                              |
| `GetMetaField(ctx, field)`                       | 获取指定元数据字段                          |
| `CheckUpdate(ctx)`                               | 检查是否存在新版本                          |
| `CallFunc(ctx, name, params, enableSafeMode)`    | 调用云函数                                  |
| `Request(ctx, method, path, params, encryption)` | 原始认证请求                                |

### T1Collection

| 方法                                   | HTTP   | 端点                                |
| -------------------------------------- | ------ | ----------------------------------- |
| `InsertOne(data)`                      | POST   | `/v5/classes/:name`                 |
| `DeleteByID(objectID)`                 | DELETE | `/v5/classes/:name/:objectID`       |
| `UpdateByID(objectID, data)`           | PUT    | `/v5/classes/:name/:objectID`       |
| `FindByID(objectID)`                   | GET    | `/v5/classes/:name/:objectID`       |
| `DeleteOne(filter)`                    | DELETE | `/v5/classes/:name/one`             |
| `UpdateOne(filter, body)`              | PUT    | `/v5/classes/:name/one`             |
| `FindOne(filter)`                      | POST   | `/v5/classes/:name/one`             |
| `InsertMany(dataList)`                 | POST   | `/v5/classes/:name/many`            |
| `DeleteMany(filter)`                   | DELETE | `/v5/classes/:name/many`            |
| `UpdateMany(filter, body)`             | PUT    | `/v5/classes/:name/many`            |
| `Find(params)`                         | POST   | `/v5/classes/:name/find`            |
| `FindSimple(page, size, sort, filter)` | POST   | `/v5/classes/:name/find`            |
| `Aggregate(pipeline)`                  | POST   | `/v5/classes/:name/aggregate`       |
| `Count(filter)`                        | POST   | `/v5/classes/:name/count`           |
| `Distinct(fieldName, filter)`          | POST   | `/v5/classes/:name/distinct/:field` |
| `Create()`                             | POST   | `/v5/schemas/:name`                 |
| `Clear()`                              | PUT    | `/v5/schemas/:name`                 |
| `Drop()`                               | DELETE | `/v5/schemas/:name`                 |

### DB 对象

| 方法                  | HTTP | 端点                     |
| --------------------- | ---- | ------------------------ |
| `Collection(name)`    | —    | 获取集合操作实例         |
| `ToObjectID(id)`      | —    | 创建 ObjectID 标记字符串 |
| `GetCollections(ctx)` | GET  | `/v5/schemas`            |

## License

MIT

Copyright (c) 2026 华易云联（杭州）网络科技有限责任公司
