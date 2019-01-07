package sqlstore

import (
	"github.com/grafana/grafana_bmtech/pkg/bus"
	m "github.com/grafana/grafana_bmtech/pkg/models"
)

func init() {
	bus.AddHandler("sql", GetDBHealthQuery)
}

func GetDBHealthQuery(query *m.GetDBHealthQuery) error {
	return x.Ping()
}
