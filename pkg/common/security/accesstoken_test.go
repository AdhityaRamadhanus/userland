// +build unit

package security_test

import (
	"testing"
	"time"

	"github.com/AdhityaRamadhanus/userland"
	"github.com/AdhityaRamadhanus/userland/pkg/common/security"
)

func TestCreateAccessToken(t *testing.T) {
	type args struct {
		user userland.User
		opt  security.AccessTokenOptions
	}
	testCases := []struct {
		name    string
		args    args
		wantErr error
	}{
		{
			name: "success",
			args: args{
				user: userland.User{
					Fullname: "Adhitya Ramadhanus",
					Email:    "adhitya.ramadhanus@gmail.com",
					ID:       1,
				},
				opt: security.AccessTokenOptions{
					Expiration: 60 * time.Second * 5,
					Scope:      security.UserTokenScope,
				},
			},
			wantErr: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := security.CreateAccessToken(tc.args.user, "jwtsecret_test", tc.args.opt); err != tc.wantErr {
				t.Fatalf("security.CreateAccessToken() err = %v; want %v", err, tc.wantErr)
			}
		})
	}
}
