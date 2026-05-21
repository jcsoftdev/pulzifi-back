package listreports

import "github.com/jcsoftdev/pulzifi-back/modules/report/domain/entities"

// Response wraps the list of reports.
type Response struct {
	Reports []*entities.Report
	Count   int
}
