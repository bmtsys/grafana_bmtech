package oracle

import (
	"context"
	"database/sql"
	"sync"

	"github.com/grafana/grafana/pkg/components/simplejson"
	"github.com/grafana/grafana/pkg/models"
	"github.com/grafana/grafana/pkg/tsdb"
)

type OracleSqlEngine struct {
	MacroEngine tsdb.SqlMacroEngine
	DB          *sql.DB
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

// InitEngine creates the db connection and inits the dsn or loads it from the dsn cache
func (e *OracleSqlEngine) InitEngine(dsInfo *models.DataSource, cnnstr string) error {
	dsnCache.Lock()
	defer dsnCache.Unlock()

	if dsn, present := dsnCache.cache[dsInfo.Id]; present {
		if version, _ := dsnCache.versions[dsInfo.Id]; version == dsInfo.Version {
			e.DB = dsn
			return nil
		}
	}

	db, err := sql.Open("oci8", cnnstr)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)
	if err != nil {
		return err
	}

	dsnCache.cache[dsInfo.Id] = db
	e.DB = db

	return nil
}

// Query is a default implementation of the Query method for an SQL data source.
// The caller of this function must implement transformToTimeSeries and transformToTable and
// pass them in as parameters.
//
// Note: Oracle uses transformToTimeSeries() and transformToTable() with different signatures (general *sql.Rows
// instead of xorm *core.Rows) so it doesn't implement SqlEngine interface
func (e *OracleSqlEngine) Query(
	ctx context.Context,
	dsInfo *models.DataSource,
	tsdbQuery *tsdb.TsdbQuery,
	transformToTimeSeries func(query *tsdb.Query, rows *sql.Rows, result *tsdb.QueryResult) error,
	transformToTable func(query *tsdb.Query, rows *sql.Rows, result *tsdb.QueryResult) error,
) (*tsdb.Response, error) {
	result := &tsdb.Response{
		Results: make(map[string]*tsdb.QueryResult),
	}

	db := e.DB
	defer db.Close()

	for _, query := range tsdbQuery.Queries {
		rawSql := query.Model.Get("rawSql").MustString()
		if rawSql == "" {
			continue
		}

		queryResult := &tsdb.QueryResult{Meta: simplejson.New(), RefId: query.RefId}
		result.Results[query.RefId] = queryResult

		rawSql, err := e.MacroEngine.Interpolate(tsdbQuery.TimeRange, rawSql)
		if err != nil {
			queryResult.Error = err
			continue
		}

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
			err := transformToTimeSeries(query, rows, queryResult)
			if err != nil {
				queryResult.Error = err
				continue
			}
		case "table":
			err := transformToTable(query, rows, queryResult)
			if err != nil {
				queryResult.Error = err
				continue
			}
		}
	}

	return result, nil
}
