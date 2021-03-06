package views

import (
	"fmt"
	"strings"
	"testing"

	"github.com/derailed/k9s/internal/resource"
	"github.com/stretchr/testify/assert"
)

func TestIsSelector(t *testing.T) {
	uu := map[string]struct {
		sel string
		e   bool
	}{
		"cool":       {"-l app=fred,env=blee", true},
		"noMode":     {"app=fred,env=blee", false},
		"noSpace":    {"-lapp=fred,env=blee", true},
		"wrongLabel": {"-f app=fred,env=blee", false},
	}

	for k, u := range uu {
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, isLabelSelector(u.sel))
		})
	}
}

func TestTrimLabelSelector(t *testing.T) {
	uu := map[string]struct {
		sel, e string
	}{
		"cool":    {"-l app=fred,env=blee", "app=fred,env=blee"},
		"noSpace": {"-lapp=fred,env=blee", "app=fred,env=blee"},
	}

	for k, u := range uu {
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, trimLabelSelector(u.sel))
		})
	}
}

func TestTVSortRows(t *testing.T) {
	uu := []struct {
		rows  resource.RowEvents
		col   int
		asc   bool
		first resource.Row
		e     []string
	}{
		{
			resource.RowEvents{
				"row1": {Fields: resource.Row{"x", "y"}},
				"row2": {Fields: resource.Row{"a", "b"}},
			},
			0,
			true,
			resource.Row{"a", "b"},
			[]string{"row2", "row1"},
		},
		{
			resource.RowEvents{
				"row1": {Fields: resource.Row{"x", "y"}},
				"row2": {Fields: resource.Row{"a", "b"}},
			},
			1,
			true,
			resource.Row{"a", "b"},
			[]string{"row2", "row1"},
		},
		{
			resource.RowEvents{
				"row1": {Fields: resource.Row{"x", "y"}},
				"row2": {Fields: resource.Row{"a", "b"}},
			},
			1,
			false,
			resource.Row{"x", "y"},
			[]string{"row1", "row2"},
		},
		{
			resource.RowEvents{
				"row1": {Fields: resource.Row{"2175h48m0.06015s", "y"}},
				"row2": {Fields: resource.Row{"403h42m34.060166s", "b"}},
			},
			0,
			true,
			resource.Row{"403h42m34.060166s", "b"},
			[]string{"row2", "row1"},
		},
	}

	var v *tableView
	for _, u := range uu {
		keys := make([]string, len(u.rows))
		v.sortRows(u.rows, v.defaultSort, sortColumn{u.col, len(u.rows), u.asc}, keys)
		assert.Equal(t, u.e, keys)
		assert.Equal(t, u.first, u.rows[u.e[0]].Fields)
	}
}

func BenchmarkTVSortRows(b *testing.B) {
	evts := resource.RowEvents{
		"row1": {Fields: resource.Row{"x", "y"}},
		"row2": {Fields: resource.Row{"a", "b"}},
	}
	sc := sortColumn{0, 2, true}
	var v *tableView
	keys := make([]string, len(evts))
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		v.sortRows(evts, v.defaultSort, sc, keys)
	}
}

func BenchmarkTitleReplace(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		fmat := strings.Replace(nsTitleFmt, "[fg", "["+"red", -1)
		fmat = strings.Replace(fmat, ":bg:", ":"+"blue"+":", -1)
		fmat = strings.Replace(fmat, "[hilite", "["+"green", 1)
		fmat = strings.Replace(fmat, "[count", "["+"yellow", 1)
		_ = fmt.Sprintf(fmat, "Pods", "default", 10)
	}
}

func BenchmarkTitleReplace1(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		fmat := strings.Replace(nsTitleFmt, "fg:bg", "red"+":"+"blue", -1)
		fmat = strings.Replace(fmat, "[hilite", "["+"green", 1)
		fmat = strings.Replace(fmat, "[count", "["+"yellow", 1)
		_ = fmt.Sprintf(fmat, "Pods", "default", 10)
	}
}
