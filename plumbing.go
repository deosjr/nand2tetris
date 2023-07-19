package main

func chanBitArray16() [16]chan bit {
    out := [16]chan bit{}
    for i:=0; i<16; i++ {
        out[i] = make(chan bit, 1)
    }
    return out
}

func fanout[T any](in chan T, n int) []chan T {
    out := make([]chan T, n)
    for i:=0; i<n; i++ {
        out[i] = make(chan T, 1)
    }
    go func() {
        for {
            x := <-in
            for i:=0; i<n; i++ {
                out[i]<-x
            }
        }
    }()
    return out
}

func split16(in chan [16]bit, out [16]chan bit) {
    go func() {
        for {
            x := <-in
            for i:=0; i<16; i++ {
                out[i] <- x[i]
            }
        }
    }()
}

func gather16(in [16]chan bit, out chan [16]bit) {
    go func() {
        for {
            value := [16]bit{}
            for i:=0; i<16; i++ {
                value[i] = <-in[i]
            }
            out<-value
        }
    }()
}

func gather16x8(in [8]chan [16]bit, out chan [8][16]bit) {
    go func() {
        for {
            value := [8][16]bit{}
            for i:=0; i<8; i++ {
                value[i] = <-in[i]
            }
            out<-value
        }
    }()
}
