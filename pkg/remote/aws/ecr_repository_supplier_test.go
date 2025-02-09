package aws

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go/service/ecr"

	"github.com/aws/aws-sdk-go/aws"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/cloudskiff/driftctl/pkg/parallel"
	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"
	"github.com/cloudskiff/driftctl/test"
	"github.com/cloudskiff/driftctl/test/goldenfile"
	testmocks "github.com/cloudskiff/driftctl/test/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/cloudskiff/driftctl/pkg/resource"
	resourceaws "github.com/cloudskiff/driftctl/pkg/resource/aws"
	awsdeserializer "github.com/cloudskiff/driftctl/pkg/resource/aws/deserializer"
	"github.com/cloudskiff/driftctl/pkg/terraform"
)

func TestEcrRepositorySupplier_Resources(t *testing.T) {
	cases := []struct {
		test    string
		dirName string
		mocks   func(client *repository.MockECRRepository)
		err     error
	}{
		{
			test:    "no repository",
			dirName: "ecr_repository_empty",
			mocks: func(client *repository.MockECRRepository) {
				client.On("ListAllRepositories").Return([]*ecr.Repository{}, nil)
			},
			err: nil,
		},
		{
			test:    "multiple repositories",
			dirName: "ecr_repository_multiple",
			mocks: func(client *repository.MockECRRepository) {
				client.On("ListAllRepositories").Return([]*ecr.Repository{
					{RepositoryName: aws.String("test_ecr")},
					{RepositoryName: aws.String("bar")},
				}, nil)
			},
			err: nil,
		},
		{
			test:    "cannot list repository",
			dirName: "ecr_repository_empty",
			mocks: func(client *repository.MockECRRepository) {
				client.On("ListAllRepositories").Return(nil, awserr.NewRequestFailure(nil, 403, ""))
			},
			err: remoteerror.NewResourceEnumerationError(awserr.NewRequestFailure(nil, 403, ""), resourceaws.AwsEcrRepositoryResourceType),
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
			supplierLibrary.AddSupplier(NewECRRepositorySupplier(provider))
		}

		t.Run(c.test, func(tt *testing.T) {
			fakeClient := repository.MockECRRepository{}
			c.mocks(&fakeClient)
			provider := testmocks.NewMockedGoldenTFProvider(c.dirName, providerLibrary.Provider(terraform.AWS), shouldUpdate)
			deserializer := awsdeserializer.NewECRRepositoryDeserializer()
			s := &ECRRepositorySupplier{
				provider,
				deserializer,
				&fakeClient,
				terraform.NewParallelResourceReader(parallel.NewParallelRunner(context.TODO(), 10)),
			}
			got, err := s.Resources()
			assert.Equal(tt, c.err, err)
			mock.AssertExpectationsForObjects(tt)
			test.CtyTestDiff(got, c.dirName, provider, deserializer, shouldUpdate, tt)
		})
	}
}
