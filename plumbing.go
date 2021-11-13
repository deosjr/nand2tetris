package main

// a great usecase for generics :)
// or go:generate ?

func chanBitArray16() [16]chan bit {
    out := [16]chan bit{}
    for i:=0; i<16; i++ {
        out[i] = make(chan bit, 1)
    }
    return out
}

func fanout(in chan bit, n int) []chan bit {
    out := make([]chan bit, n)
    for i:=0; i<n; i++ {
        out[i] = make(chan bit, 1)
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

func fanout16(in chan [16]bit, n int) []chan [16]bit {
    out := make([]chan [16]bit, n)
    for i:=0; i<n; i++ {
        out[i] = make(chan [16]bit, 1)
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

func fanout3(in chan [3]bit, n int) []chan[3]bit {
    out := make([]chan [3]bit, n)
    for i:=0; i<n; i++ {
        out[i] = make(chan [3]bit, 1)
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

func fanout6(in chan [6]bit, n int) []chan[6]bit {
    out := make([]chan [6]bit, n)
    for i:=0; i<n; i++ {
        out[i] = make(chan [6]bit, 1)
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

func fanout9(in chan [9]bit, n int) []chan[9]bit {
    out := make([]chan [9]bit, n)
    for i:=0; i<n; i++ {
        out[i] = make(chan [9]bit, 1)
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

func fanout12(in chan [12]bit, n int) []chan[12]bit {
    out := make([]chan [12]bit, n)
    for i:=0; i<n; i++ {
        out[i] = make(chan [12]bit, 1)
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
