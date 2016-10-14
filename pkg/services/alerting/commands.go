package alerting

import (
	"github.com/grafana/grafana/pkg/bus"
	"github.com/grafana/grafana/pkg/components/simplejson"
	m "github.com/grafana/grafana/pkg/models"
)

type SaveDashboardAlertsCommand struct {
	UserId      int64
	OrgId       int64
	DashboardId int64
	Alerts      []*simplejson.Json
}

func init() {
	bus.AddHandler("alerting", saveDashboardAlertsHandler)
}

func saveDashboardAlertsHandler(cmd *UpdateDashboardAlertsCommand) error {
	saveAlerts := m.SaveAlertsCommand{
		OrgId:       cmd.OrgId,
		UserId:      cmd.UserId,
		DashboardId: cmd.Dashboard.Id,
	}

	extractor := NewDashAlertExtractor(cmd.Dashboard, cmd.OrgId)

	if alerts, err := extractor.GetAlerts(); err != nil {
		return err
	} else {
		saveAlerts.Alerts = alerts
	}

	if err := bus.Dispatch(&saveAlerts); err != nil {
		return err
	}

	return nil
}
