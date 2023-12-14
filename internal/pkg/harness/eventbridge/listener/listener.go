// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package listener

import (
	"context"
	"errors"
	"fmt"
	"iatk/internal/pkg/harness"
	"iatk/internal/pkg/harness/resource/eventrule"
	"iatk/internal/pkg/slice"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/rs/xid"
)

const (
	TestHarnessType = "EventBridge.Listener"
	IDPrefix        = "iatk_eb_"
)

// Creates a Listener for an event bus resource
func New(ctx context.Context, eventBusName, targetId, ruleName string, tags map[string]string, opts Options) (*Listener, error) {
	// validate if the event bus exists
	eb, err := opts.getEventBus(ctx, opts.ebClient, eventBusName)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource group: %v", err)
	}

	r, err := opts.getRule(ctx, opts.ebClient, ruleName, eventBusName)
	if err != nil {
		return nil, fmt.Errorf("RuleName %q was provided but not found for eventbus %q failed: %v", ruleName, eventBusName, err)
	}

	target, err := opts.listTargetsByRule(ctx, opts.ebClient, targetId, ruleName, eventBusName)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource group: %v", err)
	}

	return &Listener{
		id:           xid.New().String(),
		eventPattern: r.EventPattern,
		target:       target,
		customTags:   tags,
		eventBus:     eb,
		opts:         opts,
	}, nil
}

func isValidID(id string) bool {
	if len(id) != len(IDPrefix)+len(xid.New().String()) {
		return false
	}

	if id[:len(IDPrefix)] != IDPrefix {
		return false
	}

	if _, err := xid.FromString(id[len(IDPrefix):]); err != nil {
		return false
	}

	return true
}

// Gets an existing Listener
func Get(ctx context.Context, id string, opts Options) (*Listener, error) {
	// validate if a resource group is up already
	if !isValidID(id) {
		return nil, errors.New("invalid ID")
	}
	suffix := id[len(IDPrefix):]

	qname := queueName(id)
	q, err := opts.getQueueWithName(ctx, opts.sqsClient, qname)
	if err != nil {
		log.Printf("cannot locate queue %q of eb listener %q: %v", qname, id, err)
		return nil, fmt.Errorf("faied to get eb listener %v: %w", id, err)
	}
	log.Printf("found queue %q", q.ARN)

	eventBusName, err := opts.getEventBusNameFromQueue(ctx, opts.sqsClient, q.QueueURL)
	if err != nil {
		log.Printf("unable to find event bus name of listener %v: %v", id, err)
	}

	var r *eventrule.Rule
	if eventBusName != "" {
		rname := ruleName(id)
		r, err = opts.getRule(ctx, opts.ebClient, rname, eventBusName)
		if err != nil {
			log.Printf("cannot locate rule %q of eb listener %q: %v", rname, id, err)
			log.Printf("skipping destroy rule %q", rname)
		} else {
			log.Printf("found rule %q", r.ARN)
		}
	}

	return &Listener{
		id:    suffix,
		queue: q,
		rule:  r,
		opts:  opts,
	}, nil
}

func queueName(listenerID string) string {
	return listenerID // https://aws.amazon.com/sqs/faqs/#Limits_and_restrictions
}

func ruleName(listenerID string) string {
	return listenerID // https://docs.aws.amazon.com/eventbridge/latest/APIReference/API_Rule.html
}

// Create deploys a Listener and rollback if failed
func Create(ctx context.Context, lr deployer) (*Output, error) {
	log.Printf("creating eb listener %v", lr.ID())
	errDeploy := lr.Deploy(ctx)
	if errDeploy != nil {
		log.Printf("create failed: %v", errDeploy)
		log.Printf("rolling back")
		if err := lr.Destroy(ctx); err != nil {
			log.Printf("rollback failed: %v", err)
			log.Printf(`please manually delete following resources: %v`, arns(lr.Components()))
		}
		return nil, fmt.Errorf("failed to create eb listener %v: %w", lr.ID(), errDeploy)
	}
	log.Printf("created eb listener %v", lr.ID())
	out := lr.JSON()
	return &out, nil
}

func DestroyMultiple(ctx context.Context, ids []string, cfg aws.Config, opts destroyMultipleOptions) error {
	distincts := slice.Dedup(ids)
	errs := []errDestroySingle{}

	// TODO(hawflau): use goroutines
	log.Printf("collecting eb listeners from provided ids: %v", ids)
	listeners := []*Listener{}
	for _, id := range distincts {
		lr, err := opts.Get(ctx, id, NewOptions(cfg))
		if err != nil {
			// TODO (hawflau): potentially an option to fail here if any Resource Group ID is not valid/supported.
			log.Print(err.Error())
			errs = append(errs, errDestroySingle{id, err})
		} else {
			listeners = append(listeners, lr)
		}
	}

	for _, lr := range listeners {
		err := opts.destroySingle(ctx, lr)
		if err != nil {
			log.Print(err.Error())
			errs = append(errs, errDestroySingle{lr.ID(), err})
		}
	}

	if len(errs) > 0 {
		var reasons string
		for i, e := range errs {
			reasons += fmt.Sprint(e.String())
			if i != len(errs)-1 {
				reasons += ", "
			}
		}
		return fmt.Errorf("failed to destroy following listener(s): %v", reasons)
	}

	return nil

}

func destroySingle(ctx context.Context, lr destroyer) error {
	log.Printf("destroying eb listener %q", lr.ID())

	if err := lr.Destroy(ctx); err != nil {
		log.Printf("destroy failed: %v", err)
		log.Printf("please manually delete following resources: %v", arns(lr.Components()))
		return fmt.Errorf("failed to destroy eb listener %q: %w", lr.ID(), err)
	}
	log.Printf("destroy success for eb listener %q", lr.ID())
	return nil
}

func arns(resources []harness.Resource) []string {
	l := []string{}
	for _, r := range resources {
		l = append(l, r.ARN)
	}
	return l
}

type destroyMultipleOptions struct {
	maxConcurrency int

	// funcs
	destroySingle destroySingleFunc
	Get           GetFunc
}

func NewDestroyOptions() destroyMultipleOptions {
	return destroyMultipleOptions{
		maxConcurrency: 5,
		destroySingle:  destroySingle,
		Get:            Get,
	}
}

type errDestroySingle struct {
	resourceGroupID string
	err             error
}

func (e errDestroySingle) String() string {
	return fmt.Sprintf("{resource group id: %v, reason: %v}", e.resourceGroupID, e.err)
}

//go:generate mockery --name deployer
type deployer interface {
	Deploy(ctx context.Context) error
	JSON() Output
	destroyer
}

//go:generate mockery --name destroyer
type destroyer interface {
	Destroy(ctx context.Context) error
	ID() string
	Components() []harness.Resource
}

//go:generate mockery --name GetFunc
type GetFunc func(ctx context.Context, id string, opts Options) (*Listener, error)

//go:generate mockery --name destroySingleFunc
type destroySingleFunc func(ctx context.Context, lr destroyer) error

//go:generate mockery --name poller
type poller interface {
	ReceiveEvents(ctx context.Context, waitTimeSeconds, maxNumberOfMessages int32) ([]Event, error)
	DeleteEvents(ctx context.Context, receiptHandles []string) error
}

func PollEvents(ctx context.Context, lr poller, waitTimeSeconds, maxNumberOfMessages int32) ([]string, error) {
	events, err := lr.ReceiveEvents(ctx, waitTimeSeconds, maxNumberOfMessages)
	if err != nil {
		return nil, fmt.Errorf("failed to poll events: %w", err)
	}
	handles := make([]string, 0, len(events))
	ret := make([]string, 0, len(events))
	for _, e := range events {
		handles = append(handles, e.ReceiptHandle)
		ret = append(ret, e.Body)
	}
	if len(handles) > 0 {
		err = lr.DeleteEvents(ctx, handles)
		if err != nil {
			return nil, fmt.Errorf("failed to poll events: %w", err)
		}
	}
	return ret, nil
}
