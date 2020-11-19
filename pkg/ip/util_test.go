package ip

import "testing"

func TestSplitWithMask(t *testing.T) {
	type args struct {
		ip   string
		mask uint
	}
	tests := []struct {
		name    string
		args    args
		want    uint32
		want1   uint32
		wantErr bool
	}{
		{name: "normal",
			args: args{
				ip:   "192.168.233.233",
				mask: 24,
			},
			want:    3232295168,
			want1:   233,
			wantErr: false,
		},
		{name: "normal",
			args: args{
				ip:   "10.111.23.7",
				mask: 12,
			},
			want:    174063616,
			want1:   988935,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := SplitWithMask(tt.args.ip, tt.args.mask)
			if (err != nil) != tt.wantErr {
				t.Errorf("SplitWithMask() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("SplitWithMask() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("SplitWithMask() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
