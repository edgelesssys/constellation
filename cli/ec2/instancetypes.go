package ec2

import "github.com/aws/aws-sdk-go-v2/service/ec2/types"

// InstanceTypes defines possible values for the SIZE positional argument.
var InstanceTypes = map[string]types.InstanceType{
	"4xlarge":  types.InstanceTypeC5a4xlarge,
	"8xlarge":  types.InstanceTypeC5a8xlarge,
	"12xlarge": types.InstanceTypeC5a12xlarge,
	"16xlarge": types.InstanceTypeC5a16xlarge,
	"24xlarge": types.InstanceTypeC5a24xlarge,
	// shorthands
	"4xl":  types.InstanceTypeC5a4xlarge,
	"8xl":  types.InstanceTypeC5a8xlarge,
	"12xl": types.InstanceTypeC5a12xlarge,
	"16xl": types.InstanceTypeC5a16xlarge,
	"24xl": types.InstanceTypeC5a24xlarge,
}
