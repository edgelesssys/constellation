package ec2

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

// Tag is a ec2 tag. It consits of a key and a value.
type Tag struct {
	Key   string
	Value string
}

// Tags is a set of Tags.
type Tags []Tag

// AWS returns a AWS representation of tags.
func (t Tags) AWS() []types.Tag {
	var awsTags []types.Tag
	for _, tag := range t {
		awsTags = append(awsTags, types.Tag{
			Key:   aws.String(tag.Key),
			Value: aws.String(tag.Value),
		})
	}
	return awsTags
}
