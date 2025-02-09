// GENERATED, DO NOT EDIT THIS FILE
package aws

import "github.com/zclconf/go-cty/cty"

const AwsRoute53ZoneResourceType = "aws_route53_zone"

type AwsRoute53Zone struct {
	Comment         *string           `cty:"comment"`
	DelegationSetId *string           `cty:"delegation_set_id"`
	ForceDestroy    *bool             `cty:"force_destroy" diff:"-"`
	Id              string            `cty:"id" computed:"true"`
	Name            *string           `cty:"name"`
	NameServers     []string          `cty:"name_servers" computed:"true"`
	Tags            map[string]string `cty:"tags"`
	ZoneId          *string           `cty:"zone_id" computed:"true"`
	Vpc             *[]struct {
		VpcId     *string `cty:"vpc_id"`
		VpcRegion *string `cty:"vpc_region" computed:"true"`
	} `cty:"vpc"`
	CtyVal *cty.Value `diff:"-"`
}

func (r *AwsRoute53Zone) TerraformId() string {
	return r.Id
}

func (r *AwsRoute53Zone) TerraformType() string {
	return AwsRoute53ZoneResourceType
}

func (r *AwsRoute53Zone) CtyValue() *cty.Value {
	return r.CtyVal
}
