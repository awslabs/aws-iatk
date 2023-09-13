package xray

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTree(t *testing.T) {
	cases := map[string]struct {
		mockGetTraces func(ctx context.Context, api BatchGetTracesAPI, traceIds []string) *mockGetTracesFunc
		sourceTraceId string
		rootSegment   string
		expectErr     error
	}{
		"success": {
			sourceTraceId: "1-64de5a99-5d09aa705e56bbd0152548cb",
			rootSegment:   "segment1-id",
			mockGetTraces: func(ctx context.Context, api BatchGetTracesAPI, traceIds []string) *mockGetTracesFunc {
				f := newMockGetTracesFunc(t)
				mockSegmentId1 := "segment1-id"
				mockSegmentDuration1 := float64(1.692291184114e+09)
				mockSegmentId2 := "segment2-id"
				mockSegmentDuration2 := float64(1.692291184757e+09)
				f.EXPECT().Execute(ctx, api, traceIds).
					Return(
						map[string]*Trace{
							traceIds[0]: &Trace{
								Id: &traceIds[0],
								Segments: []*Segment{
									{Id: &mockSegmentId1, StartTime: &mockSegmentDuration1},
									{Id: &mockSegmentId2, StartTime: &mockSegmentDuration2, ParentId: &mockSegmentId1},
								},
							},
						},
						nil,
					)
				return f
			},
			expectErr: nil,
		},
		"success with subsegments as parent-ids": {
			sourceTraceId: "1-64de5a99-5d09aa705e56bbd0152548cb",
			rootSegment:   "segment1-id",
			mockGetTraces: func(ctx context.Context, api BatchGetTracesAPI, traceIds []string) *mockGetTracesFunc {
				f := newMockGetTracesFunc(t)
				mockSegmentId1 := "segment1-id"
				mockSegmentDuration1 := float64(1.692291184114e+09)
				mockSegmentId2 := "segment2-id"
				mockSubSegmentId1 := "sub-segment1-id"
				mockSegmentDuration2 := float64(1.692291184757e+09)
				f.EXPECT().Execute(ctx, api, traceIds).
					Return(
						map[string]*Trace{
							traceIds[0]: &Trace{
								Id: &traceIds[0],
								Segments: []*Segment{
									{Id: &mockSegmentId1, StartTime: &mockSegmentDuration1, Subsegments: []*Subsegment{
										&Subsegment{Id: &mockSubSegmentId1},
									}},
									{Id: &mockSegmentId2, StartTime: &mockSegmentDuration2, ParentId: &mockSubSegmentId1},
								},
							},
						},
						nil,
					)
				return f
			},
			expectErr: nil,
		},
		"no segments found": {
			sourceTraceId: "1-64de5a99-5d09aa705e56bbd0152548cb",
			mockGetTraces: func(ctx context.Context, api BatchGetTracesAPI, traceIds []string) *mockGetTracesFunc {
				f := newMockGetTracesFunc(t)
				f.EXPECT().Execute(ctx, api, traceIds).
					Return(
						map[string]*Trace{
							traceIds[0]: &Trace{
								Id:       &traceIds[0],
								Segments: []*Segment{},
							},
						},
						nil,
					)
				return f
			},
			expectErr: errors.New("failed to fetch trace 1-64de5a99-5d09aa705e56bbd0152548cb with error: no trace segments found"),
		},
		"input trace not found": {
			sourceTraceId: "1-64de5a99-5d09aa705e56bbd0152548cb",
			mockGetTraces: func(ctx context.Context, api BatchGetTracesAPI, traceIds []string) *mockGetTracesFunc {
				f := newMockGetTracesFunc(t)
				f.EXPECT().Execute(ctx, api, traceIds).
					Return(
						map[string]*Trace{},
						nil,
					)
				return f
			},
			expectErr: errors.New("failed to fetch trace 1-64de5a99-5d09aa705e56bbd0152548cb with error: trace not found"),
		},
		"get traces api failed": {
			sourceTraceId: "1-64de5a99-5d09aa705e56bbd0152548cb",
			mockGetTraces: func(ctx context.Context, api BatchGetTracesAPI, traceIds []string) *mockGetTracesFunc {
				f := newMockGetTracesFunc(t)
				f.EXPECT().Execute(ctx, api, traceIds).
					Return(
						nil,
						errors.New("api failed"),
					)
				return f
			},
			expectErr: errors.New("failed to fetch trace 1-64de5a99-5d09aa705e56bbd0152548cb with error: api failed"),
		},
		"found segment with no parent in the tree": {
			sourceTraceId: "1-64de5a99-5d09aa705e56bbd0152548cb",
			rootSegment:   "segment1-id",
			mockGetTraces: func(ctx context.Context, api BatchGetTracesAPI, traceIds []string) *mockGetTracesFunc {
				f := newMockGetTracesFunc(t)
				mockSegmentId1 := "segment1-id"
				mockSegmentDuration1 := float64(1.692291184114e+09)
				mockSegmentId2 := "segment2-id"
				differentParent := "segment3-id"
				mockSegmentDuration2 := float64(1.692291184757e+09)
				f.EXPECT().Execute(ctx, api, traceIds).
					Return(
						map[string]*Trace{
							traceIds[0]: &Trace{
								Id: &traceIds[0],
								Segments: []*Segment{
									{Id: &mockSegmentId1, StartTime: &mockSegmentDuration1},
									{Id: &mockSegmentId2, StartTime: &mockSegmentDuration2, ParentId: &differentParent},
								},
							},
						},
						nil,
					)
				return f
			},
			expectErr: errors.New("failed to build trace tree 1-64de5a99-5d09aa705e56bbd0152548cb with error: found a segment segment2-id with no parent"),
		},
	}

	for name, tt := range cases {
		t.Run(name, func(t *testing.T) {
			ctx := context.TODO()
			traceIds := []string{tt.sourceTraceId}
			xrayClient := newMockXrayClient(t)
			mockGetTraces := tt.mockGetTraces(ctx, xrayClient, traceIds)
			opts := treeOptions{
				xrayClient: xrayClient,
				getTraces:  mockGetTraces.Execute,
			}
			traceTree, err := NewTree(ctx, opts, tt.sourceTraceId)
			if tt.expectErr != nil {
				assert.EqualError(t, err, tt.expectErr.Error())
			} else {
				assert.Equal(t, tt.rootSegment, *traceTree.Root.Id)
				assert.Equal(t, tt.sourceTraceId, *traceTree.SourceTrace.Id)
				assert.Len(t, traceTree.Paths, 1)
				assert.Nil(t, err)
			}
		})
	}
}

func TestInsertSegmentChild(t *testing.T) {
	rootSegmentId := "segment1-id"
	child1SegmentId := "segment2-id"
	child2SegmentId := "segment3-id"
	child1Child1SegmentId := "segment4-id"
	treeRootSegment := &Segment{
		Id: &rootSegmentId,
	}
	assert.Nil(t, treeRootSegment.children)
	child1Segment := &Segment{Id: &child1SegmentId, ParentId: &rootSegmentId}
	child2Segment := &Segment{Id: &child2SegmentId, ParentId: &rootSegmentId}
	child1Child2Segment := &Segment{Id: &child1Child1SegmentId, ParentId: &child1SegmentId}
	InsertSegmentChild(treeRootSegment, child1Segment)
	InsertSegmentChild(treeRootSegment, child2Segment)
	assert.Nil(t, child1Segment.children)
	InsertSegmentChild(child1Segment, child1Child2Segment)

	assert.Equal(t, treeRootSegment.children, []*Segment{child1Segment, child2Segment})
	assert.Equal(t, child1Segment.children, []*Segment{child1Child2Segment})
}

func TestFindLeafSegmentPaths(t *testing.T) {
	rootSegmentId := "segment1-id"
	child1SegmentId := "segment2-id"
	child2SegmentId := "segment3-id"
	child1Child1SegmentId := "segment4-id"

	treeRootSegment := Segment{
		Id:       &rootSegmentId,
		ParentId: nil,
		children: []*Segment{
			&Segment{
				Id:       &child1SegmentId,
				ParentId: &rootSegmentId,
				children: []*Segment{
					&Segment{
						Id: &child1Child1SegmentId, ParentId: &child1SegmentId,
						children: []*Segment{},
					},
				},
			},
			&Segment{
				Id: &child2SegmentId, ParentId: &rootSegmentId,
				children: []*Segment{},
			},
		},
	}

	expectedPaths := [][]*Segment{
		{&treeRootSegment, treeRootSegment.children[0], treeRootSegment.children[0].children[0]},
		{&treeRootSegment, treeRootSegment.children[1]},
	}

	var singlePath []*Segment
	leafPaths := FindLeafSegmentPaths(&treeRootSegment, singlePath)
	assert.Len(t, leafPaths, 2)
	assert.Equal(t, expectedPaths, leafPaths)

}

func TestCreateSegIdtoSegMap(t *testing.T) {
	cases := []struct {
		segments []*Segment
		want     map[string]string
	}{
		{
			segments: []*Segment{
				&Segment{Id: aws.String("segment1-id")},
				&Segment{Id: aws.String("segment2-id")},
				&Segment{Id: aws.String("segment3-id")},
			},
			want: map[string]string{
				"segment1-id": "segment1-id",
				"segment2-id": "segment2-id",
				"segment3-id": "segment3-id",
			},
		},
		{
			segments: []*Segment{
				&Segment{
					Id: aws.String("segment1-id"),
					Subsegments: []*Subsegment{
						&Subsegment{Id: aws.String("subsegment1-id")},
						&Subsegment{Id: aws.String("subsegment2-id")},
					},
				},
			},
			want: map[string]string{
				"segment1-id":    "segment1-id",
				"subsegment1-id": "segment1-id",
				"subsegment2-id": "segment1-id",
			},
		},
	}

	for i, tt := range cases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			got := CreateSegIdtoSegMap(tt.segments)
			for id, parentId := range tt.want {
				assert.Equal(t, parentId, aws.ToString(got[id].Id))
			}
			assert.Equal(t, len(got), len(tt.want))
		})
	}
}

func TestCreateSubSegIdtoSegMap(t *testing.T) {
	cases := []struct {
		segment     *Segment
		subSegments []*Subsegment
		want        []string
	}{
		{
			segment: &Segment{Id: aws.String("segment1-id")},
			subSegments: []*Subsegment{
				&Subsegment{
					Id: aws.String("subsegment1-id"),
					Subsegments: []*Subsegment{
						&Subsegment{
							Id: aws.String("nestedSubsegment1-id"),
						},
					}},
			},
			want: []string{"subsegment1-id", "nestedSubsegment1-id"},
		},
	}
	for i, tt := range cases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			got := CreateSubSegIdtoSegMap(tt.subSegments, tt.segment)
			for _, subSegmentId := range tt.want {
				assert.Equal(t, aws.ToString(tt.segment.Id), aws.ToString(got[subSegmentId].Id))
			}
			assert.Equal(t, len(got), len(tt.want))
		})
	}
}

func TestBuildTreeFromTraceDoc(t *testing.T) {
	filebytes, err := os.ReadFile("./testdata/trace01.json")
	require.NoError(t, err)
	var trace Trace
	err = json.Unmarshal(filebytes, &trace)

	traceMap := map[string]*Trace{
		*trace.Id: &trace,
	}
	tree, err := buildTree(traceMap, &trace)
	require.NoError(t, err)
	log.Println(tree.Paths)
	for _, p := range tree.Paths {
		o := []string{}
		for _, n := range p {
			st := fmt.Sprintf("%v(%v/%v)", *n.Origin, *n.Id, *n.Name)
			o = append(o, st)
		}
		log.Println(o)
	}
}
