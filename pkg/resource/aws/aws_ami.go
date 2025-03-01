// GENERATED, DO NOT EDIT THIS FILE
package aws

import "github.com/zclconf/go-cty/cty"

const AwsAmiResourceType = "aws_ami"

type AwsAmi struct {
	Architecture       *string           `cty:"architecture"`
	Arn                *string           `cty:"arn" computed:"true"`
	Description        *string           `cty:"description"`
	EnaSupport         *bool             `cty:"ena_support"`
	Id                 string            `cty:"id" computed:"true"`
	ImageLocation      *string           `cty:"image_location" computed:"true"`
	KernelId           *string           `cty:"kernel_id"`
	ManageEbsSnapshots *bool             `cty:"manage_ebs_snapshots" computed:"true"`
	Name               *string           `cty:"name"`
	RamdiskId          *string           `cty:"ramdisk_id"`
	RootDeviceName     *string           `cty:"root_device_name"`
	RootSnapshotId     *string           `cty:"root_snapshot_id" computed:"true"`
	SriovNetSupport    *string           `cty:"sriov_net_support"`
	Tags               map[string]string `cty:"tags"`
	VirtualizationType *string           `cty:"virtualization_type"`
	EbsBlockDevice     *[]struct {
		DeleteOnTermination *bool   `cty:"delete_on_termination"`
		DeviceName          *string `cty:"device_name"`
		Encrypted           *bool   `cty:"encrypted"`
		Iops                *int    `cty:"iops"`
		SnapshotId          *string `cty:"snapshot_id"`
		VolumeSize          *int    `cty:"volume_size" computed:"true"`
		VolumeType          *string `cty:"volume_type"`
	} `cty:"ebs_block_device"`
	EphemeralBlockDevice *[]struct {
		DeviceName  *string `cty:"device_name"`
		VirtualName *string `cty:"virtual_name"`
	} `cty:"ephemeral_block_device"`
	Timeouts *struct {
		Create *string `cty:"create"`
		Delete *string `cty:"delete"`
		Update *string `cty:"update"`
	} `cty:"timeouts" diff:"-"`
	CtyVal *cty.Value `diff:"-"`
}

func (r *AwsAmi) TerraformId() string {
	return r.Id
}

func (r *AwsAmi) TerraformType() string {
	return AwsAmiResourceType
}

func (r *AwsAmi) CtyValue() *cty.Value {
	return r.CtyVal
}
