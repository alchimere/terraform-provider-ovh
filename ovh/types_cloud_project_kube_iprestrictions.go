package ovh

import "fmt"

type CloudProjectKubeIpRestrictionsCreateOrUpdateOpts struct {
	Ips []string `json:"ips"`
}

func (s *CloudProjectKubeIpRestrictionsCreateOrUpdateOpts) String() string {
	return fmt.Sprintf("%s", s.Ips)
}
