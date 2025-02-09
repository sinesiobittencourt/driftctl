package aws

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go/service/kms"

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

func TestKMSAliasSupplier_Resources(t *testing.T) {
	cases := []struct {
		test    string
		dirName string
		mocks   func(client *repository.MockKMSRepository)
		err     error
	}{
		{
			test:    "no aliases",
			dirName: "kms_alias_empty",
			mocks: func(client *repository.MockKMSRepository) {
				client.On("ListAllAliases").Return([]*kms.AliasListEntry{}, nil)
			},
			err: nil,
		},
		{
			test:    "multiple aliases",
			dirName: "kms_alias_multiple",
			mocks: func(client *repository.MockKMSRepository) {
				client.On("ListAllAliases").Return([]*kms.AliasListEntry{
					{AliasName: aws.String("alias/foo")},
					{AliasName: aws.String("alias/bar")},
					{AliasName: aws.String("alias/baz20210225124429210500000001")},
				}, nil)
			},
			err: nil,
		},
		{
			test:    "cannot list aliases",
			dirName: "kms_alias_empty",
			mocks: func(client *repository.MockKMSRepository) {
				client.On("ListAllAliases").Return(nil, awserr.NewRequestFailure(nil, 403, ""))
			},
			err: remoteerror.NewResourceEnumerationError(awserr.NewRequestFailure(nil, 403, ""), resourceaws.AwsKmsAliasResourceType),
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
			supplierLibrary.AddSupplier(NewKMSAliasSupplier(provider))
		}

		t.Run(c.test, func(tt *testing.T) {
			fakeClient := repository.MockKMSRepository{}
			c.mocks(&fakeClient)
			provider := testmocks.NewMockedGoldenTFProvider(c.dirName, providerLibrary.Provider(terraform.AWS), shouldUpdate)
			deserializer := awsdeserializer.NewKMSAliasDeserializer()
			s := &KMSAliasSupplier{
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
