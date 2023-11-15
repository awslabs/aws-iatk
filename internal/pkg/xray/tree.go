package xray

import (
	"context"
	"fmt"
	"sort"

	"github.com/aws/aws-sdk-go-v2/aws"
	"golang.org/x/exp/maps"
)

const MAX_TREE_DEPTH = 5

func NewTree(ctx context.Context, opts treeOptions, sourceTraceId string, fetchLinkedTraces bool) (*Tree, error) {
	// Fetch input source trace.
	traceMap, err := opts.getTraces(ctx, opts.xrayClient, []string{sourceTraceId})

	if err != nil {
		return nil, fmt.Errorf("failed to fetch trace %s with error: %w", sourceTraceId, err)
	}

	trace := traceMap[sourceTraceId]

	if trace == nil {
		return nil, fmt.Errorf("failed to fetch trace %s with error: trace not found", sourceTraceId)
	}

	if len(trace.Segments) == 0 {
		return nil, fmt.Errorf("failed to fetch trace %s with error: no trace segments found", sourceTraceId)
	}
	depth := 0
	tree, err := buildTree(traceMap, trace, fetchLinkedTraces, ctx, opts, &depth)
	if err != nil {
		return nil, fmt.Errorf("failed to build trace tree %s with error: %w", sourceTraceId, err)
	}

	return tree, nil
}

func buildTree(traceMap map[string]*Trace, sourceTrace *Trace, fetchLinkedTraces bool, ctx context.Context, opts treeOptions, depth *int) (*Tree, error) {
	// Sort the segments by starttime before creating a tree
	sort.Slice(sourceTrace.Segments,
		func(i, j int) bool {
			return aws.ToFloat64(sourceTrace.Segments[i].StartTime) < aws.ToFloat64(sourceTrace.Segments[j].StartTime)
		})

	// First segment is the root
	treeRootSegment := sourceTrace.Segments[0]

	// Recursively get a map of segment/subsegment ids to the corresponding Segment
	mapSegSubsegsIdToSeg := CreateSegIdtoSegMap(sourceTrace.Segments)
	linkedTraceToSegment := map[string]*Segment{}
	// Insert original trace segments, skip the root segment
	for _, segment := range sourceTrace.Segments[1:] {
		if segment.ParentId != nil {
			// ParentId could be pointing to a segmentId or a subsegmentId
			if parentSegment, ok := mapSegSubsegsIdToSeg[*segment.ParentId]; ok {
				InsertSegmentChild(parentSegment, segment)
				if fetchLinkedTraces {
					getLinkedTraces(segment, linkedTraceToSegment)
				}
			} else {
				return nil, fmt.Errorf("found a segment %s with no parent", *segment.Id)
			}
		}
	}

	// Find all the leaf nodes and their complete paths using DFS algorithm
	var singlePath []*Segment

	linkedTraceIds := maps.Keys(linkedTraceToSegment)
	linkedTraceLimitExceeded := false

	//get all linked traces in one call
	if fetchLinkedTraces && len(linkedTraceIds) > 0 && *depth < MAX_TREE_DEPTH {
		linkedTraceMap, err := opts.getTraces(ctx, opts.xrayClient, linkedTraceIds)

		if err != nil {
			return nil, fmt.Errorf("failed to fetch linked traces %s with error: %w", linkedTraceIds, err)
		}
		*depth = *depth + 1
		for linkedTraceId, linkedTrace := range linkedTraceMap {
			linkedTree, err := buildTree(linkedTraceMap, linkedTrace, fetchLinkedTraces, ctx, opts, depth)
			if err != nil {
				return nil, fmt.Errorf("failed to build trace tree %s with error: %w", linkedTraceId, err)
			}

			parentSegment := linkedTraceToSegment[*linkedTree.Root.TraceId]
			parentSegment.children = append(parentSegment.children, linkedTree.Root)
		}
	}
	if *depth >= MAX_TREE_DEPTH {
		linkedTraceLimitExceeded = true
	}
	leafPaths := FindLeafSegmentPaths(treeRootSegment, singlePath)

	return &Tree{
		Root:                     treeRootSegment,
		Paths:                    leafPaths,
		SourceTrace:              sourceTrace,
		LinkedTraceLimitExceeded: linkedTraceLimitExceeded,
	}, nil
}

// check if there are any linked traces associated with the segment and add them to the map
func getLinkedTraces(segment *Segment, linksToSegMap map[string]*Segment) {
	linkedTraces := segment.Links
	for _, trace := range linkedTraces {
		if *trace.Attributes.ReferenceType == "child" {
			linksToSegMap[*trace.TraceId] = segment
		}
	}

	for _, subseg := range segment.Subsegments {
		for _, trace := range subseg.Links {
			if *trace.Attributes.ReferenceType == "child" {
				linksToSegMap[*trace.TraceId] = segment
			}
		}
	}

}

// Add the provided child segment to the parentSegment's children
func InsertSegmentChild(parentSegment *Segment, childSegment *Segment) {
	parentSegment.children = append(parentSegment.children, childSegment)
}

// Recursive function that creates a map of Segment/Subsegment id to the corresponding Segment
func CreateSubSegIdtoSegMap(subSegments []*Subsegment, segment *Segment) map[string]*Segment {
	mapSegSubSegs := map[string]*Segment{}
	for _, subSegment := range subSegments {
		mapSegSubSegs[*subSegment.Id] = segment
		if subSegment.Subsegments != nil {
			maps.Copy(mapSegSubSegs, CreateSubSegIdtoSegMap(subSegment.Subsegments, segment))
		}
	}
	return mapSegSubSegs
}

// Creates a map of Segment/Subsegment id to the corresponding Segment
func CreateSegIdtoSegMap(segments []*Segment) map[string]*Segment {
	mapSegSubsegsToSeg := map[string]*Segment{}
	for _, segment := range segments {
		mapSegSubsegsToSeg[*segment.Id] = segment
		mapSubSegsToSeg := CreateSubSegIdtoSegMap(segment.Subsegments, segment)
		maps.Copy(mapSegSubsegsToSeg, mapSubSegsToSeg)

	}
	return mapSegSubsegsToSeg
}

func FindLeafSegmentPaths(treeRootSegment *Segment, path []*Segment) [][]*Segment {
	path = append(path, treeRootSegment)

	// If treeRootSegment is a leaf, return a slice with the leaf path
	if len(treeRootSegment.children) == 0 {
		var leafPaths [][]*Segment
		leafPaths = append(leafPaths, path)
		return leafPaths
	}

	var leafPaths [][]*Segment
	for _, childSegmentNode := range treeRootSegment.children {
		// Create a copy of path so that all Nodes don't get added to the same path
		path_copy := make([]*Segment, len(path))
		copy(path_copy, path)
		childLeafPaths := FindLeafSegmentPaths(childSegmentNode, path_copy)
		leafPaths = append(leafPaths, childLeafPaths...)
	}
	return leafPaths
}
