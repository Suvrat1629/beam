// Licensed to the Apache Software Foundation (ASF) under one or more
// contributor license agreements.  See the NOTICE file distributed with
// this work for additional information regarding copyright ownership.
// The ASF licenses this file to You under the Apache License, Version 2.0
// (the "License"); you may not use this file except in compliance with
// the License.  You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package coder

import (
	"bytes"
	"math"
	"testing"

	"github.com/apache/beam/sdks/v2/go/pkg/beam/core/typex"
)

func makePaneInfo(timing typex.PaneTiming, first, last bool, index, nsIndex int64) typex.PaneInfo {
	return typex.PaneInfo{Timing: timing, IsFirst: first, IsLast: last, Index: index, NonSpeculativeIndex: nsIndex}
}

func equalPanes(left, right typex.PaneInfo) bool {
	return (left.Timing == right.Timing) && (left.IsFirst == right.IsFirst) && (left.IsLast == right.IsLast) && (left.Index == right.Index) && (left.NonSpeculativeIndex == right.NonSpeculativeIndex)
}

func TestPaneCoder(t *testing.T) {
	tests := []struct {
		name      string
		timing    typex.PaneTiming
		first     bool
		last      bool
		index     int64
		nsIndex   int64
		firstByte byte
	}{
		{
			"false bools",
			typex.PaneUnknown,
			false,
			false,
			0,
			0,
			0b00001100,
		},
		{
			"true bools",
			typex.PaneUnknown,
			true,
			true,
			0,
			0,
			0b00001111,
		},
		{
			"first pane",
			typex.PaneUnknown,
			true,
			false,
			0,
			0,
			0b00001101,
		},
		{
			"last pane",
			typex.PaneUnknown,
			false,
			true,
			0,
			0,
			0b00001110,
		},
		{
			"on time, different index and non-speculative",
			typex.PaneOnTime,
			false,
			false,
			1,
			2,
			0b00100100,
		},
		{
			"valid early pane",
			typex.PaneEarly,
			true,
			false,
			math.MaxInt64,
			-1,
			0b00010001,
		},
		{
			"on time, max non-speculative index",
			typex.PaneOnTime,
			false,
			true,
			0,
			math.MaxInt64,
			0b00100110,
		},
		{
			"late pane, max index",
			typex.PaneLate,
			false,
			false,
			math.MaxInt64,
			0,
			0b00101000,
		},
		{
			"on time, min non-speculative index",
			typex.PaneOnTime,
			false,
			true,
			0,
			math.MinInt64,
			0b00100110,
		},
		{
			"late, min index",
			typex.PaneLate,
			false,
			false,
			math.MinInt64,
			0,
			0b00101000,
		},
		{
			"last late firing",
			typex.PaneLate,
			false,
			true,
			2,
			1,
			0b00101010,
		},
		{
			"encodeByte 41",
			typex.PaneLate,
			true,
			false,
			2,
			1,
			0b00101001, // 41
		},
		{
			"encodeByte 18",
			typex.PaneEarly,
			false,
			true,
			0,
			-1,
			0b00010010, // 18
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			input := makePaneInfo(test.timing, test.first, test.last, test.index, test.nsIndex)
			var buf bytes.Buffer
			err := EncodePane(input, &buf)
			if err != nil {
				t.Fatalf("failed to encode pane %v, got %v", input, err)
			}
			first := buf.Bytes()[0]
			if got, want := first, test.firstByte; got != want {
				t.Errorf("Unexpected First Byte: got %#08b, want %#08b, for %v ", got, want, input)
			}
			got, err := DecodePane(&buf)
			if err != nil {
				t.Fatalf("failed to decode pane from buffer %v, got %v", &buf, err)
			}
			if want := input; !equalPanes(got, want) {
				t.Errorf("got pane %v, want %v", got, want)
			}
		})
	}
}

func TestEncodePane_bad(t *testing.T) {
	tests := []struct {
		name    string
		timing  typex.PaneTiming
		first   bool
		last    bool
		index   int64
		nsIndex int64
	}{
		{
			"invalid early pane, max ints",
			typex.PaneEarly,
			true,
			false,
			math.MaxInt64,
			math.MaxInt64,
		},
		{
			"invalid early pane, min ints",
			typex.PaneEarly,
			true,
			false,
			math.MinInt64,
			math.MinInt64,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			input := makePaneInfo(test.timing, test.first, test.last, test.index, test.nsIndex)
			var buf bytes.Buffer
			err := EncodePane(input, &buf)
			if err == nil {
				t.Errorf("successfully encoded pane when it should have failed, got %v", &buf)
			}
		})
	}
}
