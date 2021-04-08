package aws

import (
	"context"
	"testing"

	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"

	resourceaws "github.com/cloudskiff/driftctl/pkg/resource/aws"

	"github.com/aws/aws-sdk-go/aws/awserr"

	"github.com/cloudskiff/driftctl/pkg/parallel"

	awsdeserializer "github.com/cloudskiff/driftctl/pkg/resource/aws/deserializer"

	"github.com/aws/aws-sdk-go/aws"

	"github.com/aws/aws-sdk-go/service/iam"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/cloudskiff/driftctl/test/goldenfile"
	mocks2 "github.com/cloudskiff/driftctl/test/mocks"

	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/terraform"
	"github.com/cloudskiff/driftctl/test"
)

func TestIamUserPolicySupplier_Resources(t *testing.T) {

	cases := []struct {
		test    string
		dirName string
		mocks   func(repo *repository.MockIAMRepository)
		err     error
	}{
		{
			test:    "no iam user (no policy)",
			dirName: "iam_user_policy_empty",
			mocks: func(repo *repository.MockIAMRepository) {
				repo.On("ListUsersPages", mock.Anything, mock.Anything).Return(nil)
				repo.On("ListUserPoliciesPages", mock.Anything, mock.Anything).Panic("ListUsersPoliciesPages should not be called when there is no user")
			},
			err: nil,
		},
		{
			test:    "iam multiples users multiple policies",
			dirName: "iam_user_policy_multiple",
			mocks: func(repo *repository.MockIAMRepository) {
				repo.On("ListUsersPages",
					&iam.ListUsersInput{},
					mock.MatchedBy(func(callback func(res *iam.ListUsersOutput, lastPage bool) bool) bool {
						callback(&iam.ListUsersOutput{Users: []*iam.User{
							{
								UserName: aws.String("loadbalancer"),
							},
							{
								UserName: aws.String("loadbalancer2"),
							},
						}}, false)
						callback(&iam.ListUsersOutput{Users: []*iam.User{
							{
								UserName: aws.String("loadbalancer3"),
							},
						}}, true)
						return true
					})).Return(nil).Once()

				shouldSkipfirst := false
				shouldSkipSecond := false
				shouldSkipThird := false

				repo.On("ListUserPoliciesPages",
					&iam.ListUserPoliciesInput{
						UserName: aws.String("loadbalancer"),
					},
					mock.MatchedBy(func(callback func(res *iam.ListUserPoliciesOutput, lastPage bool) bool) bool {
						if shouldSkipfirst {
							return false
						}
						callback(&iam.ListUserPoliciesOutput{PolicyNames: []*string{
							aws.String("test"),
							aws.String("test2"),
							aws.String("test3"),
						}}, false)
						callback(&iam.ListUserPoliciesOutput{PolicyNames: []*string{
							aws.String("test4"),
						}}, true)
						shouldSkipfirst = true
						return true
					})).Return(nil).Once()

				repo.On("ListUserPoliciesPages",
					&iam.ListUserPoliciesInput{
						UserName: aws.String("loadbalancer2"),
					},
					mock.MatchedBy(func(callback func(res *iam.ListUserPoliciesOutput, lastPage bool) bool) bool {
						if shouldSkipSecond {
							return false
						}
						callback(&iam.ListUserPoliciesOutput{PolicyNames: []*string{
							aws.String("test2"),
							aws.String("test22"),
							aws.String("test23"),
						}}, false)
						callback(&iam.ListUserPoliciesOutput{PolicyNames: []*string{
							aws.String("test24"),
						}}, true)
						shouldSkipSecond = true
						return true
					})).Return(nil).Once()

				repo.On("ListUserPoliciesPages",
					&iam.ListUserPoliciesInput{
						UserName: aws.String("loadbalancer3"),
					},
					mock.MatchedBy(func(callback func(res *iam.ListUserPoliciesOutput, lastPage bool) bool) bool {
						if shouldSkipThird {
							return false
						}
						callback(&iam.ListUserPoliciesOutput{PolicyNames: []*string{
							aws.String("test3"),
							aws.String("test32"),
							aws.String("test33"),
						}}, false)
						callback(&iam.ListUserPoliciesOutput{PolicyNames: []*string{
							aws.String("test34"),
						}}, true)
						shouldSkipThird = true
						return true
					})).Return(nil).Once()

			},
			err: nil,
		},
		{
			test:    "cannot list iam user (no policy)",
			dirName: "iam_user_policy_empty",
			mocks: func(repo *repository.MockIAMRepository) {
				repo.On("ListUsersPages", mock.Anything, mock.Anything).Return(awserr.NewRequestFailure(nil, 403, ""))
			},
			err: remoteerror.NewResourceEnumerationErrorWithType(awserr.NewRequestFailure(nil, 403, ""), resourceaws.AwsIamUserPolicyResourceType, resourceaws.AwsIamUserResourceType),
		},

		{
			test:    "cannot list user policy",
			dirName: "iam_user_policy_empty",
			mocks: func(repo *repository.MockIAMRepository) {
				repo.On("ListUsersPages",
					&iam.ListUsersInput{},
					mock.MatchedBy(func(callback func(res *iam.ListUsersOutput, lastPage bool) bool) bool {
						callback(&iam.ListUsersOutput{Users: []*iam.User{
							{
								UserName: aws.String("loadbalancer"),
							},
							{
								UserName: aws.String("loadbalancer2"),
							},
						}}, true)
						return true
					})).Return(nil).Once()
				repo.On("ListUserPoliciesPages", mock.Anything, mock.Anything).Return(awserr.NewRequestFailure(nil, 403, ""))
			},
			err: remoteerror.NewResourceEnumerationError(awserr.NewRequestFailure(nil, 403, ""), resourceaws.AwsIamUserPolicyResourceType),
		},
	}
	for _, c := range cases {
		shouldUpdate := c.dirName == *goldenfile.Update

		providerLibrary := terraform.NewProviderLibrary()
		supplierLibrary := resource.NewSupplierLibrary()

		if shouldUpdate {
			provider, err := InitTestAwsProvider(providerLibrary)
			if err != nil {
				t.Fatal(err)
			}
			supplierLibrary.AddSupplier(NewIamUserPolicySupplier(provider))
		}

		t.Run(c.test, func(tt *testing.T) {
			fakeIam := repository.MockIAMRepository{}
			c.mocks(&fakeIam)

			provider := mocks2.NewMockedGoldenTFProvider(c.dirName, providerLibrary.Provider(terraform.AWS), shouldUpdate)
			deserializer := awsdeserializer.NewIamUserPolicyDeserializer()
			s := &IamUserPolicySupplier{
				provider,
				deserializer,
				&fakeIam,
				terraform.NewParallelResourceReader(parallel.NewParallelRunner(context.TODO(), 10)),
			}
			got, err := s.Resources()
			assert.Equal(tt, c.err, err)

			mock.AssertExpectationsForObjects(tt)
			test.CtyTestDiff(got, c.dirName, provider, deserializer, shouldUpdate, t)
		})
	}
}
