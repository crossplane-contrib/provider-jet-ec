package clients

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/muvaf/typewriter/pkg/test"
	"github.com/pkg/errors"
)

func TestGetAuthEnvVars(t *testing.T) {
	type args struct {
		ecCreds map[string]string
	}
	type want struct {
		envvars []string
		err     error
	}
	cases := map[string]struct {
		args args
		want want
	}{
		"UsernameAndPasswordOnly": {
			args: args{
				ecCreds: map[string]string{
					keyUsername: "admin",
					keyPassword: "s3cr3t",
				},
			},
			want: want{
				envvars: []string{
					"EC_USERNAME=admin",
					"EC_PASSWORD=s3cr3t",
				},
			},
		},
		"APIKeyOnly": {
			args: args{
				ecCreds: map[string]string{
					keyAPIKey: "12394871280347073qwqsdadad",
				},
			},
			want: want{
				envvars: []string{
					"EC_API_KEY=12394871280347073qwqsdadad",
				},
			},
		},
		"BothUsernamePasswordAndAPIKey": {
			args: args{
				ecCreds: map[string]string{
					keyUsername: "admin",
					keyPassword: "s3cr3t",
					keyAPIKey:   "12394871280347073qwqsdadad",
				},
			},
			want: want{
				err: errors.New(errEitherAPIKeyOrUP),
			},
		},
		"BothUsernameAndPassword": {
			args: args{
				ecCreds: map[string]string{
					keyUsername: "admin",
				},
			},
			want: want{
				err: errors.New(errUnameAndPword),
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {

			got, err := getAuthEnvVars(tc.args.ecCreds)

			if diff := cmp.Diff(tc.want.envvars, got); diff != "" {
				t.Errorf("GetAuthEnvVars(...): -want, +got:\n%s", diff)
			}

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("GetAuthEnvVars(...): -want, +got:\n%s", diff)
			}
		})
	}
}
