package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
)

type Screen interface {
	SendIn(uint16)
	SendLoad(bool)
	SendAddress(uint16)
	Out() uint16
	ClockTick()
}

// NOTE: no way to write to keyboard register
type Keyboard interface {
	Out() uint16
	ClockTick()
}

// writes values from an input file to a specific register
// Load clears the register and signals to read the next character
// -> so writing to this register sets it to 0 instead
type TapeReader interface {
	Out() uint16
	SendLoad(bool)
	ClockTick()
	LoadInputTape(string) error
	LoadInputTapes([]string) error
	LoadInputReaders([]io.Reader)
}

// reads values to write to an output file or stdout
// sets the register to 0 when it is _done_ writing a character
type TapeWriter interface {
	Out() uint16
	SendIn(uint16)
	SendLoad(bool)
	ClockTick()
	// NOTE: truncates!
	LoadOutputTape(string) error
	LoadOutputWriter(io.Writer)
}

type Screen512x256 struct {
	ram *BuiltinRAM16K // actually use only the first 8K
}

func NewScreen512x256() *Screen512x256 {
	return &Screen512x256{
		ram: NewBuiltinRAM16K(),
	}
}

func (s *Screen512x256) SendIn(in uint16) {
	s.ram.SendIn(in)
}

func (s *Screen512x256) SendLoad(load bool) {
	s.ram.SendLoad(load)
}

func (s *Screen512x256) SendAddress(addr uint16) {
	if addr >= 8192 {
		return
		//panic("access screen memory beyond 8K")
	}
	s.ram.SendAddress(addr)
}

func (s *Screen512x256) Out() uint16 {
	return s.ram.Out()
}

func (s *Screen512x256) ClockTick() {
	s.ram.ClockTick()
}

// Lets leave this one for posterity (ie old asm files)
// but I misunderstood the default screen dimensions :)
type Screen256x512 struct {
	ram *BuiltinRAM16K // actually use only the first 8K
}

func NewScreen256x512() *Screen256x512 {
	return &Screen256x512{
		ram: NewBuiltinRAM16K(),
	}
}

func (s *Screen256x512) SendIn(in uint16) {
	s.ram.SendIn(in)
}

func (s *Screen256x512) SendLoad(load bool) {
	s.ram.SendLoad(load)
}

func (s *Screen256x512) SendAddress(addr uint16) {
	if addr >= 8192 {
		return
		//panic("access screen memory beyond 8K")
	}
	s.ram.SendAddress(addr)
}

func (s *Screen256x512) Out() uint16 {
	return s.ram.Out()
}

func (s *Screen256x512) ClockTick() {
	s.ram.ClockTick()
}

type SimpleKeyboard struct {
	reg *BuiltinRegister
}

func NewSimpleKeyboard() *SimpleKeyboard {
	reg := NewBuiltinRegister()
	reg.SendLoad(true)
	return &SimpleKeyboard{
		reg: reg,
	}
}

func (k *SimpleKeyboard) Out() uint16 {
	return k.reg.Out()
}

func (k *SimpleKeyboard) ClockTick() {
	k.reg.ClockTick()
}

type tapeReader struct {
	reg     *BuiltinRegister
	scanner *bufio.Scanner
	eof     chan bool
}

func NewTapeReader() *tapeReader {
	reg := NewBuiltinRegister()
	reg.SendLoad(true)
	return &tapeReader{
		reg: reg,
		eof: make(chan bool),
	}
}

func (tr *tapeReader) Out() uint16 {
	return tr.reg.Out()
}

func (tr *tapeReader) SendLoad(load bool) {
	if load {
		tr.reg.SendIn(0)
	}
}

func (tr *tapeReader) ClockTick() {
	if tr.scanner == nil || tr.reg.Out() != 0 {
		tr.reg.ClockTick()
		return
	}
	// Here's where we abstract heavily over how this would work exactly
	if tr.reg.Out() == 0 {
		if !tr.scanner.Scan() {
			tr.reg.SendIn(0x1C) // ascii File Separator control character
			tr.reg.ClockTick()
			tr.eof <- true
			return
		}
		char := tr.scanner.Text()[0]
		if char == '\n' {
			char = 0x80 // newlines are ENTER ascii values
		}
		tr.reg.SendIn(uint16(char))
	}
	tr.reg.ClockTick()
}

func (tr *tapeReader) LoadInputTape(filename string) error {
	return tr.LoadInputTapes([]string{filename})
}

// TODO: closing. using ReadCloser disallows strings.NewReader
// NOTE: currently opens all files at once
func (tr *tapeReader) LoadInputTapes(filenames []string) error {
	readers := []io.Reader{}
	for _, filename := range filenames {
		f, err := os.Open(filename)
		if err != nil {
			return err
		}
		readers = append(readers, f)
	}
	tr.LoadInputReaders(readers)
	return nil
}

func (tr *tapeReader) LoadInputReaders(readers []io.Reader) {
	go func() {
		for _, r := range readers {
			tr.scanner = bufio.NewScanner(r)
			tr.scanner.Split(bufio.ScanRunes)
			<-tr.eof
		}
	}()
}

type tapeWriter struct {
	reg *BuiltinRegister
	out int
	w   io.Writer
}

func NewTapeWriter() *tapeWriter {
	reg := NewBuiltinRegister()
	return &tapeWriter{
		reg: reg,
		w:   os.Stdout,
	}
}

func (tr *tapeWriter) Out() uint16 {
	return tr.reg.Out()
}

func (tr *tapeWriter) SendIn(in uint16) {
	tr.reg.SendIn(in)
}

func (tr *tapeWriter) SendLoad(load bool) {
	tr.reg.SendLoad(load)
	if load {
		tr.out = -1
	}
}

func (tr *tapeWriter) ClockTick() {
	tr.reg.ClockTick()
	out := tr.reg.Out()
	if int(out) != tr.out {
		fmt.Fprintf(tr.w, "%04x\n", out)
		// this is of course cheating, should be represented by separate write bit
		tr.out = int(out)
	}
}

func (tr *tapeWriter) LoadOutputTape(filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	tr.LoadOutputWriter(f)
	return nil
}

func (tr *tapeWriter) LoadOutputWriter(w io.Writer) {
	tr.w = w
}
