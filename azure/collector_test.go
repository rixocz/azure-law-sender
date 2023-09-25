package azure

import (
	"net/http"
	"reflect"
	"strings"
	"testing"
	"time"
)

func Test_hashHmac256b64enc(t *testing.T) {
	type args struct {
		data   string
		b64Key string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name:    "test",
			args:    args{"secret", "c2VjcmV0S2V5"},
			want:    "PZP97b04dXXwN//a+TfezpeUcHfACGVMPr+GK7PKIYQ=",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := hashHmac256b64enc(tt.args.data, tt.args.b64Key)
			if (err != nil) != tt.wantErr {
				t.Errorf("hashHmac256b64enc() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("hashHmac256b64enc() got = %s, want %s", got, tt.want)
			}
		})
	}
}

func Test_createReqSignature(t *testing.T) {
	req, _ := http.NewRequest("POST", "https://localhost", strings.NewReader("test body"))
	timestamp := time.Date(2000, 1, 1, 1, 1, 1, 0, time.FixedZone("GMT", 0))
	req.Header.Set("X-Ms-Date", timestamp.Format(time.RFC1123))
	type args struct {
		req    *http.Request
		b64Key string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "test",
			args: args{
				req:    req,
				b64Key: "c2VjcmV0S2V5",
			},
			want:    "guCIOn+zxkkTHsERFyk2p9S9PcY+IrdyHIKO/UDxh00=",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := createReqSignature(tt.args.req, tt.args.b64Key)
			if (err != nil) != tt.wantErr {
				t.Errorf("createReqSignature() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("createReqSignature() got = %v, want %v", got, tt.want)
			}
		})
	}
}
