package aws

import (
	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"
	"github.com/cloudskiff/driftctl/pkg/remote/deserializer"
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"
	"github.com/cloudskiff/driftctl/pkg/resource"
	resourceaws "github.com/cloudskiff/driftctl/pkg/resource/aws"
	awsdeserializer "github.com/cloudskiff/driftctl/pkg/resource/aws/deserializer"
	"github.com/cloudskiff/driftctl/pkg/terraform"

	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
)

type IamRolePolicyAttachmentSupplier struct {
	reader       terraform.ResourceReader
	deserializer deserializer.CTYDeserializer
	client       repository.IAMRepository
	runner       *terraform.ParallelResourceReader
}

func NewIamRolePolicyAttachmentSupplier(provider *AWSTerraformProvider) *IamRolePolicyAttachmentSupplier {
	return &IamRolePolicyAttachmentSupplier{
		provider,
		awsdeserializer.NewIamRolePolicyAttachmentDeserializer(),
		repository.NewIAMClient(provider.session),
		terraform.NewParallelResourceReader(provider.Runner().SubRunner()),
	}
}

func (s *IamRolePolicyAttachmentSupplier) Resources() ([]resource.Resource, error) {
	roles, err := s.client.ListAllRoles()
	if err != nil {
		return nil, remoteerror.NewResourceEnumerationErrorWithType(err, resourceaws.AwsIamRolePolicyAttachmentResourceType, resourceaws.AwsIamUserResourceType)
	}
	results := make([]cty.Value, 0)
	if len(roles) > 0 {
		attachedPolicies := make([]*repository.AttachedRolePolicy, 0)
		for _, role := range roles {
			roleName := *role.RoleName
			if awsIamRoleShouldBeIgnored(roleName) {
				continue
			}
			roleAttachmentList, err := s.client.ListAllRolePolicyAttachments(resourceaws.AwsIamRolePolicyAttachmentResourceType)
			if err != nil {
				return nil, remoteerror.NewResourceEnumerationErrorWithType(err, resourceaws.AwsIamRolePolicyAttachmentResourceType, resourceaws.AwsIamRolePolicyResourceType)
			}
			attachedPolicies = append(attachedPolicies, roleAttachmentList...)
		}

		for _, attachedPolicy := range attachedPolicies {
			attached := *attachedPolicy
			s.runner.Run(func() (cty.Value, error) {
				return s.readRolePolicyAttachment(attached)
			})
		}
		results, err = s.runner.Wait()
		if err != nil {
			return nil, err
		}
	}

	return s.deserializer.Deserialize(results)
}

func (s *IamRolePolicyAttachmentSupplier) readRolePolicyAttachment(attachedPol repository.AttachedRolePolicy) (cty.Value, error) {
	res, err := s.reader.ReadResource(
		terraform.ReadResourceArgs{
			Ty: resourceaws.AwsIamRolePolicyAttachmentResourceType,
			ID: *attachedPol.PolicyName,
			Attributes: map[string]string{
				"role":       attachedPol.RoleName,
				"policy_arn": *attachedPol.PolicyArn,
			},
		},
	)

	if err != nil {
		logrus.Warnf("Error reading iam role policy attachment %s[%s]: %+v", attachedPol, resourceaws.AwsIamRolePolicyAttachmentResourceType, err)
		return cty.NilVal, err
	}
	return *res, nil
}
