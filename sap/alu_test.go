package main

import (
	"testing"
)

func TestSAP2Alu(t *testing.T) {
	for i, tt := range []struct{
			a, b uint16
			s3, s2, s1, s0 bit
			want uint16
		}{
			{
				a: 0b111111101100,
				b: 0b111100010000,
				s2: true, s0: true, // 0101: IOR
				want: 0b111111111100,
				},
		}{
			out := SAP2Alu(toBit12(tt.a), toBit12(tt.b), tt.s3, tt.s2, tt.s1, tt.s0)
			got := fromBit12(out)
			if got != tt.want {
				t.Errorf("%d): got %d but want %d", i, got, tt.want)
			}
		}
}