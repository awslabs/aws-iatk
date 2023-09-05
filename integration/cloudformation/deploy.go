package cloudformation

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/stretchr/testify/require"
)

func Deploy(t *testing.T, cfnClient *cloudformation.Client, stackName, templatePath string, capabilities []types.Capability) error {
	absPath, err := filepath.Abs(templatePath)
	if err != nil {
		return fmt.Errorf("template path not valid or not exists: %w", err)
	}
	t.Logf("create stack %q with template %q", stackName, absPath)
	filebytes, err := os.ReadFile(absPath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}
	templateBody := string(filebytes)
	_, err = cfnClient.CreateStack(context.TODO(), &cloudformation.CreateStackInput{
		StackName:    aws.String(stackName),
		TemplateBody: aws.String(templateBody),
		Capabilities: capabilities,
	})
	if err != nil {
		return fmt.Errorf("failed to create stack: %w", err)
	}

	createRetryable := func(
		ctx context.Context,
		params *cloudformation.DescribeStacksInput,
		output *cloudformation.DescribeStacksOutput,
		err error,
	) (bool, error) {
		if output.Stacks != nil {
			for _, stack := range output.Stacks {
				switch stack.StackStatus {
				case types.StackStatusCreateInProgress:
					return true, nil
				case types.StackStatusCreateFailed, types.StackStatusRollbackComplete:
					return false, errors.New(*stack.StackStatusReason)
				case types.StackStatusCreateComplete:
					return false, nil
				default:
					return false, nil
				}
			}
		}
		return false, err
	}

	maxWaitTime := 5 * time.Minute
	waiter := cloudformation.NewStackCreateCompleteWaiter(cfnClient, func(o *cloudformation.StackCreateCompleteWaiterOptions) {
		o.Retryable = createRetryable
	})

	err = waiter.Wait(context.TODO(), &cloudformation.DescribeStacksInput{
		StackName: aws.String(stackName),
	}, *aws.Duration(maxWaitTime))
	require.NoError(t, err, "error while waiting for stack creation")
	t.Logf("completed create stack %q", stackName)
	return nil
}

func Destroy(t *testing.T, cfnClient *cloudformation.Client, stackName string) error {
	t.Logf("destroy stack %q", stackName)

	_, err := cfnClient.DeleteStack(context.TODO(), &cloudformation.DeleteStackInput{
		StackName: aws.String(stackName),
	})
	require.NoError(t, err, "failed to delete stack")

	waiter := cloudformation.NewStackDeleteCompleteWaiter(cfnClient, func(options *cloudformation.StackDeleteCompleteWaiterOptions) {
		options.LogWaitAttempts = false
		options.MaxDelay = 40 * time.Second
	})

	maxWaitTime := 5 * time.Minute
	err = waiter.Wait(context.TODO(), &cloudformation.DescribeStacksInput{
		StackName: aws.String(stackName),
	}, maxWaitTime)
	require.NoError(t, err, "error while waiting for stack deletion")
	t.Logf("completed destroy stack %q", stackName)
	return nil
}
