/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package aws

import (
	"context"
	"errors"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	logs "github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
	testclock "k8s.io/utils/clock/testing"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

func TestCreateStream(t *testing.T) {
	someErr := errors.New("failed")

	testCases := map[string]struct {
		imds       *stubIMDS
		logs       *stubLogs
		wantGroup  string
		wantStream string
		wantErr    bool
	}{
		"success new stream minimal": {
			imds: &stubIMDS{
				tags: map[string]string{
					tagName: "test-instance",
					tagUID:  "uid",
				},
			},
			logs: &stubLogs{
				describeRes1: &logs.DescribeLogGroupsOutput{
					LogGroups: []types.LogGroup{
						{LogGroupName: aws.String("test-group")},
					},
				},
				listTags: map[string]map[string]string{"test-group": {tagUID: "uid"}},
			},
			wantStream: "test-instance",
			wantGroup:  "test-group",
		},
		"success one group of many": {
			imds: &stubIMDS{
				tags: map[string]string{
					tagName: "test-instance",
					tagUID:  "uid",
				},
			},
			logs: &stubLogs{
				describeRes1: &logs.DescribeLogGroupsOutput{
					LogGroups: []types.LogGroup{
						{
							LogGroupName: aws.String("random-group"),
						},
						{
							LogGroupName: aws.String("other-group"),
						},
					},
					NextToken: aws.String("next"),
				},
				describeRes2: &logs.DescribeLogGroupsOutput{
					LogGroups: []types.LogGroup{
						{
							LogGroupName: aws.String("another-group"),
						},
						{
							LogGroupName: aws.String("test-group"),
						},
					},
				},
				listTags: map[string]map[string]string{
					"random-group": {
						"some-tag": "random-tag",
					},
					"other-group": {
						tagUID: "other-uid",
					},
					"another-group": {
						"some-tag": "uid",
					},
					"test-group": {
						tagUID: "uid",
					},
				},
			},
			wantStream: "test-instance",
			wantGroup:  "test-group",
		},
		"success stream exists": {
			imds: &stubIMDS{
				tags: map[string]string{
					tagName: "test-instance",
					tagUID:  "uid",
				},
			},
			logs: &stubLogs{
				describeRes1: &logs.DescribeLogGroupsOutput{
					LogGroups: []types.LogGroup{
						{LogGroupName: aws.String("test-group")},
					},
				},
				listTags:  map[string]map[string]string{"test-group": {tagUID: "uid"}},
				createErr: &types.ResourceAlreadyExistsException{},
			},
			wantStream: "test-instance",
			wantGroup:  "test-group",
		},
		"create stream error": {
			imds: &stubIMDS{
				tags: map[string]string{
					tagName: "test-instance",
					tagUID:  "uid",
				},
			},
			logs: &stubLogs{
				describeRes1: &logs.DescribeLogGroupsOutput{
					LogGroups: []types.LogGroup{
						{LogGroupName: aws.String("test-group")},
					},
				},
				listTags:  map[string]map[string]string{"test-group": {tagUID: "uid"}},
				createErr: someErr,
			},
			wantErr: true,
		},
		"missing uid tag": {
			imds: &stubIMDS{
				tags: map[string]string{
					tagName: "test-instance",
				},
			},
			logs: &stubLogs{
				describeRes1: &logs.DescribeLogGroupsOutput{
					LogGroups: []types.LogGroup{
						{LogGroupName: aws.String("test-group")},
					},
				},
				listTags: map[string]map[string]string{"test-group": {tagUID: "uid"}},
			},
			wantErr: true,
		},
		"missing name tag": {
			imds: &stubIMDS{
				tags: map[string]string{
					tagUID: "uid",
				},
			},
			logs: &stubLogs{
				describeRes1: &logs.DescribeLogGroupsOutput{
					LogGroups: []types.LogGroup{
						{LogGroupName: aws.String("test-group")},
					},
				},
				listTags: map[string]map[string]string{"test-group": {tagUID: "uid"}},
			},
			wantErr: true,
		},
		"describe groups error": {
			imds: &stubIMDS{
				tags: map[string]string{
					tagName: "test-instance",
					tagUID:  "uid",
				},
			},
			logs: &stubLogs{
				describeErr: someErr,
				listTags:    map[string]map[string]string{"test-group": {tagUID: "uid"}},
			},
			wantErr: true,
		},
		"no matching groups": {
			imds: &stubIMDS{
				tags: map[string]string{
					tagName: "test-instance",
					tagUID:  "uid",
				},
			},
			logs: &stubLogs{
				describeRes1: &logs.DescribeLogGroupsOutput{},
				listTags:     map[string]map[string]string{"test-group": {tagUID: "uid"}},
			},
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			l := &Logger{
				api: tc.logs,
			}

			err := l.createStream(context.Background(), tc.imds)
			if tc.wantErr {
				assert.Error(err)
				return
			}

			assert.NoError(err)
			assert.Equal(tc.wantGroup, l.groupName)
			assert.Equal(tc.wantStream, l.streamName)
		})
	}
}

func TestLogging(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	logAPI := &stubLogs{}

	l := &Logger{
		api:      logAPI,
		interval: 1 * time.Millisecond,
		clock:    testclock.NewFakeClock(time.Time{}),
	}

	l.Disclose("msg")
	l.Disclose("msg")
	// no logs until we flush to the API
	assert.Len(logAPI.logs, 0)

	// flush
	require.NoError(l.flushLogs())
	assert.Len(logAPI.logs, 2)

	// flushing doesn't do anything if there are no logs
	require.NoError(l.flushLogs())
	assert.Len(logAPI.logs, 2)

	// if we flush with an incorrect sequence token,
	// we should get a new sequence token and retry
	logAPI.logSequenceToken = 15
	l.Disclose("msg")
	require.NoError(l.flushLogs())
	assert.Len(logAPI.logs, 3)

	logAPI.putErr = errors.New("failed")
	l.Disclose("msg")
	assert.Error(l.flushLogs())
}

func TestFlushLoop(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	logAPI := &stubLogs{}
	clock := testclock.NewFakeClock(time.Time{})

	l := &Logger{
		api:      logAPI,
		interval: 1 * time.Second,
		clock:    clock,
		stopCh:   make(chan struct{}, 1),
	}

	l.Disclose("msg")
	l.Disclose("msg")

	l.flushLoop()
	clock.Step(1 * time.Second)
	require.NoError(l.Close())
	assert.Len(logAPI.logs, 2)
}

func TestConcurrency(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	l := &Logger{
		api:      &stubLogs{},
		interval: 1 * time.Second,
		clock:    testclock.NewFakeClock(time.Time{}),
		stopCh:   make(chan struct{}, 1),
	}
	var wg sync.WaitGroup

	wg.Add(100)
	for i := 0; i < 100; i++ {
		go func() {
			defer wg.Done()
			l.Disclose("msg")
		}()
	}

	wg.Wait()
	assert.Len(l.logs, 100)
	require.NoError(l.flushLogs())
	assert.Len(l.logs, 0)

	wg.Add(100)
	for i := 0; i < 100; i++ {
		go func() {
			defer wg.Done()
			l.Disclose("msg")
			require.NoError(l.flushLogs())
		}()
	}
	wg.Wait()
	assert.Len(l.logs, 0)
}

type stubLogs struct {
	createErr        error
	describeErr      error
	describeRes1     *logs.DescribeLogGroupsOutput
	describeRes2     *logs.DescribeLogGroupsOutput
	listTagsErr      error
	listTags         map[string]map[string]string
	putErr           error
	logSequenceToken int
	logs             []types.InputLogEvent
}

func (s *stubLogs) CreateLogStream(context.Context, *logs.CreateLogStreamInput, ...func(*logs.Options)) (*logs.CreateLogStreamOutput, error) {
	return nil, s.createErr
}

func (s *stubLogs) DescribeLogGroups(_ context.Context, in *logs.DescribeLogGroupsInput, _ ...func(*logs.Options)) (*logs.DescribeLogGroupsOutput, error) {
	if in.NextToken == nil {
		return s.describeRes1, s.describeErr
	}
	return s.describeRes2, s.describeErr
}

func (s *stubLogs) ListTagsLogGroup(_ context.Context, in *logs.ListTagsLogGroupInput, _ ...func(*logs.Options)) (*logs.ListTagsLogGroupOutput, error) {
	return &logs.ListTagsLogGroupOutput{Tags: s.listTags[*in.LogGroupName]}, s.listTagsErr
}

func (s *stubLogs) PutLogEvents(_ context.Context, in *logs.PutLogEventsInput, _ ...func(*logs.Options)) (*logs.PutLogEventsOutput, error) {
	if s.putErr != nil {
		return nil, s.putErr
	}
	if in.SequenceToken == nil || *in.SequenceToken == "" {
		in.SequenceToken = aws.String("0")
	}
	gotSeq, err := strconv.Atoi(*in.SequenceToken)
	if err != nil {
		return nil, err
	}
	if gotSeq != s.logSequenceToken {
		return nil, &types.InvalidSequenceTokenException{ExpectedSequenceToken: aws.String(strconv.Itoa(s.logSequenceToken))}
	}

	s.logs = append(s.logs, in.LogEvents...)
	s.logSequenceToken++

	return &logs.PutLogEventsOutput{NextSequenceToken: aws.String(strconv.Itoa(s.logSequenceToken))}, nil
}
