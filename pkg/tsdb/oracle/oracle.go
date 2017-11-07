package oracle

import (
	"container/list"
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/grafana/grafana/pkg/components/null"
	"github.com/grafana/grafana/pkg/log"
	"github.com/grafana/grafana/pkg/models"
	"github.com/grafana/grafana/pkg/tsdb"
	_ "github.com/mattn/go-oci8"
)

type OracleQueryEndpoint struct {
	sqlEngine *OracleSqlEngine
	log       log.Logger
}

func init() {
	tsdb.RegisterTsdbQueryEndpoint("oracle", NewOracleQueryEndpoint)
}

func NewOracleQueryEndpoint(datasource *models.DataSource) (tsdb.TsdbQueryEndpoint, error) {
	endpoint := &OracleQueryEndpoint{
		log: log.New("tsdb.oracle"),
	}

	endpoint.sqlEngine = &OracleSqlEngine{
		MacroEngine: NewOracleMacroEngine(),
	}

	cnnstr := generateConnectionString(datasource)
	endpoint.log.Debug("getEngine", "connection", cnnstr)

	if err := endpoint.sqlEngine.InitEngine(datasource, cnnstr); err != nil {
		return nil, err
	}

	return endpoint, nil
}

func generateConnectionString(datasource *models.DataSource) string {
	password := ""
	for key, value := range datasource.SecureJsonData.Decrypt() {
		if key == "password" {
			password = value
			break
		}
	}

	// Oracle connection string format (SQL Connect URL format)
	// username/password@host[:port][/service name][:server][/instance_name]
	// Use User/password@Url/Database where Url may contains port
	return fmt.Sprintf("%s/%s@%s/%s", datasource.User, password, datasource.Url, datasource.Database)
}

func (e *OracleQueryEndpoint) Query(ctx context.Context, dsInfo *models.DataSource, tsdbQuery *tsdb.TsdbQuery) (*tsdb.Response, error) {
	return e.sqlEngine.Query(ctx, dsInfo, tsdbQuery, e.transformToTimeSeries, e.transformToTable)
}

func (e OracleQueryEndpoint) getTypedRowData(rows *sql.Rows) (tsdb.RowValues, error) {
	types, err := rows.ColumnTypes()
	if err != nil {
		return nil, err
	}

	values := make([]interface{}, len(types))
	valuePtrs := make([]interface{}, len(types))

	for i := 0; i < len(types); i++ {
		valuePtrs[i] = &values[i]
	}

	for i, coltype := range types {
		dbTypeName := coltype.DatabaseTypeName()
		// fmt.Printf("value: %v type: %s \n", values[i], dbTypeName)
		switch dbTypeName {
		case "SQLT_NUM", "SQLT_IBFLOAT", "SQLT_IBDOUBLE":
			values[i] = new(float64)
		case "SQLT_CHR", "SQLT_AFC", "SQLT_VCS", "SQLT_AVC", "SQLT_RDD":
			values[i] = new(string)
		case "SQLT_DAT", "SQLT_TIMESTAMP", "SQLT_TIMESTAMP_TZ", "SQLT_TIMESTAMP_LTZ":
			values[i] = new(time.Time)
		case "SQLT_INTERVAL_DS", "SQLT_INTERVAL_YM":
			values[i] = new(time.Duration)
		case "SQLT_BIN", "SQLT_BLOB", "SQLT_CLOB":
			values[i] = new(byte)
		default:
			values[i] = new(string)
		}
	}

	if err := rows.Scan(values...); err != nil {
		return nil, err
	}

	return values, nil
}

func (e OracleQueryEndpoint) transformToTable(query *tsdb.Query, rows *sql.Rows, result *tsdb.QueryResult) error {

	columnNames, err := rows.Columns()
	if err != nil {
		return err
	}

	table := &tsdb.Table{
		Columns: make([]tsdb.TableColumn, len(columnNames)),
		Rows:    make([]tsdb.RowValues, 0),
	}

	for i, name := range columnNames {
		table.Columns[i].Text = name
	}

	rowLimit := 1000000
	rowCount := 0

	for ; rows.Next(); rowCount++ {
		if rowCount > rowLimit {
			return fmt.Errorf("Oracle query row limit exceeded, limit %d", rowLimit)
		}

		values, err := e.getTypedRowData(rows)
		if err != nil {
			return err
		}

		table.Rows = append(table.Rows, values)
	}

	result.Tables = append(result.Tables, table)
	result.Meta.Set("rowCount", rowCount)
	return nil
}

func (e OracleQueryEndpoint) transformToTimeSeries(query *tsdb.Query, rows *sql.Rows, result *tsdb.QueryResult) error {
	pointsBySeries := make(map[string]*tsdb.TimeSeries)
	seriesByQueryOrder := list.New()
	columnNames, err := rows.Columns()

	if err != nil {
		return err
	}

	rowLimit := 1000000
	rowCount := 0
	timeIndex := -1
	metricIndex := -1

	// check columns of resultset
	for i, col := range columnNames {
		switch strings.ToUpper(col) {
		case "TIME":
			timeIndex = i
		case "METRIC":
			metricIndex = i
		}
	}

	if timeIndex == -1 {
		return fmt.Errorf("Found no column named time")
	}

	for rows.Next() {
		var timestamp float64
		var value null.Float
		var metric string

		if rowCount > rowLimit {
			return fmt.Errorf("Oracle query row limit exceeded, limit %d", rowLimit)
		}

		values, err := e.getTypedRowData(rows)
		if err != nil {
			return err
		}

		switch columnValue := values[timeIndex].(type) {
		case int64:
			timestamp = float64(columnValue * 1000)
		case float64:
			timestamp = columnValue * 1000
		case *time.Time:
			timestamp = float64(columnValue.Unix() * 1000)
		default:
			return fmt.Errorf("Invalid type for column time, must be of type timestamp or unix timestamp")
		}

		if metricIndex >= 0 {
			metricUntyped := values[metricIndex]
			switch columnMetric := metricUntyped.(type) {
			case string:
				metric = columnMetric
			case *string:
				metric = *columnMetric
			default:
				return fmt.Errorf("Column metric must be of type char,varchar or text")
			}
		}

		for i, col := range columnNames {
			if i == timeIndex || i == metricIndex {
				continue
			}

			switch columnValue := values[i].(type) {
			case int64:
				value = null.FloatFrom(float64(columnValue))
			case *int64:
				value = null.FloatFrom(float64(*columnValue))
			case float64:
				value = null.FloatFrom(columnValue)
			case *float64:
				value = null.FloatFrom(*columnValue)
			case nil:
				value.Valid = false
			default:
				return fmt.Errorf("Value column must have numeric datatype, column: %s type: %T value: %v", col, columnValue, columnValue)
			}
			if metricIndex == -1 {
				metric = col
			}
			e.appendTimePoint(pointsBySeries, seriesByQueryOrder, metric, timestamp, value)
			rowCount++

		}
	}

	for elem := seriesByQueryOrder.Front(); elem != nil; elem = elem.Next() {
		key := elem.Value.(string)
		result.Series = append(result.Series, pointsBySeries[key])
	}

	result.Meta.Set("rowCount", rowCount)
	return nil
}

func (e OracleQueryEndpoint) appendTimePoint(pointsBySeries map[string]*tsdb.TimeSeries, seriesByQueryOrder *list.List, metric string, timestamp float64, value null.Float) {
	if series, exist := pointsBySeries[metric]; exist {
		series.Points = append(series.Points, tsdb.TimePoint{value, null.FloatFrom(timestamp)})
	} else {
		series := &tsdb.TimeSeries{Name: metric}
		series.Points = append(series.Points, tsdb.TimePoint{value, null.FloatFrom(timestamp)})
		pointsBySeries[metric] = series
		seriesByQueryOrder.PushBack(metric)
	}
	e.log.Debug("Rows", "metric", metric, "time", timestamp, "value", value)
}
