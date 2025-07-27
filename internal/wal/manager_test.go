package wal

import "testing"

func TestParseWALFileName(t *testing.T) {
	tests := []struct {
		name     string
		fileName string
		want     *Segment
		wantErr  bool
	}{
		{
			name:     "valid WAL file name",
			fileName: "000000010000000000000001",
			want: &Segment{
				Timeline:  1,
				LogicalID: 0,
				SegmentID: 1,
			},
			wantErr: false,
		},
		{
			name:     "invalid WAL file name",
			fileName: "invalid",
			want:     nil,
			wantErr:  true,
		},
		{
			name:     "invalid timeline",
			fileName: "XXXXXXXX0000000000000001",
			want:     nil,
			wantErr:  true,
		},
		{
			name:     "higher timeline",
			fileName: "0000000A0000000000000001",
			want: &Segment{
				Timeline:  10,
				LogicalID: 0,
				SegmentID: 1,
			},
			wantErr: false,
		},
		{
			name:     "higher logical ID",
			fileName: "000000010000000A00000001",
			want: &Segment{
				Timeline:  1,
				LogicalID: 10,
				SegmentID: 1,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseWALFileName(tt.fileName)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseWALFileName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if got.Timeline != tt.want.Timeline {
				t.Errorf("Timeline = %v, want %v", got.Timeline, tt.want.Timeline)
			}
			if got.LogicalID != tt.want.LogicalID {
				t.Errorf("LogicalID = %v, want %v", got.LogicalID, tt.want.LogicalID)
			}
			if got.SegmentID != tt.want.SegmentID {
				t.Errorf("SegmentID = %v, want %v", got.SegmentID, tt.want.SegmentID)
			}
		})
	}
}

func TestSegment_Ordering(t *testing.T) {
	segments := []*Segment{
		{Timeline: 1, LogicalID: 0, SegmentID: 1},
		{Timeline: 1, LogicalID: 0, SegmentID: 2},
		{Timeline: 2, LogicalID: 0, SegmentID: 1},
		{Timeline: 1, LogicalID: 1, SegmentID: 1},
	}

	// Test timeline ordering
	if !isSegmentBefore(segments[0], segments[2]) {
		t.Error("Expected segment with lower timeline to be before segment with higher timeline")
	}

	// Test logical ID ordering within same timeline
	if !isSegmentBefore(segments[0], segments[3]) {
		t.Error("Expected segment with lower logical ID to be before segment with higher logical ID")
	}

	// Test segment ID ordering within same timeline and logical ID
	if !isSegmentBefore(segments[0], segments[1]) {
		t.Error("Expected segment with lower segment ID to be before segment with higher segment ID")
	}
}

func isSegmentBefore(a, b *Segment) bool {
	if a.Timeline != b.Timeline {
		return a.Timeline < b.Timeline
	}
	if a.LogicalID != b.LogicalID {
		return a.LogicalID < b.LogicalID
	}
	return a.SegmentID < b.SegmentID
}
