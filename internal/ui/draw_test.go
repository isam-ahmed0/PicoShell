package ui

import (
	"image/color"
	"testing"
)

func ops(list *DrawList) []Op { return list.Ops() }

func TestClipStackRestoresRegion(t *testing.T) {
	d := NewDrawList(R(0, 0, 100, 100))
	d.PushClip(R(0, 0, 50, 50))
	d.PushClip(R(0, 0, 25, 25))
	d.Rect(R(0, 0, 100, 100), filled)
	d.PopClip()
	d.Rect(R(0, 0, 100, 100), filled)
	d.PopClip()
	d.Rect(R(0, 0, 100, 100), filled)

	got := ops(d)
	want := []Rect{R(0, 0, 25, 25), R(0, 0, 50, 50), R(0, 0, 100, 100)}
	if len(got) != len(want) {
		t.Fatalf("got %d ops, want %d", len(got), len(want))
	}
	for i, o := range got {
		if o.Clip != want[i] {
			t.Errorf("op %d clip = %v, want %v", i, o.Clip, want[i])
		}
	}
}

func TestUnbalancedPopClipKeepsBounds(t *testing.T) {
	d := NewDrawList(R(0, 0, 100, 100))
	d.PopClip()
	d.Rect(R(0, 0, 100, 100), filled)
	if got := ops(d)[0].Clip; got != R(0, 0, 100, 100) {
		t.Errorf("clip = %v after stray PopClip, want the full bounds", got)
	}
}

func TestTextClipIsBoxWithinCurrentClip(t *testing.T) {
	font := Font{Size: 12}
	d := NewDrawList(R(0, 0, 100, 100))
	d.PushClip(R(0, 0, 60, 60))
	d.Text("label", font, R(10, 10, 40, 30), filled, AlignStart)

	o := ops(d)[0]
	if o.Kind != OpText || o.Rect != R(10, 10, 40, 30) {
		t.Fatalf("op = %+v, want a text op over its box", o)
	}
	if o.Clip != R(10, 10, 40, 30) {
		t.Errorf("text clip = %v, want the text box", o.Clip)
	}
}

func TestRectHelpers(t *testing.T) {
	r := R(10, 10, 30, 20)
	if got := r.Size(); got.W != 20 || got.H != 10 {
		t.Errorf("Size = %+v, want W=20 H=10", got)
	}
	if got := r.Inset(4, 4); got != R(14, 14, 26, 16) {
		t.Errorf("Inset = %v", got)
	}
	if got := r.Inset(50, 0); !got.Empty() {
		t.Errorf("oversized Inset = %v, want empty", got)
	}
	if got := r.Intersect(R(0, 0, 15, 40)); got != R(10, 10, 15, 20) {
		t.Errorf("Intersect = %v", got)
	}
	if !R(5, 5, 5, 5).Empty() {
		t.Error("zero-area rect should be empty")
	}
}

func TestAlignIsSeparateFromColor(t *testing.T) {
	// Guards against a swap of the Color and Align parameters, which the
	// compiler cannot catch since both are untyped-friendly.
	d := NewDrawList(R(0, 0, 10, 10))
	d.Text("x", Font{Size: 8}, R(0, 0, 10, 10), filled, AlignEnd)
	o := ops(d)[0]
	if o.Text.Align != AlignEnd || o.Text.Color != filled {
		t.Errorf("text run = %+v", o.Text)
	}
}

var filled = color.NRGBA{0xff, 0xff, 0xff, 0xff}
