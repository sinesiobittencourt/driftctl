package aws

import (
	"fmt"

	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"
	awsdeserializer "github.com/cloudskiff/driftctl/pkg/resource/aws/deserializer"

	"github.com/cloudskiff/driftctl/pkg/remote/deserializer"
	"github.com/cloudskiff/driftctl/pkg/resource"
	resourceaws "github.com/cloudskiff/driftctl/pkg/resource/aws"
	"github.com/cloudskiff/driftctl/pkg/terraform"

	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
)

type IamUserPolicySupplier struct {
	reader       terraform.ResourceReader
	deserializer deserializer.CTYDeserializer
	client       repository.IAMRepository
	runner       *terraform.ParallelResourceReader
}

func NewIamUserPolicySupplier(provider *AWSTerraformProvider) *IamUserPolicySupplier {
	return &IamUserPolicySupplier{
		provider,
		awsdeserializer.NewIamUserPolicyDeserializer(),
		repository.NewIAMClient(provider.session),
		terraform.NewParallelResourceReader(provider.Runner().SubRunner()),
	}
}

func (s *IamUserPolicySupplier) Resources() ([]resource.Resource, error) {
	users, err := s.client.ListAllUsers()
	if err != nil {
		return nil, remoteerror.NewResourceEnumerationErrorWithType(err, resourceaws.AwsIamUserPolicyResourceType, resourceaws.AwsIamUserResourceType)
	}
	results := make([]cty.Value, 0)
	if len(users) > 0 {
		policies := make([]string, 0)
		for _, user := range users {
			userName := *user.UserName
			policyList, err := s.client.ListAllUserPolicies(userName)
			if err != nil {
				return nil, remoteerror.NewResourceEnumerationError(err, resourceaws.AwsIamUserPolicyResourceType)
			}
			for _, polName := range policyList {
				policies = append(policies, fmt.Sprintf("%s:%s", userName, *polName))
			}
		}

		for _, policy := range policies {
			polName := policy
			s.runner.Run(func() (cty.Value, error) {
				return s.readUserPolicy(polName)
			})
		}
		results, err = s.runner.Wait()
		if err != nil {
			return nil, err
		}
	}
	return s.deserializer.Deserialize(results)
}

func (s *IamUserPolicySupplier) readUserPolicy(policyName string) (cty.Value, error) {
	res, err := s.reader.ReadResource(
		terraform.ReadResourceArgs{
			Ty: resourceaws.AwsIamUserPolicyResourceType,
			ID: policyName,
		},
	)
	if err != nil {
		logrus.Warnf("Error reading iam user policy %s[%s]: %+v", policyName, resourceaws.AwsIamUserResourceType, err)
		return cty.NilVal, err
	}

	return *res, nil
}
