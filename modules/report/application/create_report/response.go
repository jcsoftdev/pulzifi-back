package createreport

import "github.com/jcsoftdev/pulzifi-back/modules/report/domain/entities"

// Response wraps the created report.
type Response struct {
	Report *entities.Report
}
