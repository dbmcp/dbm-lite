package service

import (
	"context"
	"database/sql"
	"sync"
	"time"

	"github.com/google/uuid"
)

type ExecutionID string

type ExecutionContext struct {
	ID        ExecutionID
	Ctx       context.Context
	Cancel    context.CancelFunc
	StartTime time.Time
	SQL       string
	DatasourceID string
}

var (
	executions   = make(map[ExecutionID]*ExecutionContext)
	executionsMu sync.RWMutex
)

func NewExecutionContext(datasourceID, sql string) *ExecutionContext {
	ctx, cancel := context.WithCancel(context.Background())
	id := ExecutionID(uuid.NewString())
	
	ec := &ExecutionContext{
		ID:            id,
		Ctx:           ctx,
		Cancel:        cancel,
		StartTime:     time.Now(),
		SQL:           sql,
		DatasourceID:  datasourceID,
	}
	
	executionsMu.Lock()
	executions[id] = ec
	executionsMu.Unlock()
	
	return ec
}

func (ec *ExecutionContext) Done() {
	executionsMu.Lock()
	delete(executions, ec.ID)
	executionsMu.Unlock()
}

func CancelExecution(id ExecutionID) bool {
	executionsMu.Lock()
	defer executionsMu.Unlock()
	
	ec, exists := executions[id]
	if !exists {
		return false
	}
	
	ec.Cancel()
	delete(executions, id)
	return true
}

func GetExecution(id ExecutionID) (*ExecutionContext, bool) {
	executionsMu.RLock()
	defer executionsMu.RUnlock()
	
	ec, exists := executions[id]
	return ec, exists
}

type QueryResult struct {
	Columns      []string                 `json:"columns"`
	Rows         []map[string]interface{} `json:"rows"`
	AffectedRows int64                    `json:"affectedRows"`
	Err          error                   `json:"-"`
}

func QueryWithContext(ctx context.Context, db *sql.DB, query string) (*QueryResult, error) {
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, err
	}
	defer rows.Close()
	
	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	
	dataRows := []map[string]interface{}{}
	for rows.Next() {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
		
		vals := make([]interface{}, len(cols))
		ptrs := make([]interface{}, len(cols))
		for i := range vals {
			ptrs[i] = &vals[i]
		}
		
		if err := rows.Scan(ptrs...); err != nil {
			return nil, err
		}
		
		row := make(map[string]interface{})
		for i, c := range cols {
			if b, ok := vals[i].([]byte); ok {
				row[c] = string(b)
			} else {
				row[c] = vals[i]
			}
		}
		dataRows = append(dataRows, row)
	}
	
	return &QueryResult{
		Columns: cols,
		Rows:    dataRows,
	}, nil
}
