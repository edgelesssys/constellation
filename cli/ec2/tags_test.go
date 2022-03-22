package ec2

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/stretchr/testify/assert"
)

func TestTagsAws(t *testing.T) {
	assert := assert.New(t)

	testTags := Tags{
		{
			Key:   "Name",
			Value: "Test",
		},
		{
			Key:   "Foo",
			Value: "Bar",
		},
	}
	expected := []types.Tag{
		{
			Key:   aws.String("Name"),
			Value: aws.String("Test"),
		},
		{
			Key:   aws.String("Foo"),
			Value: aws.String("Bar"),
		},
	}

	awsTags := testTags.AWS()
	assert.Equal(expected, awsTags)
}
