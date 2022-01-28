package modconfig

import "github.com/turbot/steampipe/utils"

type ReportChartLabels struct {
	Display *string `cty:"display" hcl:"display" json:"display,omitempty"`
	Format  *string `cty:"format" hcl:"format" json:"format,omitempty"`
}

func (l *ReportChartLabels) Equals(other *ReportChartLabels) bool {
	if other == nil {
		return false
	}

	return utils.SafeStringsEqual(l.Display, other.Display) &&
		utils.SafeStringsEqual(l.Format, other.Format)
}

func (l *ReportChartLabels) Merge(other *ReportChartLabels) {
	if l.Display == nil {
		l.Display = other.Display
	}
	if l.Format == nil {
		l.Format = other.Format
	}
}
