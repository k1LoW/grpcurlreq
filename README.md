# grpcurlreq

`grpcurlreq` is parser for [gRPCurl](https://github.com/fullstorydev/grpcurl) command.

## Usage

``` go
package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/k1LoW/grpcurlreq"
)

func main() {
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
```

