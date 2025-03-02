package aws

import (
	"context"
	"testing"

	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"

	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"

	resourceaws "github.com/cloudskiff/driftctl/pkg/resource/aws"

	"github.com/stretchr/testify/assert"

	"github.com/aws/aws-sdk-go/aws/awserr"

	"github.com/cloudskiff/driftctl/pkg/parallel"

	awsdeserializer "github.com/cloudskiff/driftctl/pkg/resource/aws/deserializer"

	"github.com/cloudskiff/driftctl/test/goldenfile"

	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/terraform"
	"github.com/cloudskiff/driftctl/test"
	"github.com/cloudskiff/driftctl/test/mocks"

	awssdk "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
)

func TestDBInstanceSupplier_Resources(t *testing.T) {

	tests := []struct {
		test    string
		dirName string
		mocks   func(client *repository.MockRDSRepository)
		err     error
	}{
		{
			test:    "no dbs",
			dirName: "db_instance_empty",
			mocks: func(client *repository.MockRDSRepository) {
				client.On("ListAllDBInstances").Return([]*rds.DBInstance{}, nil)
			},
			err: nil,
		},
		{
			test:    "single db",
			dirName: "db_instance_single",
			mocks: func(client *repository.MockRDSRepository) {
				client.On("ListAllDBInstances").Return([]*rds.DBInstance{
					{
						DBInstanceIdentifier: awssdk.String("terraform-20201015115018309600000001"),
					},
				}, nil)
			},
			err: nil,
		},
		{
			test:    "multiples mixed db",
			dirName: "db_instance_multiple",
			mocks: func(client *repository.MockRDSRepository) {
				client.On("ListAllDBInstances").Return([]*rds.DBInstance{
					{
						DBInstanceIdentifier: awssdk.String("terraform-20201015115018309600000001"),
					},
					{
						DBInstanceIdentifier: awssdk.String("database-1"),
					},
				}, nil)
			},
			err: nil,
		},
		{
			test:    "multiples mixed db",
			dirName: "db_instance_multiple",
			mocks: func(client *repository.MockRDSRepository) {
				client.On("ListAllDBInstances").Return([]*rds.DBInstance{
					{
						DBInstanceIdentifier: awssdk.String("terraform-20201015115018309600000001"),
					},
					{
						DBInstanceIdentifier: awssdk.String("database-1"),
					},
				}, nil)
			},
			err: nil,
		},
		{
			test:    "Cannot list db instances",
			dirName: "db_instance_empty",
			mocks: func(client *repository.MockRDSRepository) {
				client.On("ListAllDBInstances").Return([]*rds.DBInstance{}, awserr.NewRequestFailure(nil, 403, ""))
			},
			err: remoteerror.NewResourceEnumerationError(awserr.NewRequestFailure(nil, 403, ""), resourceaws.AwsDbInstanceResourceType),
		},
	}
	for _, tt := range tests {
		shouldUpdate := tt.dirName == *goldenfile.Update

		providerLibrary := terraform.NewProviderLibrary()
		supplierLibrary := resource.NewSupplierLibrary()

		if shouldUpdate {
			provider, err := InitTestAwsProvider(providerLibrary)
			if err != nil {
				t.Fatal(err)
			}
			supplierLibrary.AddSupplier(NewDBInstanceSupplier(provider))
		}

		t.Run(tt.test, func(t *testing.T) {
			provider := mocks.NewMockedGoldenTFProvider(tt.dirName, providerLibrary.Provider(terraform.AWS), shouldUpdate)
			deserializer := awsdeserializer.NewDBInstanceDeserializer()

			client := &repository.MockRDSRepository{}
			tt.mocks(client)
			s := &DBInstanceSupplier{
				provider,
				deserializer,
				client,
				terraform.NewParallelResourceReader(parallel.NewParallelRunner(context.TODO(), 10)),
			}
			got, err := s.Resources()
			assert.Equal(t, tt.err, err)

			test.CtyTestDiff(got, tt.dirName, provider, deserializer, shouldUpdate, t)
		})
	}
}
