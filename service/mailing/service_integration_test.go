// +build all mailing_service

package mailing_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/gocraft/work"
	"github.com/gomodule/redigo/redis"
	"github.com/sarulabs/di"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	"github.com/AdhityaRamadhanus/userland/service/mailing"
	"github.com/stretchr/testify/suite"
)

type MailingServiceTestSuite struct {
	suite.Suite
	RedisPool      *redis.Pool
	Enqueuer       *work.Enqueuer
	MailingService mailing.Service
}

func (suite *MailingServiceTestSuite) BuildContainer() di.Container {
	builder, _ := di.NewBuilder()
	builder.Add(
		di.Def{
			Name:  "redis-pool-connection",
			Scope: di.App,
			Build: func(ctn di.Container) (interface{}, error) {
				redisAddr := fmt.Sprintf("%s:%s", os.Getenv("TEST_REDIS_HOST"), os.Getenv("TEST_REDIS_PORT"))
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
				return redisPool, nil
			},
		},
		di.Def{
			Name:  "work-enqueuer",
			Scope: di.App,
			Build: func(ctn di.Container) (interface{}, error) {
				workerNamespace := "userland-mail-worker"
				redisPool := ctn.Get("redis-pool-connection").(*redis.Pool)
				enqueuer := work.NewEnqueuer(workerNamespace, redisPool)
				return enqueuer, nil
			},
		},
		mailing.ServiceBuilder,
		mailing.ServiceInstrumentorBuilder,
	)

	return builder.Build()
}

// before each test
func (suite *MailingServiceTestSuite) SetupSuite() {
	godotenv.Load("../../.env")
	os.Setenv("ENV", "testing")

	ctn := suite.BuildContainer()
	suite.Enqueuer = ctn.Get("work-enqueuer").(*work.Enqueuer)
	suite.MailingService = ctn.Get("mailing-instrumentor-service").(mailing.Service)
}

func TestMailingService(t *testing.T) {
	suiteTest := new(MailingServiceTestSuite)
	suite.Run(t, suiteTest)
}

func (suite *MailingServiceTestSuite) TestSendOTPEmail() {
	testCases := []struct {
		Recipient mailing.MailAddress
		OTP       string
		OTPType   string
	}{
		{
			Recipient: mailing.MailAddress{
				Name:    "test",
				Address: "test@coba.com",
			},
			OTP:     "545453",
			OTPType: "Login TFA",
		},
	}

	for _, testCase := range testCases {
		err := suite.MailingService.SendOTPEmail(testCase.Recipient, testCase.OTPType, testCase.OTP)
		suite.Nil(err, "should send otp email")
	}
}

func (suite *MailingServiceTestSuite) TestSendVerificationEmail() {
	testCases := []struct {
		Recipient        mailing.MailAddress
		VerificationLink string
	}{
		{
			Recipient: mailing.MailAddress{
				Name:    "test",
				Address: "test@coba.com",
			},
			VerificationLink: "545453",
		},
	}

	for _, testCase := range testCases {
		err := suite.MailingService.SendVerificationEmail(testCase.Recipient, testCase.VerificationLink)
		suite.Nil(err, "should send verification email")
	}
}
