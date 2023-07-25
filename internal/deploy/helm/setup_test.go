package helm

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	"github.com/edgelesssys/constellation/v2/internal/cloud/aws"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test(t *testing.T) {
	c, err := aws.NewLBClient(context.Background())
	require.NoError(t, err)
	assert.NotNil(t, c)
	res, err := c.DescribeLoadBalancers(context.Background(), &elasticloadbalancingv2.DescribeLoadBalancersInput{})
	require.NoError(t, err)

	assert.Nil(t, res)
}
