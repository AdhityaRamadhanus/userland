// +build all common unit

package security_test

import (
	"testing"

	"github.com/AdhityaRamadhanus/userland/pkg/common/security"
)

func TestGenerateOTP(t *testing.T) {
	type args struct {
		length int
	}
	testCases := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "len=6",
			args: args{
				length: 6,
			},
			wantErr: false,
		},
		{
			name: "len=2",
			args: args{
				length: 2,
			},
			wantErr: false,
		},
		{
			name: "len=-1",
			args: args{
				length: -1,
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			code, err := security.GenerateOTP(tc.args.length)
			if err != nil && !tc.wantErr {
				t.Fatalf("security.GenerateOTP() err = %v; want nil", err)
			}

			if err == nil && tc.wantErr {
				t.Fatal("security.GenerateOTP() err = nil; want not nil", err)
			}

			if tc.wantErr {
				return
			}

			if len(code) != tc.args.length {
				t.Errorf("security.GenerateOTP() len(code) = %d; want %d", len(code), tc.args.length)
			}
		})
	}
}
