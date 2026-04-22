package main

import (
	"testing"
)

type testDebugger struct {
	t     *testing.T
	i     int
	index int
}

func (*testDebugger) BeforeLoop() {}
func (td *testDebugger) BeforeTick(c Computer) {
	td.i++
}
func (td *testDebugger) AfterTick(c Computer) {
	if td.i > 1000000 {
		td.t.Fatalf("%d): took too long", td.index)
	}
}

func TestSAP1(t *testing.T) {
	for i, tt := range []struct {
		in   [16]string
		want uint8
	}{
		{
			in: [16]string{
				"LDA R9", "ADD RA", "OUT", "HLT", "", "", "", "", "", "15", "20", "", "", "", "", "",
			},
			want: 35,
		},
	} {
		program, err := assembleSAP1FromStrings(tt.in)
		if err != nil {
			t.Fatal(err)
		}

		computer := NewSAP1()
		computer.LoadProgram(program)

		td := &testDebugger{t: t, index: i}
		run(computer, td)

		out := computer.O.Out()
		if tt.want != out {
			t.Errorf("%d): want %v got %v", i, tt.want, out)
		}
	}
}
