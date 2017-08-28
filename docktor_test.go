package main

import (
	"testing"
	"time"
)

func Test_stringToTime(t *testing.T) {
	type args struct {
		envVar      string
		defaultTime uint
	}
	tests := []struct {
		name    string
		args    args
		want    time.Duration
		wantErr bool
	}{
		{
			name:    "Default value",
			args:    args{"", 5},
			want:    5 * time.Second,
			wantErr: false,
		},
		{
			args:    args{"42", 5},
			want:    42 * time.Second,
			wantErr: false,
		},
		{
			name:    "Wrong string",
			args:    args{"z42", 5},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := stringToTime(tt.args.envVar, tt.args.defaultTime)
			if (err != nil) != tt.wantErr {
				t.Errorf("stringToTime() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("stringToTime() = %v, want %v", got, tt.want)
			}
		})
	}
}
