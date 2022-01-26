package reportevents

import "github.com/turbot/steampipe/pkg/report/reportinterfaces"

type ReportError struct {
	Report reportinterfaces.ReportNodeRun
}

// IsReportEvent implements ReportEvent interface
func (*ReportError) IsReportEvent() {}
