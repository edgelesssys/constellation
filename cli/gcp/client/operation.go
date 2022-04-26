package client

import (
	"context"
	"errors"
	"fmt"

	computepb "google.golang.org/genproto/googleapis/cloud/compute/v1"
)

// waitForOperations waits until every operation in the opIDs slice is
// done or returns the first occurring error.
func (c *Client) waitForOperations(ctx context.Context, ops []Operation) error {
	for _, op := range ops {
		switch {
		case op.Proto() == nil:
			return errors.New("proto of operation is nil")
		case op.Proto().Zone != nil:
			if err := c.waitForZoneOperation(ctx, op); err != nil {
				return err
			}
		case op.Proto().Region != nil:
			if err := c.waitForRegionOperation(ctx, op); err != nil {
				return err
			}
		default:
			if err := c.waitForGlobalOperation(ctx, op); err != nil {
				return err
			}
		}
	}
	return nil
}

func (c *Client) waitForGlobalOperation(ctx context.Context, op Operation) error {
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		waitReq := &computepb.WaitGlobalOperationRequest{
			Operation: *op.Proto().Name,
			Project:   c.project,
		}
		zoneOp, err := c.operationGlobalAPI.Wait(ctx, waitReq)
		if err != nil {
			return fmt.Errorf("unable to wait for the operation: %w", err)
		}
		if *zoneOp.Status.Enum() == computepb.Operation_DONE {
			if opErr := zoneOp.Error; opErr != nil {
				return fmt.Errorf("operation failed: %s", opErr.String())
			}
			return nil
		}
	}
}

func (c *Client) waitForZoneOperation(ctx context.Context, op Operation) error {
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		waitReq := &computepb.WaitZoneOperationRequest{
			Operation: *op.Proto().Name,
			Project:   c.project,
			Zone:      c.zone,
		}
		zoneOp, err := c.operationZoneAPI.Wait(ctx, waitReq)
		if err != nil {
			return fmt.Errorf("unable to wait for the operation: %w", err)
		}
		if *zoneOp.Status.Enum() == computepb.Operation_DONE {
			if opErr := zoneOp.Error; opErr != nil {
				return fmt.Errorf("operation failed: %s", opErr.String())
			}
			return nil
		}
	}
}

func (c *Client) waitForRegionOperation(ctx context.Context, op Operation) error {
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		waitReq := &computepb.WaitRegionOperationRequest{
			Operation: *op.Proto().Name,
			Project:   c.project,
			Region:    c.region,
		}
		regionOp, err := c.operationRegionAPI.Wait(ctx, waitReq)
		if err != nil {
			return fmt.Errorf("unable to wait for the operation: %w", err)
		}
		if *regionOp.Status.Enum() == computepb.Operation_DONE {
			if opErr := regionOp.Error; opErr != nil {
				return fmt.Errorf("operation failed: %s", opErr.String())
			}
			return nil
		}
	}
}
