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

package nodemon

import (
	"fmt"
	"sync"

	"github.com/0xsoniclabs/hyperion/driver/monitoring"
	"golang.org/x/exp/constraints"
)

func init() {
	smaPeriods := []int{10, 100, 1000}
	for _, period := range smaPeriods {

		// TransactionThroughputSMA is a metric that aggregates output from another series and computes
		// Simple Moving Average.
		TransactionThroughputSMA := monitoring.Metric[monitoring.Node, monitoring.Series[monitoring.BlockNumber, float32]]{
			Name:        fmt.Sprintf("TransactionsThroughputSMA_%d", period),
			Description: "Transaction throughput standard moving average",
		}

		currentPeriod := period // capture current value of the period
		smaFactory := func(input monitoring.Series[monitoring.BlockNumber, float32]) monitoring.Series[monitoring.BlockNumber, float32] {
			return monitoring.NewSMASeries[monitoring.BlockNumber, float32](input, currentPeriod)
		}
		metricsFactory := func(monitor *monitoring.Monitor) monitoring.Source[monitoring.Node, monitoring.Series[monitoring.BlockNumber, float32]] {
			return newNodeBlockSeriesTransformation(monitor, TransactionThroughputSMA, TransactionsThroughput, smaFactory)
		}

		if err := monitoring.RegisterSource(TransactionThroughputSMA, metricsFactory); err != nil {
			panic(fmt.Sprintf("failed to register metric source: %v", err))
		}
	}
}

// NodeBlockSeriesTransformation is a source that captures an input source, and computes certain transformation on top of it.
// The input source for this type must have the node as a subject and the Series as a value.
// This type produces the same Nodes as the subjects and the series with the required transformation.
type NodeBlockSeriesTransformation[K constraints.Ordered, T any, X monitoring.Series[K, T]] struct {
	metric        monitoring.Metric[monitoring.Node, X]
	source        monitoring.Metric[monitoring.Node, X] // source metrics to transform
	monitor       *monitoring.Monitor
	series        map[monitoring.Node]X
	seriesFactory func(X) X // transform input series to the output series
	seriesLock    *sync.Mutex
}

// NewNodeSeriesTransformation creates a new source that can transform input source to the output source.
// This transformation is limited to the source where the Node is the subject and values are series.
// The output source is transformed to contain the same subjects, which addresses new, transformed, series
func NewNodeSeriesTransformation[K constraints.Ordered, T any, X monitoring.Series[K, T]](
	monitor *monitoring.Monitor,
	metric monitoring.Metric[monitoring.Node, X],
	source monitoring.Metric[monitoring.Node, X],
	seriesFactory func(X) X) *NodeBlockSeriesTransformation[K, T, X] {

	m := &NodeBlockSeriesTransformation[K, T, X]{
		metric:        metric,
		source:        source,
		seriesFactory: seriesFactory,
		monitor:       monitor,
		series:        make(map[monitoring.Node]X, 50),
		seriesLock:    &sync.Mutex{},
	}

	return m
}

// newNodeBlockSeriesTransformation creates the same instance as public NewNodeSeriesTransformation but typed to the BlockSeries as a Series.
func newNodeBlockSeriesTransformation[T any](
	monitor *monitoring.Monitor,
	metric monitoring.Metric[monitoring.Node, monitoring.Series[monitoring.BlockNumber, T]],
	source monitoring.Metric[monitoring.Node, monitoring.Series[monitoring.BlockNumber, T]],
	seriesFactory func(monitoring.Series[monitoring.BlockNumber, T]) monitoring.Series[monitoring.BlockNumber, T]) monitoring.Source[monitoring.Node, monitoring.Series[monitoring.BlockNumber, T]] {

	res := NewNodeSeriesTransformation[monitoring.BlockNumber, T, monitoring.Series[monitoring.BlockNumber, T]](monitor, metric, source, seriesFactory)
	return res
}

func (s *NodeBlockSeriesTransformation[K, T, X]) GetMetric() monitoring.Metric[monitoring.Node, X] {
	return s.metric
}

func (s *NodeBlockSeriesTransformation[K, T, X]) GetSubjects() []monitoring.Node {
	return monitoring.GetSubjects(s.monitor, s.source)
}

func (s *NodeBlockSeriesTransformation[K, T, X]) GetData(node monitoring.Node) (X, bool) {
	s.seriesLock.Lock()
	defer s.seriesLock.Unlock()

	res, exists := s.series[node]
	if !exists {
		source, existsSource := monitoring.GetData(s.monitor, node, s.source)
		if existsSource {
			newSeries := s.seriesFactory(source)
			s.series[node] = newSeries
			return newSeries, true
		} else {
			return res, false
		}
	}

	return res, exists
}

func (s *NodeBlockSeriesTransformation[K, T, X]) Shutdown() error {
	return nil
}

func (s *NodeBlockSeriesTransformation[K, T, X]) ForEachRecord(consumer func(r monitoring.Record)) {
	subjects := monitoring.GetSubjects(s.monitor, s.source)
	for _, subject := range subjects {
		series, exists := monitoring.GetData(s.monitor, subject, s.source)
		if !exists {
			continue
		}

		r := monitoring.Record{}
		r.SetSubject(subject)

		var first K
		latest := series.GetLatest()
		if latest == nil {
			continue
		}
		allData := series.GetRange(first, latest.Position)
		for _, point := range allData {
			r.SetPosition(point.Position).SetValue(point.Value)
			consumer(r)
		}
		r.SetPosition(latest.Position).SetValue(latest.Value)
		consumer(r)
	}
}
