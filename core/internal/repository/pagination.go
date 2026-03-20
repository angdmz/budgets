package repository

import (
	"fmt"
	"strings"
	"time"
)

// SortOrder represents the sort direction
type SortOrder string

const (
	SortOrderAsc  SortOrder = "ASC"
	SortOrderDesc SortOrder = "DESC"
)

// FilterOperator represents comparison operators for filters
type FilterOperator string

const (
	FilterOpEquals      FilterOperator = "="
	FilterOpNotEquals   FilterOperator = "!="
	FilterOpGreaterThan FilterOperator = ">"
	FilterOpLessThan    FilterOperator = "<"
	FilterOpGTE         FilterOperator = ">="
	FilterOpLTE         FilterOperator = "<="
	FilterOpLike        FilterOperator = "LIKE"
	FilterOpILike       FilterOperator = "ILIKE"
	FilterOpIn          FilterOperator = "IN"
	FilterOpIsNull      FilterOperator = "IS NULL"
	FilterOpIsNotNull   FilterOperator = "IS NOT NULL"
)

// Filter represents a single filter condition
type Filter struct {
	Field    string
	Operator FilterOperator
	Value    interface{}
}

func NewFilter(field string, op FilterOperator, value interface{}) Filter {
	return Filter{
		Field:    field,
		Operator: op,
		Value:    value,
	}
}

func Equals(field string, value interface{}) Filter {
	return NewFilter(field, FilterOpEquals, value)
}

func NotEquals(field string, value interface{}) Filter {
	return NewFilter(field, FilterOpNotEquals, value)
}

func GreaterThan(field string, value interface{}) Filter {
	return NewFilter(field, FilterOpGreaterThan, value)
}

func LessThan(field string, value interface{}) Filter {
	return NewFilter(field, FilterOpLessThan, value)
}

func Like(field string, pattern string) Filter {
	return NewFilter(field, FilterOpLike, pattern)
}

func ILike(field string, pattern string) Filter {
	return NewFilter(field, FilterOpILike, pattern)
}

func IsNull(field string) Filter {
	return NewFilter(field, FilterOpIsNull, nil)
}

func IsNotNull(field string) Filter {
	return NewFilter(field, FilterOpIsNotNull, nil)
}

// SortField represents a field to sort by
type SortField struct {
	Field string
	Order SortOrder
}

func Asc(field string) SortField {
	return SortField{Field: field, Order: SortOrderAsc}
}

func Desc(field string) SortField {
	return SortField{Field: field, Order: SortOrderDesc}
}

// Cursor represents a pagination cursor
type Cursor struct {
	CreatedAt time.Time
	ID        int64
}

func (c *Cursor) IsZero() bool {
	return c.CreatedAt.IsZero() && c.ID == 0
}

// PaginationRequest represents pagination parameters
type PaginationRequest struct {
	Limit      int
	After      *Cursor
	Before     *Cursor
	SortFields []SortField
}

func NewPaginationRequest(limit int) PaginationRequest {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	return PaginationRequest{
		Limit:      limit,
		SortFields: []SortField{Desc("created_at")},
	}
}

func (p *PaginationRequest) WithAfter(cursor *Cursor) *PaginationRequest {
	p.After = cursor
	return p
}

func (p *PaginationRequest) WithBefore(cursor *Cursor) *PaginationRequest {
	p.Before = cursor
	return p
}

func (p *PaginationRequest) WithSort(fields ...SortField) *PaginationRequest {
	p.SortFields = fields
	return p
}

// PageInfo contains pagination metadata
type PageInfo struct {
	HasNextPage     bool    `json:"has_next_page"`
	HasPreviousPage bool    `json:"has_previous_page"`
	StartCursor     *Cursor `json:"-"`
	EndCursor       *Cursor `json:"-"`
	TotalCount      int64   `json:"total_count"`
}

// PaginatedResult represents a paginated list result
type PaginatedResult[T any] struct {
	Items    []T      `json:"items"`
	PageInfo PageInfo `json:"page_info"`
}

func NewPaginatedResult[T any](items []T, pageInfo PageInfo) PaginatedResult[T] {
	return PaginatedResult[T]{
		Items:    items,
		PageInfo: pageInfo,
	}
}

// QueryBuilder helps build SQL queries with filters and pagination
type QueryBuilder struct {
	baseQuery   string
	filters     []Filter
	sortFields  []SortField
	pagination  *PaginationRequest
	paramOffset int
}

func NewQueryBuilder(baseQuery string) *QueryBuilder {
	return &QueryBuilder{
		baseQuery:   baseQuery,
		filters:     make([]Filter, 0),
		sortFields:  make([]SortField, 0),
		paramOffset: 0,
	}
}

func (qb *QueryBuilder) WithParamOffset(offset int) *QueryBuilder {
	qb.paramOffset = offset
	return qb
}

func (qb *QueryBuilder) WithFilters(filters ...Filter) *QueryBuilder {
	qb.filters = append(qb.filters, filters...)
	return qb
}

func (qb *QueryBuilder) WithSort(fields ...SortField) *QueryBuilder {
	qb.sortFields = fields
	return qb
}

func (qb *QueryBuilder) WithPagination(pagination *PaginationRequest) *QueryBuilder {
	qb.pagination = pagination
	if pagination != nil && len(pagination.SortFields) > 0 {
		qb.sortFields = pagination.SortFields
	}
	return qb
}

func (qb *QueryBuilder) Build() (string, []interface{}) {
	var conditions []string
	var args []interface{}
	paramIdx := qb.paramOffset + 1

	for _, f := range qb.filters {
		switch f.Operator {
		case FilterOpIsNull, FilterOpIsNotNull:
			conditions = append(conditions, fmt.Sprintf("%s %s", f.Field, f.Operator))
		case FilterOpIn:
			conditions = append(conditions, fmt.Sprintf("%s = ANY($%d)", f.Field, paramIdx))
			args = append(args, f.Value)
			paramIdx++
		default:
			conditions = append(conditions, fmt.Sprintf("%s %s $%d", f.Field, f.Operator, paramIdx))
			args = append(args, f.Value)
			paramIdx++
		}
	}

	// Add cursor-based pagination conditions
	if qb.pagination != nil && qb.pagination.After != nil && !qb.pagination.After.IsZero() {
		conditions = append(conditions, fmt.Sprintf(
			"(created_at, id) < ($%d, $%d)",
			paramIdx, paramIdx+1,
		))
		args = append(args, qb.pagination.After.CreatedAt, qb.pagination.After.ID)
		paramIdx += 2
	}

	if qb.pagination != nil && qb.pagination.Before != nil && !qb.pagination.Before.IsZero() {
		conditions = append(conditions, fmt.Sprintf(
			"(created_at, id) > ($%d, $%d)",
			paramIdx, paramIdx+1,
		))
		args = append(args, qb.pagination.Before.CreatedAt, qb.pagination.Before.ID)
		paramIdx += 2
	}

	query := qb.baseQuery
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	// Add ORDER BY
	if len(qb.sortFields) > 0 {
		var orderParts []string
		for _, sf := range qb.sortFields {
			orderParts = append(orderParts, fmt.Sprintf("%s %s", sf.Field, sf.Order))
		}
		query += " ORDER BY " + strings.Join(orderParts, ", ")
	}

	// Add LIMIT
	if qb.pagination != nil && qb.pagination.Limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", qb.pagination.Limit+1) // +1 to check for next page
	}

	return query, args
}

func (qb *QueryBuilder) BuildCount() (string, []interface{}) {
	var conditions []string
	var args []interface{}
	paramIdx := qb.paramOffset + 1

	for _, f := range qb.filters {
		switch f.Operator {
		case FilterOpIsNull, FilterOpIsNotNull:
			conditions = append(conditions, fmt.Sprintf("%s %s", f.Field, f.Operator))
		case FilterOpIn:
			conditions = append(conditions, fmt.Sprintf("%s = ANY($%d)", f.Field, paramIdx))
			args = append(args, f.Value)
			paramIdx++
		default:
			conditions = append(conditions, fmt.Sprintf("%s %s $%d", f.Field, f.Operator, paramIdx))
			args = append(args, f.Value)
			paramIdx++
		}
	}

	query := qb.baseQuery
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	return query, args
}
