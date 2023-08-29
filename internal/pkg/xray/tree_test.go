package xray

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewTree(t *testing.T) {
	cases := map[string]struct {
		mockGetTraces          func(ctx context.Context, api BatchGetTracesAPI, traceIds []string) *mockGetTracesFunc
		sourceTraceId          string
		fetchChildLinkedTraces bool
		rootSegment            string
		expectErr              error
	}{
		"success": {
			sourceTraceId:          "1-64de5a99-5d09aa705e56bbd0152548cb",
			rootSegment:            "segment1-id",
			fetchChildLinkedTraces: false,
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
		"no segments found": {
			sourceTraceId:          "1-64de5a99-5d09aa705e56bbd0152548cb",
			fetchChildLinkedTraces: false,
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
		"get traces api failed": {
			sourceTraceId:          "1-64de5a99-5d09aa705e56bbd0152548cb",
			fetchChildLinkedTraces: false,
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
			sourceTraceId:          "1-64de5a99-5d09aa705e56bbd0152548cb",
			rootSegment:            "segment1-id",
			fetchChildLinkedTraces: false,
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
			expectErr: errors.New("found a segment segment2-id with no parent"),
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
			traceTree, err := NewTree(ctx, opts, tt.sourceTraceId, tt.fetchChildLinkedTraces)
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

func TestInsertTraceTreeNode(t *testing.T) {
	rootSegmentId := "segment1-id"
	child1SegmentId := "segment2-id"
	child2SegmentId := "segment3-id"
	child1Child1SegmentId := "segment4-id"
	parentNotFoundSegmentId := "segment5-id"
	unknownParentSegmentId := "unknown-id"
	traceTreeRoot := InsertTraceTreeNode(nil, &Segment{Id: &rootSegmentId, ParentId: nil})
	assert.Nil(t, traceTreeRoot.Children)

	child1TreeNode := InsertTraceTreeNode(traceTreeRoot, &Segment{Id: &child1SegmentId, ParentId: &rootSegmentId})
	child2TreeNode := InsertTraceTreeNode(traceTreeRoot, &Segment{Id: &child2SegmentId, ParentId: &rootSegmentId})
	child1Child1TreeNode := InsertTraceTreeNode(traceTreeRoot, &Segment{Id: &child1Child1SegmentId, ParentId: &child1SegmentId})
	unknownParentTreeNode := InsertTraceTreeNode(traceTreeRoot, &Segment{Id: &parentNotFoundSegmentId, ParentId: &unknownParentSegmentId})

	assert.Equal(t, rootSegmentId, *traceTreeRoot.SegmentObject.Id)
	assert.Equal(t, traceTreeRoot.Children, []*TraceTreeNode{child1TreeNode, child2TreeNode})
	assert.Equal(t, child1TreeNode.Children, []*TraceTreeNode{child1Child1TreeNode})
	assert.Nil(t, unknownParentTreeNode)

}

func TestFindLeafSegmentPaths(t *testing.T) {
	rootSegmentId := "segment1-id"
	child1SegmentId := "segment2-id"
	child2SegmentId := "segment3-id"
	child1Child1SegmentId := "segment4-id"

	traceTreeRoot := TraceTreeNode{
		SegmentObject: &Segment{Id: &rootSegmentId, ParentId: nil},
		Children: []*TraceTreeNode{
			&TraceTreeNode{
				SegmentObject: &Segment{Id: &child1SegmentId, ParentId: &rootSegmentId},
				Children: []*TraceTreeNode{
					&TraceTreeNode{
						SegmentObject: &Segment{Id: &child1Child1SegmentId, ParentId: &child1SegmentId},
						Children:      []*TraceTreeNode{},
					},
				},
			},
			&TraceTreeNode{
				SegmentObject: &Segment{Id: &child2SegmentId, ParentId: &rootSegmentId},
				Children:      []*TraceTreeNode{},
			},
		},
	}

	expectedPaths := [][]*Segment{
		{traceTreeRoot.SegmentObject, traceTreeRoot.Children[0].SegmentObject, traceTreeRoot.Children[0].Children[0].SegmentObject},
		{traceTreeRoot.SegmentObject, traceTreeRoot.Children[1].SegmentObject},
	}

	var singlePath []*Segment
	leafPaths := FindLeafSegmentPaths(&traceTreeRoot, singlePath)
	assert.Len(t, leafPaths, 2)
	assert.Equal(t, expectedPaths, leafPaths)

}
