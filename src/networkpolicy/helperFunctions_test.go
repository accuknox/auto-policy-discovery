package networkpolicy

import (
	"testing"

	"github.com/accuknox/knoxAutoPolicy/src/types"
	"github.com/stretchr/testify/assert"
)

const ShouldBeEqual = "they should be equal"

// =========== //
// == Label == //
// =========== //

func TestDescendingLabelCountMap(t *testing.T) {
	labelCountMap := map[string]int{
		"a": 5,
		"b": 1,
		"c": 2,
	}
	expected := []LabelCount{LabelCount{Label: "a", Count: 5.01},
		LabelCount{Label: "c", Count: 2.01},
		LabelCount{Label: "b", Count: 1.01}}

	results := descendingLabelCountMap(labelCountMap)

	assert.Equal(t, expected, results, ShouldBeEqual)
}

func TestContainLabelByConfiguration(t *testing.T) {
	ignoreLabels := []string{"ignore=test"}
	flowLabels := []string{"nonignore=test"}

	results := containLabelByConfiguration(ignoreLabels, flowLabels)

	assert.Equal(t, false, results, ShouldBeEqual)
}

func TestCombinationLabels(t *testing.T) {
	elements := []string{"a", "b", "c"}

	results := combinationLabels(elements, 2)
	expected := [][]string{{"a", "b"}, {"a", "c"}, {"b", "c"}}

	assert.Equal(t, expected, results, ShouldBeEqual)
}

// ==================================== //
// == Removing an Element from Slice == //
// ==================================== //

func TestRemoveSrcFromSlice(t *testing.T) {
	srcs := []SrcSimple{
		SrcSimple{Namespace: "test1",
			PodName: "testPod1"},
		SrcSimple{Namespace: "test2",
			PodName: "testPod2"},
	}

	removedSrc := SrcSimple{Namespace: "test2",
		PodName: "testPod2"}

	results := removeSrcFromSlice(srcs, removedSrc)
	expected := []SrcSimple{
		SrcSimple{Namespace: "test1",
			PodName: "testPod1"},
	}

	assert.Equal(t, expected, results, ShouldBeEqual)
}

func TestRemoveDstFromSlice(t *testing.T) {
	dsts := []Dst{
		Dst{Namespace: "test1",
			PodName: "testPod1"},
		Dst{Namespace: "test2",
			PodName: "testPod2"},
	}

	removedDst := Dst{Namespace: "test2",
		PodName: "testPod2"}

	results := removeDstFromSlice(dsts, removedDst)
	expected := []Dst{
		Dst{Namespace: "test1",
			PodName: "testPod1"},
	}

	assert.Equal(t, expected, results, ShouldBeEqual)
}

func TestRemoveDstFromMergedDstSlice(t *testing.T) {
	ports := []MergedPortDst{
		MergedPortDst{
			Namespace: "test1",
			PodName:   "testPod1",
			ToPorts: []types.SpecPort{types.SpecPort{
				Port:     "80",
				Protocol: "tcp",
			}},
		},
		MergedPortDst{
			Namespace: "test2",
			PodName:   "testPod2",
			ToPorts: []types.SpecPort{types.SpecPort{
				Port:     "8080",
				Protocol: "tcp",
			}},
		},
	}

	removedPort := MergedPortDst{
		Namespace: "test2",
		PodName:   "testPod2",
		ToPorts: []types.SpecPort{types.SpecPort{
			Port:     "8080",
			Protocol: "tcp",
		}},
	}

	results := removeDstFromMergedDstSlice(ports, removedPort)
	expected := []MergedPortDst{
		MergedPortDst{
			Namespace: "test1",
			PodName:   "testPod1",
			ToPorts: []types.SpecPort{types.SpecPort{
				Port:     "80",
				Protocol: "tcp",
			}},
		},
	}

	assert.Equal(t, expected, results, ShouldBeEqual)
}
