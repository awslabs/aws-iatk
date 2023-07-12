// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package harness

type Resource struct {
	Type       string `json:"Type"`
	PhysicalID string `json:"PhysicalID"`
	ARN        string `json:"ARN"`
}
