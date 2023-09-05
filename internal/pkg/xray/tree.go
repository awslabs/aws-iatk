package xray

import (
	"context"
	"fmt"
	"sort"

	"github.com/aws/aws-sdk-go-v2/aws"
	"golang.org/x/exp/maps"
)

func NewTree(ctx context.Context, opts treeOptions, sourceTraceId string) (*Tree, error) {

	// Fetch input source trace.
	traceMap, err := opts.getTraces(ctx, opts.xrayClient, []string{sourceTraceId})

	if err != nil {
		return nil, fmt.Errorf("failed to fetch trace %s with error: %w", sourceTraceId, err)
	}

	trace := traceMap[sourceTraceId]

	if trace == nil {
		return nil, fmt.Errorf("failed to fetch trace %s with error: trace not found", sourceTraceId)
	}

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
	treeRootSegment := trace.Segments[0]

	// Recursively get a map of segment/subsegment ids to the corresponding Segment
	mapSegSubsegsIdToSeg := CreateSegIdtoSegMap(trace.Segments)

	// Insert original trace segments, skip the root segment
	for _, segment := range trace.Segments[1:] {
		if segment.ParentId != nil {
			// ParentId could be pointing to a segmentId or a subsegmentId
			if parentSegment, ok := mapSegSubsegsIdToSeg[*segment.ParentId]; ok {
				InsertSegmentChild(parentSegment, segment)
			} else {
				return nil, fmt.Errorf("found a segment %s with no parent", *segment.Id)
			}
		}
	}

	// Find all the leaf nodes and their complete paths using DFS algorithm
	var singlePath []*Segment
	leafPaths := FindLeafSegmentPaths(treeRootSegment, singlePath)

	return &Tree{
		Root:        treeRootSegment,
		Paths:       leafPaths,
		SourceTrace: trace,
	}, nil
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
