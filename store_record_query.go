package vaultstore

import (
	"errors"
	"strings"

	"github.com/doug-martin/goqu/v9"
	"github.com/dracory/sb"
	"github.com/dromara/carbon/v2"
)

// ============================================================================//
// CONSTRUCTOR
// ============================================================================//

// RecordQuery creates a new record query
func RecordQuery() RecordQueryInterface {
	return &recordQueryImpl{
		properties: make(map[string]interface{}),
	}
}

// ============================================================================//
// TYPE recordQueryImpl
// ============================================================================//

// recordQueryImpl implements the RecordQueryInterface
type recordQueryImpl struct {
	properties map[string]interface{}
}

// verify it extends the interface
var _ RecordQueryInterface = (*recordQueryImpl)(nil)

// Validate validates the record query
func (q *recordQueryImpl) Validate() error {
	if q.properties == nil {
		return errors.New("properties cannot be nil")
	}

	if q.IsIDSet() && q.GetID() == "" {
		return errors.New("id cannot be empty")
	}
	if q.IsTokenSet() && q.GetToken() == "" {
		return errors.New("token cannot be empty")
	}
	if q.IsIDInSet() && len(q.GetIDIn()) == 0 {
		return errors.New("idIn cannot be empty")
	}
	if q.IsTokenInSet() && len(q.GetTokenIn()) == 0 {
		return errors.New("tokenIn cannot be empty")
	}
	if q.IsLimitSet() && q.GetLimit() < 0 {
		return errors.New("limit cannot be negative")
	}
	if q.IsOffsetSet() && q.GetOffset() < 0 {
		return errors.New("offset cannot be negative")
	}
	if q.IsSortOrderSet() && !strings.EqualFold(q.GetSortOrder(), sb.ASC) && !strings.EqualFold(q.GetSortOrder(), sb.DESC) {
		return errors.New("sortOrder must be 'asc' or 'desc'")
	}

	if q.IsCountOnlySet() && (q.IsLimitSet() || q.IsOffsetSet()) {
		return errors.New("countOnly cannot be used with limit or offset")
	}
	return nil
}

func (rq *recordQueryImpl) toSelectDataset(store StoreInterface) (selectDataset *goqu.SelectDataset, selectColumns []any, err error) {
	if store == nil {
		return nil, []any{}, errors.New("store is nil")
	}

	if err := rq.Validate(); err != nil {
		return nil, []any{}, err
	}

	q := goqu.Dialect(store.GetDbDriverName()).From(store.GetVaultTableName())

	if rq.IsIDSet() && rq.GetID() != "" {
		q = q.Where(goqu.C(COLUMN_ID).Eq(rq.GetID()))
	}

	if rq.IsTokenSet() && rq.GetToken() != "" {
		q = q.Where(goqu.C(COLUMN_VAULT_TOKEN).Eq(rq.GetToken()))
	}

	if rq.IsIDInSet() && len(rq.GetIDIn()) > 0 {
		q = q.Where(goqu.C(COLUMN_ID).In(rq.GetIDIn()))
	}

	if rq.IsTokenInSet() && len(rq.GetTokenIn()) > 0 {
		q = q.Where(goqu.C(COLUMN_VAULT_TOKEN).In(rq.GetTokenIn()))
	}

	if !rq.IsCountOnlySet() {
		if rq.IsLimitSet() && rq.GetLimit() > 0 {
			q = q.Limit(uint(rq.GetLimit()))
		}

		if rq.IsOffsetSet() && rq.GetOffset() > 0 {
			q = q.Offset(uint(rq.GetOffset()))
		}
	}

	sortOrder := sb.DESC
	if rq.IsSortOrderSet() && rq.GetSortOrder() != "" {
		sortOrder = rq.GetSortOrder()
	}

	if rq.IsOrderBySet() && rq.GetOrderBy() != "" {
		if strings.EqualFold(sortOrder, sb.ASC) {
			q = q.Order(goqu.I(rq.GetOrderBy()).Asc())
		} else {
			q = q.Order(goqu.I(rq.GetOrderBy()).Desc())
		}
	}

	columns := []any{}

	for _, column := range rq.GetColumns() {
		columns = append(columns, column)
	}

	if rq.IsSoftDeletedIncludeSet() {
		// soft deleted requested specifically
		return q, columns, nil
	}

	// To exclude soft-deleted records, we need to find records where:
	// soft_deleted_at is greater than current time (not soft-deleted)
	// When a record is soft-deleted, soft_deleted_at is set to the current time
	// By default, soft_deleted_at is set to MAX_DATETIME (not soft-deleted)
	softDeletedFilter := goqu.C(COLUMN_SOFT_DELETED_AT).
		Gt(carbon.Now(carbon.UTC).ToDateTimeString())

	return q.Where(softDeletedFilter), columns, nil
}

func (q *recordQueryImpl) IsColumnsSet() bool {
	return q.hasProperty("columns")
}

func (q *recordQueryImpl) GetColumns() []string {
	if q.IsColumnsSet() {
		return q.properties["columns"].([]string)
	}
	return []string{}
}

func (q *recordQueryImpl) SetColumns(columns []string) RecordQueryInterface {
	q.properties["columns"] = columns
	return q
}

func (q *recordQueryImpl) IsIDSet() bool {
	return q.hasProperty("id")
}

func (q *recordQueryImpl) GetID() string {
	if q.IsIDSet() {
		return q.properties["id"].(string)
	}
	return ""
}

func (q *recordQueryImpl) SetID(id string) RecordQueryInterface {
	q.properties["id"] = id
	return q
}

func (q *recordQueryImpl) IsTokenSet() bool {
	return q.hasProperty("token")
}

func (q *recordQueryImpl) GetToken() string {
	if q.IsTokenSet() {
		return q.properties["token"].(string)
	}
	return ""
}

func (q *recordQueryImpl) SetToken(token string) RecordQueryInterface {
	q.properties["token"] = token
	return q
}

func (q *recordQueryImpl) IsIDInSet() bool {
	return q.hasProperty("idIn")
}

func (q *recordQueryImpl) GetIDIn() []string {
	if q.IsIDInSet() {
		return q.properties["idIn"].([]string)
	}
	return []string{}
}

func (q *recordQueryImpl) SetIDIn(idIn []string) RecordQueryInterface {
	q.properties["idIn"] = idIn
	return q
}

func (q *recordQueryImpl) IsTokenInSet() bool {
	return q.hasProperty("tokenIn")
}

func (q *recordQueryImpl) GetTokenIn() []string {
	if q.IsTokenInSet() {
		return q.properties["tokenIn"].([]string)
	}
	return []string{}
}

func (q *recordQueryImpl) SetTokenIn(tokenIn []string) RecordQueryInterface {
	q.properties["tokenIn"] = tokenIn
	return q
}

func (q *recordQueryImpl) IsOffsetSet() bool {
	return q.hasProperty("offset")
}

func (q *recordQueryImpl) GetOffset() int {
	if q.IsOffsetSet() {
		return q.properties["offset"].(int)
	}
	return 0
}

func (q *recordQueryImpl) SetOffset(offset int) RecordQueryInterface {
	q.properties["offset"] = offset
	return q
}

func (q *recordQueryImpl) IsOrderBySet() bool {
	return q.hasProperty("orderBy")
}

func (q *recordQueryImpl) GetOrderBy() string {
	if q.IsOrderBySet() {
		return q.properties["orderBy"].(string)
	}
	return ""
}

func (q *recordQueryImpl) SetOrderBy(orderBy string) RecordQueryInterface {
	q.properties["orderBy"] = orderBy
	return q
}

func (q *recordQueryImpl) IsCountOnlySet() bool {
	return q.hasProperty("countOnly")
}

func (q *recordQueryImpl) GetCountOnly() bool {
	if q.IsCountOnlySet() {
		return q.properties["countOnly"].(bool)
	}
	return false
}

func (q *recordQueryImpl) SetCountOnly(countOnly bool) RecordQueryInterface {
	q.properties["countOnly"] = countOnly
	return q
}

func (q *recordQueryImpl) IsSortOrderSet() bool {
	return q.hasProperty("sortOrder")
}

func (q *recordQueryImpl) GetSortOrder() string {
	if q.IsSortOrderSet() {
		return q.properties["sortOrder"].(string)
	}
	return ""
}

func (q *recordQueryImpl) SetSortOrder(sortOrder string) RecordQueryInterface {
	q.properties["sortOrder"] = sortOrder
	return q
}

func (q *recordQueryImpl) IsSoftDeletedIncludeSet() bool {
	return q.hasProperty("softDeletedInclude")
}

func (q *recordQueryImpl) GetSoftDeletedInclude() bool {
	if q.IsSoftDeletedIncludeSet() {
		return q.properties["softDeletedInclude"].(bool)
	}
	return false
}

func (q *recordQueryImpl) SetSoftDeletedInclude(softDeletedInclude bool) RecordQueryInterface {
	q.properties["softDeletedInclude"] = softDeletedInclude
	return q
}

func (q *recordQueryImpl) IsLimitSet() bool {
	return q.hasProperty("limit")
}

func (q *recordQueryImpl) GetLimit() int {
	if q.IsLimitSet() {
		return q.properties["limit"].(int)
	}
	return 0
}

func (q *recordQueryImpl) SetLimit(limit int) RecordQueryInterface {
	q.properties["limit"] = limit
	return q
}

func (q *recordQueryImpl) hasProperty(key string) bool {
	_, ok := q.properties[key]
	return ok
}
