package main

type Port interface {
	SendIn(uint16)
	SendLoad(bool)
	Out() uint16
	ClockTick()
}

type InputDevice interface {
	OnPortRead(Port)
}

type OutputDevice interface {
	OnPortWrite(Port)
}

type TapeReader struct {
	tape []uint16
	head int
}

func (tr *TapeReader) OnPortRead(p Port) {
	if tr.head >= len(tr.tape) {
		return
	}
	p.SendIn(tr.tape[tr.head])
	p.SendLoad(true)
	tr.head++
}

type TapeWriter struct {
	tape []uint16
}

func (tw *TapeWriter) OnPortWrite(p Port) {
	tw.tape = append(tw.tape, p.Out())
}

func (tw *TapeWriter) Out() []uint16 {
	return tw.tape
}
