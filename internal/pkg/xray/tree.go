package xray

import (
	"context"
	"fmt"
	"sort"

	"github.com/aws/aws-sdk-go-v2/aws"
)

func NewTree(ctx context.Context, opts treeOptions, sourceTraceId string, fetchChildLinkedTraces bool) (*Tree, error) {

	// Fetch input source trace.
	traceMap, err := opts.getTraces(ctx, opts.xrayClient, []string{sourceTraceId})

	if err != nil {
		return nil, fmt.Errorf("failed to fetch trace %s with error: %v", sourceTraceId, err)
	}

	trace := traceMap[sourceTraceId]

	// If the trace returns 0 segments, raise an error
	if len(trace.Segments) == 0 {
		return nil, fmt.Errorf("failed to fetch trace %s with error: no trace segments found", sourceTraceId)
	}

	// Sort the segments by starttime before creating a tree
	sort.Slice(trace.Segments,
		func(i, j int) bool {
			return aws.ToFloat64(trace.Segments[i].StartTime) < aws.ToFloat64(trace.Segments[j].StartTime)
		})

	// First segment is the root
	traceTreeRoot := InsertTraceTreeNode(nil, trace.Segments[0])

	// Insert original trace segments, skip the root segment
	for _, segment := range trace.Segments[1:] {
		treeNodeInserted := InsertTraceTreeNode(traceTreeRoot, segment)

		if treeNodeInserted == nil {
			return nil, fmt.Errorf("found a segment %s with no parent", aws.ToString(segment.Id))
		}
	}

	// TODO: Create a function to recursively find all child traces from subsegments if fetchChildLinkedTraces = true

	// TODO: Insert child traces segments into the tree if fetchChildLinkedTraces = true

	// Find all the leaf nodes and their complete paths using DFS algorithm
	var singlePath []*Segment
	leafPaths := FindLeafSegmentPaths(traceTreeRoot, singlePath)

	return &Tree{
		Root:        traceTreeRoot.SegmentObject,
		Paths:       leafPaths,
		SourceTrace: trace,
	}, nil
}

func InsertTraceTreeNode(rootTreeNode *TraceTreeNode, newSegment *Segment) *TraceTreeNode {

	// If the root Segment is provided, create and return a TraceTreeNode
	if newSegment.ParentId == nil {
		return &TraceTreeNode{
			SegmentObject: newSegment,
		}
	}

	if aws.ToString(rootTreeNode.SegmentObject.Id) == aws.ToString(newSegment.ParentId) {
		newTreeNode := TraceTreeNode{
			SegmentObject: newSegment,
		}
		rootTreeNode.Children = append(rootTreeNode.Children, &newTreeNode)
		return &newTreeNode
	}

	for _, childTreeNode := range rootTreeNode.Children {
		// If a node under this child inserted the new segment, return the node
		insertedGrandChild := InsertTraceTreeNode(childTreeNode, newSegment)
		if insertedGrandChild != nil {
			return insertedGrandChild
		}
	}

	// If none of the Segment Nodes are parents of the new segment, return nil
	return nil
}

func FindLeafSegmentPaths(rootTreeNode *TraceTreeNode, path []*Segment) [][]*Segment {
	path = append(path, rootTreeNode.SegmentObject)

	// If rootTreeNode is a leaf, return a slice with the leaf path
	if len(rootTreeNode.Children) == 0 {
		var leafPaths [][]*Segment
		leafPaths = append(leafPaths, path)
		return leafPaths
	}

	var leafPaths [][]*Segment
	for _, childTreeNode := range rootTreeNode.Children {
		// Create a copy of path so that all Nodes don't get added to the same path
		path_copy := make([]*Segment, len(path))
		copy(path_copy, path)
		childLeafPaths := FindLeafSegmentPaths(childTreeNode, path_copy)
		leafPaths = append(leafPaths, childLeafPaths...)
	}
	return leafPaths
}
