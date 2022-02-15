package dashboardevents

import "github.com/turbot/steampipe/dashboard/dashboardinterfaces"

type LeafNodeError struct {
	Node dashboardinterfaces.DashboardNodeRun
}

// IsDashboardEvent implements DashboardEvent interface
func (*LeafNodeError) IsDashboardEvent() {}
