# t1yOS SDK for Go

[中文文档](./README.zh-CN.md)

[t1yOS](https://www.t1y.net) Serverless Platform Go SDK — cloud database, metadata, and cloud functions client.

## Installation

```bash
go get github.com/t1yOS/t1y-sdk-go
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"

    t1y "github.com/t1yOS/t1y-sdk-go"
)

func main() {
    // 1. Create client
    client, err := t1y.NewClient(&t1y.Config{
        AppID:     1001,                                    // Required: application ID (>= 1001)
        APIKey:    "4fd7448cdc684431a62d8a0111dc6973",     // Required: 32-character API Key
        SecretKey: "17b784e359c946ffa65eebbf9ce29752",     // Required: 32-character Secret Key
        // Optional with defaults:
        // BaseURL: "https://myapp.t1y.net",
        // Version: 0,
        // IsSafeMode: false,
        // TimeFormat: "YYYY-MM-DD HH:mm:ss",
        // Offset: 0,
    })
    if err != nil {
        log.Fatal(err)
    }

    // 2. Initialize (syncs time offset and safe mode with server)
    ctx := context.Background()
    if err := client.Init(ctx); err != nil {
        log.Printf("Warning: %v", err)
    }

    // 3. Use the database!
    resp, err := client.DB.Collection("users").InsertOne(ctx, map[string]any{
        "name":   "Alice",
        "age":    25,
        "active": true,
    })
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(resp.Data.ObjectID)
}
```

## Database Operations

### Single Document

```go
db := client.DB.Collection("users")
ctx := context.Background()

// Insert one
resp, err := db.InsertOne(ctx, map[string]any{"name": "Alice", "age": 25})
fmt.Println(resp.Data.ObjectID) // "507f1f77bcf86cd799439011"

// Find by ObjectID
resp, err := db.FindByID(ctx, "507f1f77bcf86cd799439011")
fmt.Println(resp.Data.Result) // { "_id": "507f1f77...", "name": "Alice", ... }

// Update by ObjectID
_, err = db.UpdateByID(ctx, "507f1f77bcf86cd799439011", map[string]any{
    "$set": map[string]any{"age": 26},
})

// Delete by ObjectID
_, err = db.DeleteByID(ctx, "507f1f77bcf86cd799439011")
```

### Filter-based Operations

```go
// Find one by filter
resp, err := db.FindOne(ctx, map[string]any{"name": "Alice"})

// Update one by filter
_, err = db.UpdateOne(
    ctx,
    map[string]any{"name": "Alice"},                       // filter
    map[string]any{"$set": map[string]any{"age": 27}},     // update body
)

// Delete one by filter
_, err = db.DeleteOne(ctx, map[string]any{"name": "Alice"})
```

### Bulk Operations

```go
// Insert many
resp, err := db.InsertMany(ctx, []map[string]any{
    {"name": "Alice", "age": 25},
    {"name": "Bob", "age": 30},
})
fmt.Println(resp.Data.InsertedCount) // 2

// Delete many
_, err = db.DeleteMany(ctx, map[string]any{"age": map[string]any{"$lt": 18}})

// Update many
_, err = db.UpdateMany(
    ctx,
    map[string]any{"status": "inactive"},
    map[string]any{"$set": map[string]any{"status": "archived"}},
)
```

### Advanced Queries

```go
// Paginated find
resp, err := db.FindSimple(ctx, 1, 20,
    map[string]int{"createdAt": -1},               // sort (newest first)
    map[string]any{"age": map[string]any{"$gte": 18}}, // filter
)
fmt.Println(resp.Data.Results)    // Array of documents
fmt.Println(resp.Data.Pagination) // { totalItems: 42, totalPages: 3 }

// Or use FindParams
resp, err = db.Find(ctx, t1y.FindParams{
    Page:   1,
    Size:   20,
    Sort:   map[string]int{"createdAt": -1},
    Filter: map[string]any{"age": map[string]any{"$gte": 18}},
})

// Aggregation pipeline
resp, err := db.Aggregate(ctx, []map[string]any{
    {"$match": map[string]any{"status": "completed"}},
    {"$group": map[string]any{
        "_id": "$category",
        "total": map[string]any{"$sum": "$amount"},
    }},
    {"$sort": map[string]any{"total": -1}},
})

// Count
countResp, err := db.Count(ctx, map[string]any{"status": "active"})
fmt.Println(countResp.Data.Count)

// Distinct values
distinctResp, err := db.Distinct(ctx, "city", nil)
// With filter
distinctResp, err = db.Distinct(ctx, "city", map[string]any{"country": "China"})
```

### Schema Management

```go
// Get all collections
resp, err := client.DB.GetCollections(ctx)
fmt.Println(resp.Data.Results) // ["users", "orders", "products"]

// Create a collection
_, err = client.DB.Collection("posts").Create(ctx)

// Clear a collection
clearResp, err := client.DB.Collection("posts").Clear(ctx)
fmt.Println(clearResp.Data.DeletedCount)

// Drop a collection
_, err = client.DB.Collection("posts").Drop(ctx)
```

## Special Types

The SDK provides helper functions that produce server-recognized type markers:

```go
import t1y "github.com/t1yOS/t1y-sdk-go"

client.DB.Collection("users").InsertOne(ctx, map[string]any{
    // ObjectID reference
    "userId": t1y.ObjectID("507f1f77bcf86cd799439011"),

    // Date types
    "birthday":  t1y.Date("2000-01-01T00:00:00Z"),
    "eventTime": t1y.DateTime("2024-06-15T14:30:00Z"),
    "loginAt":   t1y.Timestamp(1705312200),

    // Numeric types
    "active":        t1y.Boolean(true),
    "quantity":      t1y.Integer(42),
    "bigNumber":     t1y.Bigint(9007199254740991),
    "rating":        t1y.Float(4.5),
    "preciseValue":  t1y.Double(3.141592653589793),

    // Structured types
    "tags":     t1y.Array([]any{"javascript", "typescript"}),
    "metadata": t1y.Map_(map[string]any{"theme": "dark", "lang": "en"}),
    "history":  t1y.MapArray([]map[string]any{
        {"action": "login"},
        {"action": "logout"},
    }),

    // Null values
    "deletedAt":  t1y.Null,  // server converts to nil
    "middleName": t1y.None,  // server converts to nil

    // Server-time helpers
    "customTimeAt":      t1y.TimeNow.Now(),              // server's time.Now()
    "unixCreatedAt":     t1y.TimeNow.NowUnix(),          // server's time.Now().Unix()
})
```

## Metadata

```go
// Get all metadata
resp, err := client.GetMeta(ctx, "")
fmt.Println(resp.Data.Results) // { "version": 1, "collections": [...], ... }

// Get specific field
fieldResp, err := client.GetMetaField(ctx, "version")
fmt.Println(fieldResp.Data.Result) // 1

// Check for updates
hasUpdate, err := client.CheckUpdate(ctx)
```

## Cloud Functions

```go
// Call a .jsc cloud function
resp, err := client.CallFunc(ctx, "hello", map[string]any{"name": "World"}, nil)

// With safe mode enabled for this specific call
safeMode := true
resp, err = client.CallFunc(ctx, "secureFunc", params, &safeMode)
```

## Security

### Authentication Headers

Every request includes:

- `X-T1Y-Application-ID` — Your application ID
- `X-T1Y-API-Key` — Your 32-character API key
- `X-T1Y-Safe-Timestamp` — Unix timestamp (UTC + time offset from init)
- `X-T1Y-Safe-Sign` — HMAC-SHA256 signature (64 hex chars)

### Signature Algorithm

```
message = METHOD + "\n" + URL_PATH + "\n" + SHA256(body) + "\n" + appId + "\n" + timestamp
signature = HMAC-SHA256(secretKey, message)
```

### Safe Mode (AES-256-GCM)

When safe mode is enabled (via `IsSafeMode: true` or auto-detected from init), request bodies are encrypted with AES-256-GCM using your SecretKey, and server responses are decrypted automatically.

## API Reference

### T1YOS

| Method                                           | Description                                        |
| ------------------------------------------------ | -------------------------------------------------- |
| `NewClient(config)`                              | Create client (validates AppID, APIKey, SecretKey) |
| `Init(ctx)`                                      | Sync time offset and safe mode with server         |
| `GetMeta(ctx, field)`                            | Get application metadata                           |
| `GetMetaField(ctx, field)`                       | Get a specific metadata field                      |
| `CheckUpdate(ctx)`                               | Check if newer version exists                      |
| `CallFunc(ctx, name, params, enableSafeMode)`    | Call a cloud function                              |
| `Request(ctx, method, path, params, encryption)` | Raw authenticated request                          |

### T1Collection

| Method                                 | HTTP   | Endpoint                            |
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

### DB Object

| Method                | HTTP | Endpoint                      |
| --------------------- | ---- | ----------------------------- |
| `Collection(name)`    | —    | Get a collection instance     |
| `ToObjectID(id)`      | —    | Create ObjectID marker string |
| `GetCollections(ctx)` | GET  | `/v5/schemas`                 |

## License

MIT

Copyright (c) 2026 华易云联（杭州）网络科技有限责任公司
