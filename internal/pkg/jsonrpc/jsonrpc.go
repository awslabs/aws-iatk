package jsonrpc

import (
	"bytes"
	"encoding/json"
	"log"
	"regexp"
	"strings"

	"golang.org/x/exp/slices"
)

type Request struct {
	// JSONRPC describes the version of the JSON RPC protocol. Defaults to `2.0`.
	JSONRPC string `json:"jsonrpc"`
	// ID identifies a unique request.
	ID *string `json:"id"`
	// Method describes the intention of the request.
	Method string `json:"method"`
	// Params contains the payload of the request. Usually parsed into a specific struct for further processing.
	Params json.RawMessage `json:"params"`

	// Metadata contains information about the client
	Metadata *Metadata `json:"metadata"`
}

type Metadata struct {
	Client        string `json:"client"`
	Version       string `json:"version"`
	Caller        string `json:"caller"`
	ClientVersion string `json:"client_version"`
}

// Returns a string that serves as a user agent key, indicating the client version and the caller method, e.g. 0.0.3#retry_get_trace_tree_until
func (m *Metadata) UserAgentValue() string {
	supportedClients := []string{"python"}
	if !slices.Contains(supportedClients, strings.ToLower(m.Client)) {
		log.Printf("unregonized client: %v", strings.ToLower(m.Client))
		return "unknown"
	}

	// NOTE: pattern from https://semver.org/#is-there-a-suggested-regular-expression-regex-to-check-a-semver-string
	match, err := regexp.Match(`^(?P<major>0|[1-9]\d*)\.(?P<minor>0|[1-9]\d*)\.(?P<patch>0|[1-9]\d*)(?:-(?P<prerelease>(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+(?P<buildmetadata>[0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$`, []byte(m.Version))
	if err != nil || !match {
		log.Printf("invalid version: %v", m.Version)
		return "unknown"
	}

	// NOTE: limit caller to has max length of 100 characters, and limit to alphabetical characters and . and _ only
	match, err = regexp.Match(`^[a-zA-Z_][a-zA-Z0-9_.]{1,99}$`, []byte(m.Caller))
	if err != nil || !match {
		log.Printf("invalid caller: %v", m.Caller)
		return "unknown"
	}
	return strings.ToLower(m.Client) + "#" + m.ClientVersion + "#" + m.Version + "#" + m.Caller
}

type Response struct {
	// JSONRPC describes the version of the JSON RPC protocol. Defaults to `2.0`.
	JSONRPC string `json:"jsonrpc"`
	// ID identifies a unique request.
	ID *string `json:"id"`
	// Result contains the payload of the response.
	Result interface{} `json:"result,omitempty"`
	// Error contains the error response details.
	Error *ErrIatk `json:"error,omitempty"`
}

type ErrIatk struct {
	Code    int    `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
	// Data should be here but not sure about the type (jfuss)
}

func (r Response) Encode() ([]byte, error) {
	var buffer bytes.Buffer
	enc := json.NewEncoder(&buffer)
	if err := enc.Encode(r); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}
