package reportevents

import (
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
)

type ReportChanged struct {
	ChangedPanels  []*modconfig.ReportTreeItemDiffs
	ChangedReports []*modconfig.ReportTreeItemDiffs

	NewPanels  []*modconfig.Panel
	NewReports []*modconfig.Report

	DeletedPanels  []*modconfig.Panel
	DeletedReports []*modconfig.Report
}

// IsReportEvent implements ReportEvent interface
func (*ReportChanged) IsReportEvent() {}

func (c *ReportChanged) HasChanges() bool {
	return len(c.ChangedPanels)+
		len(c.ChangedReports)+
		len(c.NewPanels)+
		len(c.NewReports)+
		len(c.DeletedPanels)+
		len(c.DeletedReports) > 0
}
