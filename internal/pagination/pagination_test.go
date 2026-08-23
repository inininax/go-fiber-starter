package pagination

import "testing"

func TestNewPageQuery_Clamps(t *testing.T) {
	cases := []struct {
		name      string
		page, lim int
		wantPage  int
		wantLimit int
	}{
		{"zeroValues", 0, 0, DefaultPage, DefaultLimit},
		{"negative", -5, -10, DefaultPage, DefaultLimit},
		{"overMax", 3, MaxLimit + 1, 3, MaxLimit},
		{"validPassthrough", 4, 50, 4, 50},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			q := NewPageQuery(tc.page, tc.lim)
			if q.Page != tc.wantPage || q.Limit != tc.wantLimit {
				t.Fatalf("want (%d,%d), got %+v", tc.wantPage, tc.wantLimit, q)
			}
		})
	}
}

func TestPageQuery_Offset(t *testing.T) {
	q := PageQuery{Page: 3, Limit: 20}
	if got := q.Offset(); got != 40 {
		t.Fatalf("want 40, got %d", got)
	}
}

func TestNewPageMeta_Edges(t *testing.T) {
	if m := NewPageMeta(PageQuery{Page: 1, Limit: 20}, 0); m.TotalPages != 0 {
		t.Fatalf("empty result must have 0 pages, got %d", m.TotalPages)
	}
	if m := NewPageMeta(PageQuery{Page: 2, Limit: 20}, 20); m.TotalPages != 1 {
		t.Fatalf("exact multiple must be 1 page, got %d", m.TotalPages)
	}
	if m := NewPageMeta(PageQuery{Page: 1, Limit: 20}, 21); m.TotalPages != 2 {
		t.Fatalf("want 2 pages, got %d", m.TotalPages)
	}
}
