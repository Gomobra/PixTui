package native

// TextWorkCounts tracks text layout/seek/scan work for tests.
// Outside test builds the record is effectively void and every Add
// compiles to nothing (in Go it is a no-op).
// Mirrors text-work.zig Counts.
type TextWorkCounts struct {
	Layouts              uint64
	LinesLaidOut         uint64
	RopeDiffNodes        uint64
	LinesShifted         uint64
	Splices              uint64
	MeasurePasses        uint64
	MeasureLines         uint64
	LineSeeks            uint64
	LineSeekNodes        uint64
	HighlightLeaves      uint64
	RemeasureNodes       uint64
	ClusterScanSteps     uint64
	LeftSkipSteps        uint64
	SpanSearchSteps      uint64
	ClustersDrawn        uint64
	PrefixSeeks          uint64
	GraphemeWidthReads   uint64
	BytesScanned         uint64
	AsciiBytes           uint64
	StyledTextSets       uint64
	StyledTextBytes      uint64
	StyleRegistrations   uint64
	StyleDependentVisits uint64
	GraphemeJoins        uint64
	GraphemeJoinBytes    uint64
	LineSpanBuilds       uint64
	LineSpanHighlights   uint64
}

var textWorkEnabled = false

// setTextWorkEnabled enables or disables text work counting.
// Used by tests to activate counting.
func SetTextWorkEnabled(enabled bool) {
	textWorkEnabled = enabled
}

var textWorkCounts TextWorkCounts

// Add increments a text work counter by n.
// Mirrors text-work.zig add.
func Add(comptimeField string, n uint64) {
	if !textWorkEnabled {
		return
	}
	switch comptimeField {
	case "layouts":
		textWorkCounts.Layouts += n
	case "lines_laid_out":
		textWorkCounts.LinesLaidOut += n
	case "rope_diff_nodes":
		textWorkCounts.RopeDiffNodes += n
	case "lines_shifted":
		textWorkCounts.LinesShifted += n
	case "splices":
		textWorkCounts.Splices += n
	case "measure_passes":
		textWorkCounts.MeasurePasses += n
	case "measure_lines":
		textWorkCounts.MeasureLines += n
	case "line_seeks":
		textWorkCounts.LineSeeks += n
	case "line_seek_nodes":
		textWorkCounts.LineSeekNodes += n
	case "highlight_leaves":
		textWorkCounts.HighlightLeaves += n
	case "remeasure_nodes":
		textWorkCounts.RemeasureNodes += n
	case "cluster_scan_steps":
		textWorkCounts.ClusterScanSteps += n
	case "left_skip_steps":
		textWorkCounts.LeftSkipSteps += n
	case "span_search_steps":
		textWorkCounts.SpanSearchSteps += n
	case "clusters_drawn":
		textWorkCounts.ClustersDrawn += n
	case "prefix_seeks":
		textWorkCounts.PrefixSeeks += n
	case "grapheme_width_reads":
		textWorkCounts.GraphemeWidthReads += n
	case "bytes_scanned":
		textWorkCounts.BytesScanned += n
	case "ascii_bytes":
		textWorkCounts.AsciiBytes += n
	case "styled_text_sets":
		textWorkCounts.StyledTextSets += n
	case "styled_text_bytes":
		textWorkCounts.StyledTextBytes += n
	case "style_registrations":
		textWorkCounts.StyleRegistrations += n
	case "style_dependent_visits":
		textWorkCounts.StyleDependentVisits += n
	case "grapheme_joins":
		textWorkCounts.GraphemeJoins += n
	case "grapheme_join_bytes":
		textWorkCounts.GraphemeJoinBytes += n
	case "line_span_builds":
		textWorkCounts.LineSpanBuilds += n
	case "line_span_highlights":
		textWorkCounts.LineSpanHighlights += n
	}
}

// Take returns the counts since the last call and resets them.
// Mirrors text-work.zig take.
func Take() TextWorkCounts {
	result := textWorkCounts
	textWorkCounts = TextWorkCounts{}
	return result
}
