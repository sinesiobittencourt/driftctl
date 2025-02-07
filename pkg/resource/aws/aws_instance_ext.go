package aws

import "github.com/cloudskiff/driftctl/pkg/resource"

func (r *AwsInstance) NormalizeForState() (resource.Resource, error) {
	if r.RootBlockDevice != nil && len(*r.RootBlockDevice) == 0 {
		r.RootBlockDevice = nil
	}
	if r.EbsBlockDevice != nil && len(*r.EbsBlockDevice) == 0 {
		r.EbsBlockDevice = nil
	}
	return r, nil
}

func (r *AwsInstance) NormalizeForProvider() (resource.Resource, error) {
	if r.RootBlockDevice != nil {
		r.RootBlockDevice = nil
	}
	if r.EbsBlockDevice != nil {
		r.EbsBlockDevice = nil
	}
	return r, nil
}
