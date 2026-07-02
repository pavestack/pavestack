package apiserver

import "fmt"

var (
	errMissingFields    = fmt.Errorf("requester, namespace, and level are required")
	errApproverRequired = fmt.Errorf("approver is required to decide an access request")
)

func errServiceNotFound(name string) error {
	return fmt.Errorf("service %q not found in catalog", name)
}

func errJobNotFound(id string) error {
	return fmt.Errorf("job %q not found", id)
}
