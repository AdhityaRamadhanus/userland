// +build integration

package mailing_test

import (
	"fmt"
	"testing"

	"github.com/AdhityaRamadhanus/userland/pkg/common/metrics"
	"github.com/AdhityaRamadhanus/userland/pkg/config"
	"github.com/AdhityaRamadhanus/userland/pkg/service/mailing"
	"github.com/gocraft/work"
	"github.com/gomodule/redigo/redis"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/suite"
)

type MailingServiceTestSuite struct {
	suite.Suite
	Config         *config.Configuration
	RedisPool      *redis.Pool
	Enqueuer       *work.Enqueuer
	MailingService mailing.Service
}

func NewMailingServiceTestSuite(cfg *config.Configuration) *MailingServiceTestSuite {
	return &MailingServiceTestSuite{
		Config: cfg,
	}
}

func (suite *MailingServiceTestSuite) Teardown() {
	suite.T().Log("Teardown MailingServiceTestSuite")
	suite.RedisPool.Close()
}

// before each test
func (suite *MailingServiceTestSuite) SetupSuite() {
	redisAddr := fmt.Sprintf("%s:%d", suite.Config.Redis.Host, suite.Config.Redis.Port)
	suite.T().Logf("Connecting to redis at %s", redisAddr)
	redisPool := &redis.Pool{
		MaxActive: 5,
		MaxIdle:   5,
		Wait:      true,
		Dial: func() (redis.Conn, error) {
			return redis.Dial(
				"tcp",
				redisAddr,
				redis.DialDatabase(1),
			)
		},
	}

	suite.RedisPool = redisPool
	suite.Enqueuer = work.NewEnqueuer(suite.Config.Mail.WorkerSpace, redisPool)
	suite.MailingService = mailing.NewService(suite.Config.Mail.Queue, suite.Config.Mail.Sender, suite.Enqueuer)
	suite.MailingService = mailing.NewInstrumentorService(metrics.PrometheusRequestLatency("service", "mailing", mailing.MetricKeys), suite.MailingService)
}

func (suite MailingServiceTestSuite) TestSendOTPEmail() {
	type args struct {
		recipient mailing.MailAddress
		otp       string
		otpType   string
	}
	testCases := []struct {
		name    string
		args    args
		wantErr error
	}{
		{
			name: "success",
			args: args{
				recipient: mailing.MailAddress{
					Name:    "test",
					Address: "test@coba.com",
				},
				otp:     "545453",
				otpType: "Login TFA",
			},
			wantErr: nil,
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			if err := suite.MailingService.SendOTPEmail(tc.args.recipient, tc.args.otp, tc.args.otpType); err != tc.wantErr {
				t.Fatalf("MailingService.SendOTPEmail() err = %v; want %v", err, tc.wantErr)
			}
		})
	}
}

func (suite MailingServiceTestSuite) TestSendVerificationEmail() {
	type args struct {
		recipient        mailing.MailAddress
		verificationLink string
	}
	testCases := []struct {
		name    string
		args    args
		wantErr error
	}{
		{
			name: "success",
			args: args{
				recipient: mailing.MailAddress{
					Name:    "test",
					Address: "test@coba.com",
				},
				verificationLink: "http://www.coba.com/email",
			},
			wantErr: nil,
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			if err := suite.MailingService.SendVerificationEmail(tc.args.recipient, tc.args.verificationLink); err != tc.wantErr {
				t.Fatalf("MailingService.SendVerificationEmail() err = %v; want %v", err, tc.wantErr)
			}
		})
	}
}
