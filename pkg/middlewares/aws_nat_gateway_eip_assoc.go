package middlewares

import (
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
	"github.com/sirupsen/logrus"
)

type AwsNatGatewayEipAssoc struct{}

func NewAwsNatGatewayEipAssoc() AwsNatGatewayEipAssoc {
	return AwsNatGatewayEipAssoc{}
}

// When creating a nat gateway, we associate an EIP to the gateway
// It implies that driftctl read a aws_eip_association resource from remote
// As we cannot use aws_eip_association in terraform to assign an eip to an aws_nat_gateway
// we should remove this association to ensure we do not output noise in unmanaged resources
func (a AwsNatGatewayEipAssoc) Execute(remoteResources, resourcesFromState *[]resource.Resource) error {
	newRemoteResources := make([]resource.Resource, 0, len(*remoteResources))

	for _, remoteResource := range *remoteResources {
		// Ignore all resources other than aws_eip_association
		if remoteResource.TerraformType() != aws.AwsEipAssociationResourceType {
			newRemoteResources = append(newRemoteResources, remoteResource)
			continue
		}

		eipAssoc, _ := remoteResource.(*aws.AwsEipAssociation)
		isAssociatedToNatGateway := false

		// Search for a nat gateway associated with our EIP
		for _, res := range *remoteResources {
			if res.TerraformType() == aws.AwsNatGatewayResourceType {
				gateway, _ := res.(*aws.AwsNatGateway)
				if gateway.AllocationId != nil &&
					eipAssoc.AllocationId != nil &&
					*gateway.AllocationId == *eipAssoc.AllocationId {
					isAssociatedToNatGateway = true
					break
				}
			}
		}

		if isAssociatedToNatGateway {
			logrus.WithFields(logrus.Fields{
				"id":   remoteResource.TerraformId(),
				"type": remoteResource.TerraformType(),
			}).Debug("Ignoring aws_eip_association as it is associated to a nat gateway")
			continue
		}

		newRemoteResources = append(newRemoteResources, eipAssoc)
	}

	*remoteResources = newRemoteResources

	return nil
}
