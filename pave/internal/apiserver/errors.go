package apiserver

import "fmt"

var (
	errMissingFields    = fmt.Errorf("requester, namespace, and level are required")
	errApproverRequired = fmt.Errorf("approver is required to decide an access request")
	errInternal         = fmt.Errorf("internal server error")
	errRateLimited      = fmt.Errorf("rate limit exceeded, try again shortly")
)

func errServiceNotFound(name string) error {
	return fmt.Errorf("service %q not found in catalog", name)
}

func errJobNotFound(id string) error {
	return fmt.Errorf("job %q not found", id)
}
