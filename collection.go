package t1y

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// T1Collection provides chainable CRUD and schema operations for a database collection.
// Created via client.DB.Collection("name") — never instantiated directly.
type T1Collection struct {
	client *T1YOS
	name   string
}

// newCollection creates a new T1Collection instance.
func newCollection(client *T1YOS, name string) *T1Collection {
	if name == "" {
		panic("Collection name must be a non-empty string")
	}
	return &T1Collection{
		client: client,
		name:   name,
	}
}

// ==================== Single Document Operations ====================

// InsertOne inserts one document into the collection.
//
// The data must be a non-empty map.
// Returns the inserted document's ObjectID.
func (c *T1Collection) InsertOne(ctx context.Context, data map[string]any) (*ApiResponse[InsertResult], error) {
	if !IsNonEmptyObject(data) {
		return nil, NewValidationError("insertOne data must be a non-empty map")
	}
	return request[InsertResult](c.client, ctx, http.MethodPost, c.path(), data, c.client.config.isSafeMode)
}

// DeleteByID deletes one document by ObjectID.
func (c *T1Collection) DeleteByID(ctx context.Context, objectID string) (*ApiResponse[DeleteResult], error) {
	if err := assertObjectID(objectID); err != nil {
		return nil, err
	}
	return request[DeleteResult](c.client, ctx, http.MethodDelete, c.path()+"/"+objectID, nil, c.client.config.isSafeMode)
}

// UpdateByID updates one document by ObjectID.
//
// The data must be a non-empty map with MongoDB update operators.
func (c *T1Collection) UpdateByID(ctx context.Context, objectID string, data map[string]any) (*ApiResponse[UpdateResult], error) {
	if err := assertObjectID(objectID); err != nil {
		return nil, err
	}
	if !IsNonEmptyObject(data) {
		return nil, NewValidationError("update data must be a non-empty map")
	}
	return request[UpdateResult](c.client, ctx, http.MethodPut, c.path()+"/"+objectID, data, c.client.config.isSafeMode)
}

// FindByID finds one document by ObjectID.
func (c *T1Collection) FindByID(ctx context.Context, objectID string) (*ApiResponse[FindResult], error) {
	if err := assertObjectID(objectID); err != nil {
		return nil, err
	}
	return request[FindResult](c.client, ctx, http.MethodGet, c.path()+"/"+objectID, nil, c.client.config.isSafeMode)
}

// ==================== Filter-based Single Operations ====================

// DeleteOne deletes one document matching the filter.
//
// The filter must be a non-empty map.
func (c *T1Collection) DeleteOne(ctx context.Context, filter map[string]any) (*ApiResponse[DeleteResult], error) {
	if !IsNonEmptyObject(filter) {
		return nil, NewValidationError("deleteOne filter must be a non-empty map")
	}
	return request[DeleteResult](c.client, ctx, http.MethodDelete, c.path()+"/one", filter, c.client.config.isSafeMode)
}

// UpdateOne updates one document matching the filter.
//
// The filter must be a non-empty map.
// The body must be a non-empty map with MongoDB update operators.
func (c *T1Collection) UpdateOne(ctx context.Context, filter, body map[string]any) (*ApiResponse[UpdateResult], error) {
	if !IsNonEmptyObject(filter) {
		return nil, NewValidationError("updateOne filter must be a non-empty map")
	}
	if !IsNonEmptyObject(body) {
		return nil, NewValidationError("updateOne body must be a non-empty map")
	}
	return request[UpdateResult](c.client, ctx, http.MethodPut, c.path()+"/one", map[string]any{
		"filter": filter,
		"body":   body,
	}, c.client.config.isSafeMode)
}

// FindOne finds one document matching the filter.
//
// The filter must be a non-empty map.
func (c *T1Collection) FindOne(ctx context.Context, filter map[string]any) (*ApiResponse[FindResult], error) {
	if !IsNonEmptyObject(filter) {
		return nil, NewValidationError("findOne filter must be a non-empty map")
	}
	return request[FindResult](c.client, ctx, http.MethodPost, c.path()+"/one", filter, c.client.config.isSafeMode)
}

// ==================== Bulk Operations ====================

// InsertMany inserts multiple documents into the collection.
//
// The dataList must be a non-empty slice of non-empty maps.
func (c *T1Collection) InsertMany(ctx context.Context, dataList []map[string]any) (*ApiResponse[InsertManyResult], error) {
	if !IsNonEmptyArrayWithNonEmptyObjects(dataList) {
		return nil, NewValidationError("insertMany dataList must be a non-empty slice of non-empty maps")
	}
	return request[InsertManyResult](c.client, ctx, http.MethodPost, c.path()+"/many", dataList, c.client.config.isSafeMode)
}

// DeleteMany deletes multiple documents matching the filter.
//
// The filter can be an empty map to match all documents (use with caution).
func (c *T1Collection) DeleteMany(ctx context.Context, filter map[string]any) (*ApiResponse[DeleteManyResult], error) {
	if !IsPlainObject(filter) {
		return nil, NewValidationError("deleteMany filter must be a map")
	}
	return request[DeleteManyResult](c.client, ctx, http.MethodDelete, c.path()+"/many", filter, c.client.config.isSafeMode)
}

// UpdateMany updates multiple documents matching the filter.
//
// The filter must be a map.
// The body must be a non-empty map with MongoDB update operators.
func (c *T1Collection) UpdateMany(ctx context.Context, filter, body map[string]any) (*ApiResponse[UpdateManyResult], error) {
	if !IsPlainObject(filter) {
		return nil, NewValidationError("updateMany filter must be a map")
	}
	if !IsNonEmptyObject(body) {
		return nil, NewValidationError("updateMany body must be a non-empty map")
	}
	return request[UpdateManyResult](c.client, ctx, http.MethodPut, c.path()+"/many", map[string]any{
		"filter": filter,
		"body":   body,
	}, c.client.config.isSafeMode)
}

// ==================== Advanced Queries ====================

// FindParams contains the parameters for a paginated find query.
type FindParams struct {
	// Page number (1-based). Default: 1.
	Page int `json:"page"`
	// Page size (1-100). Default: 10, max: 100.
	Size int `json:"size"`
	// Sort specification, e.g. {"createdAt": -1} for newest first.
	Sort map[string]int `json:"sort"`
	// Query filter.
	Filter map[string]any `json:"filter"`
}

// Find executes a paginated find query with sorting and filtering.
func (c *T1Collection) Find(ctx context.Context, params FindParams) (*ApiResponse[PaginationResult], error) {
	if params.Page < 1 {
		params.Page = 1
	}
	if params.Size < 1 {
		params.Size = DefaultPageSize
	}
	if params.Size > MaxPageSize {
		params.Size = MaxPageSize
	}
	if params.Sort == nil || len(params.Sort) == 0 {
		params.Sort = map[string]int{"createdAt": -1}
	}
	if params.Filter == nil {
		params.Filter = make(map[string]any)
	}

	return request[PaginationResult](c.client, ctx, http.MethodPost, c.path()+"/find", map[string]any{
		"page":   params.Page,
		"size":   params.Size,
		"sort":   params.Sort,
		"filter": params.Filter,
	}, c.client.config.isSafeMode)
}

// FindSimple is a convenience method for simple paginated queries.
func (c *T1Collection) FindSimple(ctx context.Context, page, size int, sort map[string]int, filter map[string]any) (*ApiResponse[PaginationResult], error) {
	return c.Find(ctx, FindParams{
		Page:   page,
		Size:   size,
		Sort:   sort,
		Filter: filter,
	})
}

// Aggregate executes a MongoDB aggregation pipeline.
func (c *T1Collection) Aggregate(ctx context.Context, pipeline []map[string]any) (*ApiResponse[AggregateResult], error) {
	if pipeline == nil {
		return nil, NewValidationError("aggregate pipeline must be a non-nil slice")
	}
	return request[AggregateResult](c.client, ctx, http.MethodPost, c.path()+"/aggregate", pipeline, c.client.config.isSafeMode)
}

// Count counts documents matching a filter.
//
// Uses POST /v5/classes/:name/count with the filter as the request body.
// Pass an empty map to count all documents.
func (c *T1Collection) Count(ctx context.Context, filter map[string]any) (*ApiResponse[CountResult], error) {
	if !IsPlainObject(filter) {
		return nil, NewValidationError("count filter must be a map")
	}
	return request[CountResult](c.client, ctx, http.MethodPost, c.path()+"/count", filter, c.client.config.isSafeMode)
}

// Distinct returns distinct values for a field, optionally filtered.
//
// Uses POST /v5/classes/:name/distinct/:field with the filter as the request body.
func (c *T1Collection) Distinct(ctx context.Context, fieldName string, filter map[string]any) (*ApiResponse[DistinctResult], error) {
	if fieldName == "" {
		return nil, NewValidationError("distinct fieldName must be a non-empty string")
	}
	if !IsPlainObject(filter) {
		return nil, NewValidationError("distinct filter must be a map")
	}
	return request[DistinctResult](
		c.client, ctx, http.MethodPost,
		fmt.Sprintf("%s/distinct/%s", c.path(), url.PathEscape(fieldName)),
		filter,
		c.client.config.isSafeMode,
	)
}

// ==================== Schema Management ====================

// Create creates this collection (table) in the application's database.
func (c *T1Collection) Create(ctx context.Context) (*ApiResponse[any], error) {
	return request[any](c.client, ctx, http.MethodPost, fmt.Sprintf("/v5/schemas/%s", url.PathEscape(c.name)), nil, c.client.config.isSafeMode)
}

// Clear clears all documents from this collection without dropping it.
func (c *T1Collection) Clear(ctx context.Context) (*ApiResponse[ClearResult], error) {
	return request[ClearResult](c.client, ctx, http.MethodPut, fmt.Sprintf("/v5/schemas/%s", url.PathEscape(c.name)), nil, c.client.config.isSafeMode)
}

// Drop drops (deletes) this collection entirely.
func (c *T1Collection) Drop(ctx context.Context) (*ApiResponse[any], error) {
	return request[any](c.client, ctx, http.MethodDelete, fmt.Sprintf("/v5/schemas/%s", url.PathEscape(c.name)), nil, c.client.config.isSafeMode)
}

// ==================== Helpers ====================

// path returns the API path for this collection.
func (c *T1Collection) path() string {
	return fmt.Sprintf("/v5/classes/%s", url.PathEscape(c.name))
}
