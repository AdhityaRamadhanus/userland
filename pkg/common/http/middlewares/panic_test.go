//+build unit

package middlewares_test

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AdhityaRamadhanus/userland/pkg/common/http/middlewares"
	"github.com/pkg/errors"
)

func TestPanic(t *testing.T) {
	type args struct {
		handler http.Handler
	}
	testCases := []struct {
		name           string
		args           args
		wantStatusCode int
	}{
		{
			name: "return panic with pkg errors",
			args: args{
				handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					panic(errors.New("Panic"))
				}),
			},
			wantStatusCode: http.StatusInternalServerError,
		},
		{
			name: "return OK",
			args: args{
				handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte("OK"))
				}),
			},
			wantStatusCode: http.StatusOK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mw := middlewares.PanicHandler(tc.args.handler)
			req, err := http.NewRequest(http.MethodGet, "/", nil)
			if err != nil {
				t.Fatalf("http.NewRequest() err = %v; want nil", err)
			}
			res := httptest.NewRecorder()
			mw.ServeHTTP(res, req)

			statusCode := res.Result().StatusCode
			if statusCode != tc.wantStatusCode {
				body, _ := ioutil.ReadAll(res.Result().Body)
				defer res.Result().Body.Close()
				t.Logf("response %s\n", string(body))
				t.Errorf("middlewares.PanicHandler() res.StatusCode = %d; want %d", statusCode, tc.wantStatusCode)
			}
		})
	}
}
