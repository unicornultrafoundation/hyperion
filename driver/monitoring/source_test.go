// Copyright 2024 Fantom Foundation
// This file is part of Hyperion System Testing Infrastructure for Sonic.
//
// Hyperion is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Hyperion is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with Hyperion. If not, see <http://www.gnu.org/licenses/>.

package monitoring

import (
	"sort"
	"testing"

	"golang.org/x/exp/slices"
)

var (
	TestNodeMetric = Metric[Node, Series[BlockNumber, int]]{
		Name:        "TestNodeMetric",
		Description: "A test metric for unit tests.",
	}
)

// TestSource is a data source providing a stand-in for actual sources in
// tests. This is required since gomock is (yet) not supporting the generation
// of generic mocks.
type TestSource struct {
	data map[Node]Series[BlockNumber, int]
}

func (s *TestSource) GetMetric() Metric[Node, Series[BlockNumber, int]] {
	return TestNodeMetric
}

func (s *TestSource) GetSubjects() []Node {
	res := make([]Node, 0, len(s.data))
	for node := range s.data {
		res = append(res, node)
	}
	sort.Slice(res, func(i, j int) bool { return res[i] < res[j] })
	return res
}

func (s *TestSource) GetData(node Node) (Series[BlockNumber, int], bool) {
	res, exists := s.data[node]
	return res, exists
}

func (s *TestSource) ForEachRecord(consumer func(Record)) {
	for subject, series := range s.data {
		r := Record{}
		r.SetSubject(subject)

		latest := series.GetLatest()
		if latest == nil {
			continue
		}
		allData := series.GetRange(BlockNumber(0), latest.Position+1)
		for _, point := range allData {
			r.SetPosition(point.Position).SetValue(point.Value)
			consumer(r)
		}
	}
}

func (s *TestSource) Start() error {
	// Nothing to do.
	return nil
}

func (s *TestSource) Shutdown() error {
	// Nothing to do.
	return nil
}

func (s *TestSource) setData(node Node, data Series[BlockNumber, int]) {
	if s.data == nil {
		s.data = map[Node]Series[BlockNumber, int]{}
	}
	s.data[node] = data
}

func TestTestSourceIsSource(t *testing.T) {
	var source TestSource
	var _ Source[Node, Series[BlockNumber, int]] = &source
}

func TestTestSource_ListsCorrectSubjects(t *testing.T) {
	source := TestSource{}
	want := []Node{}
	if got := source.GetSubjects(); !slices.Equal(got, want) {
		t.Errorf("invalid subject list, wanted %v, got %v", want, got)
	}
	source.setData(Node("A"), &TestBlockSeries{[]int{1, 2, 3}})
	want = []Node{Node("A")}
	if got := source.GetSubjects(); !slices.Equal(got, want) {
		t.Errorf("invalid subject list, wanted %v, got %v", want, got)
	}
	source.setData(Node("B"), &TestBlockSeries{[]int{1}})
	want = []Node{Node("A"), Node("B")}
	if got := source.GetSubjects(); !slices.Equal(got, want) {
		t.Errorf("invalid subject list, wanted %v, got %v", want, got)
	}
}

func TestTestSource_RetrievesCorrectDataSeries(t *testing.T) {
	seriesA := &TestBlockSeries{[]int{1, 2}}
	seriesB := &TestBlockSeries{[]int{3, 4, 5}}

	source := TestSource{}
	source.setData(Node("A"), seriesA)
	source.setData(Node("B"), seriesB)

	if series, exists := source.GetData(Node("A")); !exists || series != seriesA {
		t.Errorf("test source returned wrong series")
	}
	if series, exists := source.GetData(Node("B")); !exists || series != seriesB {
		t.Errorf("test source returned wrong series")
	}
	if series, exists := source.GetData(Node("C")); exists || series != nil {
		t.Errorf("test source returned wrong series")
	}
}
