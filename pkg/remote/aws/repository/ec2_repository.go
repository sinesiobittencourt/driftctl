package repository

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
)

type EC2Repository interface {
	ListAllImages() ([]*ec2.Image, error)
	ListAllSnapshots() ([]*ec2.Snapshot, error)
	ListAllVolumes() ([]*ec2.Volume, error)
	ListAllAddresses() ([]*ec2.Address, error)
	ListAllAddressesAssociation() ([]string, error)
	ListAllInstances() ([]*ec2.Instance, error)
	ListAllKeyPairs() ([]*ec2.KeyPairInfo, error)
}

type EC2Client interface {
	ec2iface.EC2API
}

type ec2Repository struct {
	client ec2iface.EC2API
}

func NewEC2Repository(session *session.Session) *ec2Repository {
	return &ec2Repository{
		ec2.New(session),
	}
}

func (r *ec2Repository) ListAllImages() ([]*ec2.Image, error) {
	input := &ec2.DescribeImagesInput{
		Owners: []*string{
			aws.String("self"),
		},
	}
	images, err := r.client.DescribeImages(input)
	if err != nil {
		return nil, err
	}
	return images.Images, err
}

func (r *ec2Repository) ListAllSnapshots() ([]*ec2.Snapshot, error) {
	var snapshots []*ec2.Snapshot
	input := &ec2.DescribeSnapshotsInput{
		OwnerIds: []*string{
			aws.String("self"),
		},
	}
	err := r.client.DescribeSnapshotsPages(input, func(res *ec2.DescribeSnapshotsOutput, lastPage bool) bool {
		snapshots = append(snapshots, res.Snapshots...)
		return !lastPage
	})
	if err != nil {
		return nil, err
	}
	return snapshots, err
}

func (r *ec2Repository) ListAllVolumes() ([]*ec2.Volume, error) {
	var volumes []*ec2.Volume
	input := &ec2.DescribeVolumesInput{}
	err := r.client.DescribeVolumesPages(input, func(res *ec2.DescribeVolumesOutput, lastPage bool) bool {
		volumes = append(volumes, res.Volumes...)
		return !lastPage
	})
	if err != nil {
		return nil, err
	}
	return volumes, nil
}

func (r *ec2Repository) ListAllAddresses() ([]*ec2.Address, error) {
	input := &ec2.DescribeAddressesInput{}
	response, err := r.client.DescribeAddresses(input)
	if err != nil {
		return nil, err
	}
	return response.Addresses, nil
}

func (r *ec2Repository) ListAllAddressesAssociation() ([]string, error) {
	results := make([]string, 0)
	addresses, err := r.ListAllAddresses()
	if err != nil {
		return nil, err
	}
	for _, address := range addresses {
		if address.AssociationId != nil {
			results = append(results, aws.StringValue(address.AssociationId))
		}
	}
	return results, nil
}

func (r *ec2Repository) ListAllInstances() ([]*ec2.Instance, error) {
	var instances []*ec2.Instance
	input := &ec2.DescribeInstancesInput{}
	err := r.client.DescribeInstancesPages(input, func(res *ec2.DescribeInstancesOutput, lastPage bool) bool {
		for _, reservation := range res.Reservations {
			instances = append(instances, reservation.Instances...)
		}
		return !lastPage
	})
	if err != nil {
		return nil, err
	}
	return instances, nil
}

func (r *ec2Repository) ListAllKeyPairs() ([]*ec2.KeyPairInfo, error) {
	input := &ec2.DescribeKeyPairsInput{}
	pairs, err := r.client.DescribeKeyPairs(input)
	if err != nil {
		return nil, err
	}
	return pairs.KeyPairs, err
}
