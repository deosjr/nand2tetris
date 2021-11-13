package main

import "testing"

func TestBuiltinRegister(t *testing.T) {
    for _, tt := range []struct{
        sequence func(*BuiltinRegister)
        want uint16
    }{
        {
            sequence: func(b *BuiltinRegister) {
                b.SendIn(42)
            },
            want: 0,
        },
        {
            sequence: func(b *BuiltinRegister) {
                b.SendIn(42)
                b.SendLoad(true)
            },
            want: 0,
        },
        {
            sequence: func(b *BuiltinRegister) {
                b.SendIn(42)
                b.SendLoad(true)
                b.ClockTick()
            },
            want: 42,
        },
    }{
        reg := NewBuiltinRegister()
        tt.sequence(reg)
        got := reg.Out()
        if got != tt.want {
            t.Errorf("got %d but want %d", got, tt.want)
        }
    }
}
