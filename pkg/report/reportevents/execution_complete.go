package reportevents

import "github.com/turbot/steampipe/pkg/report/reportinterfaces"

type ExecutionComplete struct {
	Report reportinterfaces.ReportNodeRun
}

// IsReportEvent implements ReportEvent interface
func (*ExecutionComplete) IsReportEvent() {}
