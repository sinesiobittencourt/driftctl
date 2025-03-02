package aws

import (
	"context"
	"testing"

	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"

	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"

	resourceaws "github.com/cloudskiff/driftctl/pkg/resource/aws"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/cloudskiff/driftctl/pkg/parallel"
	"github.com/stretchr/testify/assert"

	awsdeserializer "github.com/cloudskiff/driftctl/pkg/resource/aws/deserializer"

	"github.com/cloudskiff/driftctl/test/goldenfile"

	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/terraform"
	"github.com/cloudskiff/driftctl/test"
	"github.com/cloudskiff/driftctl/test/mocks"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

func TestEC2EbsSnapshotSupplier_Resources(t *testing.T) {
	tests := []struct {
		test    string
		dirName string
		mock    func(mock *repository.MockEC2Repository)
		err     error
	}{
		{
			test:    "no snapshots",
			dirName: "ec2_ebs_snapshot_empty",
			mock: func(mock *repository.MockEC2Repository) {
				mock.On("ListAllSnapshots").Return([]*ec2.Snapshot{}, nil)
			},
			err: nil,
		},
		{
			test:    "with snapshots",
			dirName: "ec2_ebs_snapshot_multiple",
			mock: func(mock *repository.MockEC2Repository) {
				mock.On("ListAllSnapshots").Return([]*ec2.Snapshot{
					{
						SnapshotId: aws.String("snap-0c509a2a880d95a39"),
					},
					{
						SnapshotId: aws.String("snap-00672558cecd93a61"),
					},
				}, nil)
			},
			err: nil,
		},
		{
			test:    "cannot list snapshots",
			dirName: "ec2_ebs_snapshot_empty",
			mock: func(mock *repository.MockEC2Repository) {
				mock.On("ListAllSnapshots").Return([]*ec2.Snapshot{}, awserr.NewRequestFailure(nil, 403, ""))
			},
			err: remoteerror.NewResourceEnumerationError(awserr.NewRequestFailure(nil, 403, ""), resourceaws.AwsEbsSnapshotResourceType),
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
			supplierLibrary.AddSupplier(NewEC2EbsSnapshotSupplier(provider))
		}

		t.Run(tt.test, func(t *testing.T) {
			provider := mocks.NewMockedGoldenTFProvider(tt.dirName, providerLibrary.Provider(terraform.AWS), shouldUpdate)
			deserializer := awsdeserializer.NewEC2EbsSnapshotDeserializer()
			client := &repository.MockEC2Repository{}
			tt.mock(client)
			s := &EC2EbsSnapshotSupplier{
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
