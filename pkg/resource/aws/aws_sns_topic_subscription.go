// GENERATED, DO NOT EDIT THIS FILE
package aws

import "github.com/zclconf/go-cty/cty"

const AwsSnsTopicSubscriptionResourceType = "aws_sns_topic_subscription"

type AwsSnsTopicSubscription struct {
	Arn                          *string    `cty:"arn" computed:"true"`
	ConfirmationTimeoutInMinutes *int       `cty:"confirmation_timeout_in_minutes"`
	DeliveryPolicy               *string    `cty:"delivery_policy" jsonstring:"true"`
	Endpoint                     *string    `cty:"endpoint"`
	EndpointAutoConfirms         *bool      `cty:"endpoint_auto_confirms"`
	FilterPolicy                 *string    `cty:"filter_policy" jsonstring:"true"`
	Id                           string     `cty:"id" computed:"true"`
	Protocol                     *string    `cty:"protocol"`
	RawMessageDelivery           *bool      `cty:"raw_message_delivery"`
	TopicArn                     *string    `cty:"topic_arn"`
	CtyVal                       *cty.Value `diff:"-"`
}

func (r *AwsSnsTopicSubscription) TerraformId() string {
	return r.Id
}

func (r *AwsSnsTopicSubscription) TerraformType() string {
	return AwsSnsTopicSubscriptionResourceType
}

func (r *AwsSnsTopicSubscription) CtyValue() *cty.Value {
	return r.CtyVal
}
