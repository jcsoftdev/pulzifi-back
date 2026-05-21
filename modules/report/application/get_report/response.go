package getreport

import "github.com/jcsoftdev/pulzifi-back/modules/report/domain/entities"

// Response wraps the fetched report.
type Response struct {
	Report *entities.Report
}
