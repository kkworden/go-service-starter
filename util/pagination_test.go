package util_test

import (
	"testing"

	"go-service-starter/util"
)

func TestClampPagination(t *testing.T) {
	tests := []struct {
		name       string
		limit      int
		offset     int
		wantLimit  int
		wantOffset int
	}{
		{"defaults", 0, 0, util.DefaultPageLimit, 0},
		{"negative limit", -5, 0, util.DefaultPageLimit, 0},
		{"negative offset", 10, -3, 10, 0},
		{"over max", 500, 0, util.MaxPageLimit, 0},
		{"valid", 50, 20, 50, 20},
		{"exactly max", util.MaxPageLimit, 0, util.MaxPageLimit, 0},
		{"exactly default", util.DefaultPageLimit, 0, util.DefaultPageLimit, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotLimit, gotOffset := util.ClampPagination(tt.limit, tt.offset)
			if gotLimit != tt.wantLimit {
				t.Errorf("limit = %d, want %d", gotLimit, tt.wantLimit)
			}
			if gotOffset != tt.wantOffset {
				t.Errorf("offset = %d, want %d", gotOffset, tt.wantOffset)
			}
		})
	}
}
