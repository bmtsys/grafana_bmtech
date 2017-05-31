package odbc

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	"errors"

	_ "github.com/alexbrainman/odbc"
	"github.com/go-sql-driver/mysql"
	"github.com/grafana/grafana/pkg/components/simplejson"
	"github.com/grafana/grafana/pkg/log"
	"github.com/grafana/grafana/pkg/models"
	"github.com/grafana/grafana/pkg/tsdb"
)

type OdbcExecutor struct {
	DataSource *models.DataSource
	DB         *sql.DB
	Log        log.Logger
}

type DsnCacheType struct {
	cache    map[int64]*sql.DB
	versions map[int64]int
	sync.Mutex
}

var dsnCache = DsnCacheType{
	cache:    make(map[int64]*sql.DB),
	versions: make(map[int64]int),
}

func init() {
	tsdb.RegisterExecutor("odbc", NewOdbcExecutor)
}

func NewOdbcExecutor(datasource *models.DataSource) (tsdb.Executor, error) {
	executor := &OdbcExecutor{
		DataSource: datasource,
		Log:        log.New("tsdb.odbc"),
	}

	err := executor.initDsn()
	if err != nil {
		return nil, err
	}

	return executor, nil
}

func (e *OdbcExecutor) initDsn() error {
	dsnCache.Lock()
	defer dsnCache.Unlock()

	if db, present := dsnCache.cache[e.DataSource.Id]; present {
		if version, _ := dsnCache.versions[e.DataSource.Id]; version == e.DataSource.Version {
			e.DB = db
			return nil
		}
	}

	dsn := fmt.Sprintf("DSN=%s", e.DataSource.Database)
	e.Log.Debug("initDsn", "connection", dsn)

	db, err := sql.Open("odbc", dsn)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)
	if err != nil {
		return err
	}

	dsnCache.cache[e.DataSource.Id] = db
	e.DB = db
	return nil
}

func (e *OdbcExecutor) Execute(ctx context.Context, queries tsdb.QuerySlice, context *tsdb.QueryContext) *tsdb.BatchResult {
	result := &tsdb.BatchResult{
		QueryResults: make(map[string]*tsdb.QueryResult),
	}

	// macroEngine := NewMysqlMacroEngine(context.TimeRange)
	// session := e.engine.NewSession()
	db := e.DB
	defer db.Close()

	for _, query := range queries {
		rawSql := query.Model.Get("rawSql").MustString()
		if rawSql == "" {
			continue
		}

		queryResult := &tsdb.QueryResult{Meta: simplejson.New(), RefId: query.RefId}
		result.QueryResults[query.RefId] = queryResult

		// rawSql, err := macroEngine.Interpolate(rawSql)
		// if err != nil {
		// 	queryResult.Error = err
		// 	continue
		// }

		queryResult.Meta.Set("sql", rawSql)

		rows, err := db.Query(rawSql)
		if err != nil {
			queryResult.Error = err
			continue
		}

		defer rows.Close()

		format := query.Model.Get("format").MustString("time_series")

		switch format {
		case "time_series":
			queryResult.Error = errors.New("time_series for odbc not implemented yet")
			continue
		case "table":
			err := e.TransformToTable(query, rows, queryResult)
			if err != nil {
				queryResult.Error = err
				continue
			}
		}
	}

	return result
}

func (e OdbcExecutor) TransformToTable(query *tsdb.Query, rows *sql.Rows, result *tsdb.QueryResult) error {
	columnNames, err := rows.Columns()
	columnCount := len(columnNames)

	if err != nil {
		return err
	}

	table := &tsdb.Table{
		Columns: make([]tsdb.TableColumn, columnCount),
		Rows:    make([]tsdb.RowValues, 0),
	}

	for i, name := range columnNames {
		table.Columns[i].Text = name
	}

	columnTypes, err := rows.ColumnTypes()
	if err != nil {
		return err
	}

	rowLimit := 1000000
	rowCount := 0

	for ; rows.Next(); rowCount += 1 {
		if rowCount > rowLimit {
			return fmt.Errorf("MySQL query row limit exceeded, limit %d", rowLimit)
		}

		values, err := e.getTypedRowData(columnTypes, rows)
		if err != nil {
			return err
		}

		table.Rows = append(table.Rows, values)
	}

	result.Tables = append(result.Tables, table)
	result.Meta.Set("rowCount", rowCount)
	return nil
}

func (e OdbcExecutor) getTypedRowData(types []*sql.ColumnType, rows *sql.Rows) (tsdb.RowValues, error) {
	values := make([]interface{}, len(types))

	for i, stype := range types {
		switch stype.DatabaseTypeName() {
		case mysql.FieldTypeNameTiny:
			values[i] = new(int8)
		case mysql.FieldTypeNameInt24:
			values[i] = new(int32)
		case mysql.FieldTypeNameShort:
			values[i] = new(int16)
		case mysql.FieldTypeNameVarString:
			values[i] = new(string)
		case mysql.FieldTypeNameVarChar:
			values[i] = new(string)
		case mysql.FieldTypeNameLong:
			values[i] = new(int)
		case mysql.FieldTypeNameLongLong:
			values[i] = new(int64)
		case mysql.FieldTypeNameDouble:
			values[i] = new(float64)
		case mysql.FieldTypeNameDecimal:
			values[i] = new(float32)
		case mysql.FieldTypeNameNewDecimal:
			values[i] = new(float64)
		case mysql.FieldTypeNameTimestamp:
			values[i] = new(time.Time)
		case mysql.FieldTypeNameDateTime:
			values[i] = new(time.Time)
		case mysql.FieldTypeNameTime:
			values[i] = new(time.Duration)
		case mysql.FieldTypeNameYear:
			values[i] = new(int16)
		case mysql.FieldTypeNameNULL:
			values[i] = nil
		default:
			values[i] = new(string)
			// return nil, fmt.Errorf("Database type %s not supported", stype.DatabaseTypeName())
		}
	}

	if err := rows.Scan(values...); err != nil {
		return nil, err
	}

	return values, nil
}
