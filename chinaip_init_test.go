package main

import (
	"testing"
)

func Test_cidrCalc(t *testing.T) {
	type args struct {
		mask string
	}
	tests := []struct {
		name    string
		args    args
		want    uint
		wantErr bool
	}{
		{"correctness test", args{mask: "18"}, 16384, false},
		{"boundary test", args{mask: "0"}, 4294967296, false},
		{"boundary test", args{mask: "32"}, 1, false},
		{"boundary test", args{mask: "33"}, 0, true},
		{"format test", args{mask: "18\n"}, 0, true},
		{"format test", args{mask: "18 "}, 0, true},
		{"format test", args{mask: "/18"}, 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := cidrCalc(tt.args.mask)
			if (err != nil) != tt.wantErr {
				t.Errorf("cidrCalc() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("cidrCalc() = %v, want %v", got, tt.want)
			}
		})
	}
}
