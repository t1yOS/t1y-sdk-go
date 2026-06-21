package t1y

// ApiResponse is the standard API response wrapper returned by the t1yOS server.
type ApiResponse[T any] struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    T      `json:"data"`
}

// InsertResult is the response from insertOne.
type InsertResult struct {
	ObjectID string `json:"objectId"`
}

// InsertManyResult is the response from insertMany.
type InsertManyResult struct {
	ObjectIDs     []string `json:"objectIds"`
	InsertedCount int      `json:"insertedCount"`
}

// DeleteResult is the response from deleteOne / deleteById.
type DeleteResult struct {
	DeletedCount int `json:"deletedCount"`
}

// DeleteManyResult is the response from deleteMany.
type DeleteManyResult struct {
	DeletedCount int `json:"deletedCount"`
}

// UpdateResult is the response from updateOne / updateById.
type UpdateResult struct {
	ModifiedCount int `json:"modifiedCount"`
}

// UpdateManyResult is the response from updateMany.
type UpdateManyResult struct {
	ModifiedCount int `json:"modifiedCount"`
}

// FindResult is the response from findOne / findById.
type FindResult struct {
	Result map[string]any `json:"result"`
}

// Pagination contains pagination metadata.
type Pagination struct {
	TotalItems int `json:"totalItems"`
	TotalPages int `json:"totalPages"`
}

// PaginationResult is the response from find (paginated query).
type PaginationResult struct {
	Results    []map[string]any `json:"results"`
	Page       int              `json:"page"`
	Size       int              `json:"size"`
	Pagination Pagination       `json:"pagination"`
}

// AggregateResult is the response from aggregate.
type AggregateResult struct {
	Results []map[string]any `json:"results"`
}

// InitResult is the response from GET /init/:appId.
type InitResult struct {
	Unix       int64 `json:"unix"`
	IsSafeMode bool  `json:"is_safe_mode"`
}

// CountResult is the response from count.
type CountResult struct {
	Count int `json:"count"`
}

// DistinctResult is the response from distinct.
type DistinctResult struct {
	Results []any `json:"results"`
}

// CollectionsResult is the response from getCollections.
type CollectionsResult struct {
	Results []string `json:"results"`
}

// MetaResult is the response from getMeta with a specific field.
type MetaResult struct {
	Result any `json:"result"`
}

// MetaResults is the response from getMeta without a field.
type MetaResults struct {
	Results map[string]any `json:"results"`
}

// ClearResult is the response from clear.
type ClearResult struct {
	DeletedCount int `json:"deletedCount"`
}
