package grpcurlreq

import (
	"bufio"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/fullstorydev/grpcurl"
	"github.com/mattn/go-shellwords"
	"google.golang.org/grpc/metadata"
)

const (
	stateBlank  = ""
	stateHeader = "header"
	stateUA     = "user-agent"
	stateData   = "data"
)

type Parsed struct {
	SubCmd   string                   `json:"cmd,omitempty"`
	Addr     string                   `json:"addr,omitempty"`
	Headers  metadata.MD              `json:"headers,omitempty"`
	Method   string                   `json:"method,omitempty"`
	Messages []map[string]interface{} `json:"messages,omitempty"`
}

var addrRe = regexp.MustCompile(`:[0-9]+$`)
var methodRe = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_./]+$`)

// Parse a grpcurl command.
func Parse(cmd ...string) (*Parsed, error) {
	args, err := cmdToArgs(cmd...)
	if err != nil {
		return nil, err
	}
	out := newParsed()
	state := stateBlank
	headers := []string{}

	for _, a := range args {
		switch {
		case a == "-H" || a == "--rpc-header":
			state = stateHeader
		case a == "--user-agent":
			state = stateUA
		case a == "-d":
			state = stateData
		case a == "list" || a == "describe":
			out.SubCmd = a
		case state == stateBlank && addrRe.MatchString(a):
			out.Addr = a
		case state == stateBlank && methodRe.MatchString(a):
			out.Method = a
		case a != "" && !strings.HasPrefix(a, "-"):
			switch state {
			case stateHeader:
				headers = append(headers, a)
				state = stateBlank
			case stateUA:
				headers = append(headers, fmt.Sprintf("user-agent: %s", a))
				state = stateBlank
			case stateData:
				m, err := toMessages(a)
				if err != nil {
					return nil, err
				}
				out.Messages = m
				state = stateBlank
			}
		}
	}

	out.Headers = grpcurl.MetadataFromHeaders(headers)

	return out, nil
}

func cmdToArgs(cmd ...string) ([]string, error) {
	var err error
	if len(cmd) == 1 {
		cmd, err = shellwords.Parse(cmd[0])
		if err != nil {
			return nil, err
		}
	}
	if cmd[0] != "grpcurl" {
		return nil, fmt.Errorf("invalid grpcurl command: %s", cmd)
	}
	if len(cmd) == 1 {
		return nil, fmt.Errorf("invalid grpcurl command: %s", cmd)
	}
	return cmd[1:], nil
}

func toMessages(in string) ([]map[string]interface{}, error) {
	const (
		delimStart = '{'
		delimEnd   = '}'
	)
	sf := func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		c := 0
		for i := 0; i < len(data); i++ {
			if data[i] == delimStart {
				c += 1
				continue
			}
			if data[i] == delimEnd {
				c -= 1
				if c == 0 {
					return i + 1, data[:i+1], nil
				}
			}
		}
		if atEOF {
			return 0, data, bufio.ErrFinalToken
		}
		return 0, nil, nil
	}

	scanner := bufio.NewScanner(strings.NewReader(in))
	scanner.Split(sf)
	messages := []map[string]interface{}{}
	for scanner.Scan() {
		s := strings.TrimSpace(scanner.Text())
		if s == "" {
			continue
		}
		m := map[string]interface{}{}
		if err := json.Unmarshal([]byte(s), &m); err != nil {
			return nil, err
		}
		messages = append(messages, m)
	}
	return messages, nil
}

func newParsed() *Parsed {
	return &Parsed{
		Headers: metadata.MD{},
	}
}
