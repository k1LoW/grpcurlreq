package grpcurlreq_test

import (
	"encoding/json"
	"fmt"
	"log"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/k1LoW/grpcurlreq"
	"google.golang.org/grpc/metadata"
)

func TestParse(t *testing.T) {
	tests := []struct {
		input string
		want  *grpcurlreq.Parsed
	}{
		{
			`grpcurl grpc.server.com:443 my.custom.server.Service/Method`,
			&grpcurlreq.Parsed{
				Addr:    "grpc.server.com:443",
				Method:  "my.custom.server.Service/Method",
				Headers: metadata.MD{},
			},
		},
		{
			`grpcurl -plaintext grpc.server.com:80 my.custom.server.Service/Method`,
			&grpcurlreq.Parsed{
				Addr:    "grpc.server.com:80",
				Method:  "my.custom.server.Service/Method",
				Headers: metadata.MD{},
			},
		},
		{
			`grpcurl -d '{"id": 1234, "tags": ["foo","bar"]}' grpc.server.com:443 my.custom.server.Service/Method`,
			&grpcurlreq.Parsed{
				Addr:    "grpc.server.com:443",
				Method:  "my.custom.server.Service/Method",
				Headers: metadata.MD{},
				Messages: []map[string]interface{}{
					{
						"id":   float64(1234),
						"tags": []interface{}{"foo", "bar"},
					},
				},
			},
		},
		{
			`grpcurl -d '{"id": 1234, "tags": ["foo","bar"]}{"id": 2345, "tags": ["bar","baz"]}' grpc.server.com:443 my.custom.server.Service/Method`,
			&grpcurlreq.Parsed{
				Addr:    "grpc.server.com:443",
				Method:  "my.custom.server.Service/Method",
				Headers: metadata.MD{},
				Messages: []map[string]interface{}{
					{
						"id":   float64(1234),
						"tags": []interface{}{"foo", "bar"},
					},
					{
						"id":   float64(2345),
						"tags": []interface{}{"bar", "baz"},
					},
				},
			},
		},
		{
			`grpcurl -d '{"id": 1234, "tags": ["foo","bar"]}  {"id": 2345, "tags": ["bar","baz"]}' grpc.server.com:443 my.custom.server.Service/Method`,
			&grpcurlreq.Parsed{
				Addr:    "grpc.server.com:443",
				Method:  "my.custom.server.Service/Method",
				Headers: metadata.MD{},
				Messages: []map[string]interface{}{
					{
						"id":   float64(1234),
						"tags": []interface{}{"foo", "bar"},
					},
					{
						"id":   float64(2345),
						"tags": []interface{}{"bar", "baz"},
					},
				},
			},
		},
		{
			`grpcurl -d '{"id": 1234, "tags": {"foo":"bar"}}{"id": 2345, "tags": {"bar":"baz"}}' grpc.server.com:443 my.custom.server.Service/Method`,
			&grpcurlreq.Parsed{
				Addr:    "grpc.server.com:443",
				Method:  "my.custom.server.Service/Method",
				Headers: metadata.MD{},
				Messages: []map[string]interface{}{
					{
						"id": float64(1234),
						"tags": map[string]interface{}{
							"foo": "bar",
						},
					},
					{
						"id": float64(2345),
						"tags": map[string]interface{}{
							"bar": "baz",
						},
					},
				},
			},
		},
		{
			`grpcurl localhost:8787 list`,
			&grpcurlreq.Parsed{
				SubCmd:  "list",
				Addr:    "localhost:8787",
				Headers: metadata.MD{},
			},
		},
		{
			`grpcurl -import-path ../protos -proto my-stuff.proto list`,
			&grpcurlreq.Parsed{
				SubCmd:  "list",
				Headers: metadata.MD{},
			},
		},
		{
			`grpcurl localhost:8787 list my.custom.server.Service`,
			&grpcurlreq.Parsed{
				SubCmd:  "list",
				Addr:    "localhost:8787",
				Method:  "my.custom.server.Service",
				Headers: metadata.MD{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := grpcurlreq.Parse(tt.input)
			if err != nil {
				t.Error(err)
			}
			if diff := cmp.Diff(got, tt.want, nil); diff != "" {
				t.Errorf("%s", diff)
			}
		})
	}
}

func Example() {
	cmd := `grpcurl -d '{"id": 1234, "tags": ["foo","bar"]}' grpc.server.com:443 my.custom.server.Service/Method`
	p, err := grpcurlreq.Parse(cmd)
	if err != nil {
		log.Fatal(err)
	}
	b, err := json.Marshal(p)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(b))

	// Output:
	// {"addr":"grpc.server.com:443","method":"my.custom.server.Service/Method","messages":[{"id":1234,"tags":["foo","bar"]}]}
}
