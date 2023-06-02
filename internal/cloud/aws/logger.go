/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package aws

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/ec2/imds"
	logs "github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/edgelesssys/constellation/v2/internal/cloud"
	"k8s.io/utils/clock"
)

// Logger is a Cloud Logger for AWS.
// Log messages are collected and periodically flushed to AWS Cloudwatch Logs.
type Logger struct {
	api logAPI

	ec2API  ec2API
	imdsAPI imdsAPI

	groupName  string
	streamName string

	logs          []types.InputLogEvent
	sequenceToken *string

	mux      sync.Mutex
	interval time.Duration
	clock    clock.WithTicker
	wg       sync.WaitGroup
	stopCh   chan struct{}
}

// NewLogger creates a new Cloud Logger for AWS.
func NewLogger(ctx context.Context) (*Logger, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithEC2IMDSRegion())
	if err != nil {
		return nil, err
	}
	client := logs.NewFromConfig(cfg)

	l := &Logger{
		api:      client,
		ec2API:   ec2.NewFromConfig(cfg),
		imdsAPI:  imds.NewFromConfig(cfg),
		interval: time.Second,
		clock:    clock.RealClock{},
		wg:       sync.WaitGroup{},
		stopCh:   make(chan struct{}, 1),
	}

	if err := l.createStream(ctx); err != nil {
		return nil, err
	}

	l.flushLoop()

	return l, nil
}

// Disclose adds a message to the log queue.
// The messages are flushed periodically to AWS Cloudwatch Logs.
func (l *Logger) Disclose(msg string) {
	l.mux.Lock()
	defer l.mux.Unlock()
	l.logs = append(l.logs, types.InputLogEvent{
		Message:   aws.String(msg),
		Timestamp: aws.Int64(l.clock.Now().UnixMilli()),
	})
}

// Close flushes the logs a final time and stops the flush loop.
func (l *Logger) Close() error {
	l.stopCh <- struct{}{}
	l.wg.Wait()
	return l.flushLogs()
}

// flushLogs flushes the aggregated log messages to AWS Cloudwatch Logs.
func (l *Logger) flushLogs() error {
	// make sure only one flush operation is running at a time
	l.mux.Lock()
	defer l.mux.Unlock()

	if len(l.logs) == 0 {
		return nil // no logs to flush
	}

	ctx := context.Background()
	logRequest := &logs.PutLogEventsInput{
		LogEvents:     l.logs,
		LogGroupName:  &l.groupName,
		LogStreamName: &l.streamName,
		SequenceToken: l.sequenceToken,
	}

	for res, err := l.api.PutLogEvents(ctx, logRequest); ; res, err = l.api.PutLogEvents(ctx, logRequest) {
		if err == nil {
			l.sequenceToken = res.NextSequenceToken
			l.logs = nil
			return nil
		}
		// If the flush operation was called on a pre-existing stream,
		// or another operation sent logs to the same stream,
		// the sequence token may not be set correctly.
		// We can retrieve the correct sequence token from the error message.
		var sequenceErr *types.InvalidSequenceTokenException
		if !errors.As(err, &sequenceErr) {
			return err
		}
		logRequest.SequenceToken = sequenceErr.ExpectedSequenceToken
	}
}

// flushLoop periodically flushes the logs to AWS Cloudwatch Logs.
func (l *Logger) flushLoop() {
	l.wg.Add(1)
	ticker := l.clock.NewTicker(l.interval)

	go func() {
		defer l.wg.Done()
		defer ticker.Stop()

		for {
			_ = l.flushLogs()
			select {
			case <-ticker.C():
			case <-l.stopCh:
				return
			}
		}
	}()
}

// createStream creates a new log stream in AWS Cloudwatch Logs.
func (l *Logger) createStream(ctx context.Context) error {
	name, uid, err := l.getNameAndUID(ctx)
	if err != nil {
		return err
	}
	l.streamName = name

	// find log group with matching Constellation UID
	describeInput := &logs.DescribeLogGroupsInput{}
	for res, err := l.api.DescribeLogGroups(ctx, describeInput); ; res, err = l.api.DescribeLogGroups(ctx, describeInput) {
		if err != nil {
			return err
		}

		for _, group := range res.LogGroups {
			tags, err := l.api.ListTagsLogGroup(ctx, &logs.ListTagsLogGroupInput{LogGroupName: group.LogGroupName})
			if err != nil {
				continue // we may not have permission to read the tags of a log group outside the Constellation scope
			}
			if tags.Tags[cloud.TagUID] == uid {
				l.groupName = *group.LogGroupName
				res.NextToken = nil // stop pagination
				break
			}
		}
		if res.NextToken == nil {
			break
		}
		describeInput.NextToken = res.NextToken
	}
	if l.groupName == "" {
		return fmt.Errorf("failed to find log group for UID %s", uid)
	}

	// create or use existing log stream
	if _, err := l.api.CreateLogStream(ctx, &logs.CreateLogStreamInput{
		LogGroupName:  &l.groupName,
		LogStreamName: &l.streamName,
	}); err != nil {
		// Ignore error if the stream already exists
		var createErr *types.ResourceAlreadyExistsException
		if !errors.As(err, &createErr) {
			return err
		}
	}

	return nil
}

func (l *Logger) getNameAndUID(ctx context.Context) (string, string, error) {
	identity, err := l.imdsAPI.GetInstanceIdentityDocument(ctx, &imds.GetInstanceIdentityDocumentInput{})
	if err != nil {
		return "", "", fmt.Errorf("retrieving instance identity: %w", err)
	}

	out, err := l.ec2API.DescribeInstances(ctx, &ec2.DescribeInstancesInput{
		InstanceIds: []string{identity.InstanceID},
	})
	if err != nil {
		return "", "", fmt.Errorf("descibing instances: %w", err)
	}

	if len(out.Reservations) != 1 || len(out.Reservations[0].Instances) != 1 {
		return "", "", fmt.Errorf("expected 1 instance, got %d", len(out.Reservations[0].Instances))
	}

	uid, err := findTag(out.Reservations[0].Instances[0].Tags, cloud.TagUID)
	if err != nil {
		return "", "", fmt.Errorf("finding tag %s: %w", cloud.TagUID, err)
	}

	return identity.InstanceID, uid, err
}

type logAPI interface {
	CreateLogStream(context.Context, *logs.CreateLogStreamInput, ...func(*logs.Options)) (*logs.CreateLogStreamOutput, error)
	DescribeLogGroups(context.Context, *logs.DescribeLogGroupsInput, ...func(*logs.Options)) (*logs.DescribeLogGroupsOutput, error)
	ListTagsLogGroup(context.Context, *logs.ListTagsLogGroupInput, ...func(*logs.Options)) (*logs.ListTagsLogGroupOutput, error)
	PutLogEvents(context.Context, *logs.PutLogEventsInput, ...func(*logs.Options)) (*logs.PutLogEventsOutput, error)
}
