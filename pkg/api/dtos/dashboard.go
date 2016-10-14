package dtos

import "github.com/grafana/grafana/pkg/components/simplejson"

type SaveDashboardCmd struct {
	Dashboard *simplejson.Json   `json:"dashboard" binding:"Required"`
	Alerts    []*simplejson.Json `json:"alerts"`
	Overwrite bool               `json:"overwrite"`
}
