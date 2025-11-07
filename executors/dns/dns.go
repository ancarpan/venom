package dns

import (
	"context"
	"fmt"
	"time"

	"github.com/miekg/dns"
	"github.com/mitchellh/mapstructure"
	"github.com/ovh/venom"
)

// Name of executor
const Name = "dns"

// New returns a new Executor
func New() venom.Executor {
	return &Executor{}
}

// Executor represents a DNS Test Exec
type Executor struct {
	Server  string `json:"server,omitempty" yaml:"server,omitempty"`
	Query   string `json:"query,omitempty" yaml:"query,omitempty"`
	QType   string `json:"qtype,omitempty" yaml:"qtype,omitempty"`
	Timeout int    `json:"timeout,omitempty" yaml:"timeout,omitempty"`
}

// Result represents a step result
type Result struct {
	Query         string      `json:"query,omitempty" yaml:"query,omitempty"`
	QType         string      `json:"qtype,omitempty" yaml:"qtype,omitempty"`
	Server        string      `json:"server,omitempty" yaml:"server,omitempty"`
	RCode         string      `json:"rcode,omitempty" yaml:"rcode,omitempty"`
	Message       interface{} `json:"message,omitempty" yaml:"message,omitempty"`
	Systemout     string      `json:"systemout,omitempty" yaml:"systemout,omitempty"`
	SystemoutJSON interface{} `json:"systemoutjson,omitempty" yaml:"systemoutjson,omitempty"`
	Systemerr     string      `json:"systemerr,omitempty" yaml:"systemerr,omitempty"`
	SystemerrJSON interface{} `json:"systemerrjson,omitempty" yaml:"systemerrjson,omitempty"`
	Err           string      `json:"err,omitempty" yaml:"err,omitempty"`
	TimeSeconds   float64     `json:"timeseconds,omitempty" yaml:"timeseconds,omitempty"`
}

// ZeroValueResult return an empty implementation of this executor result
func (Executor) ZeroValueResult() interface{} {
	return Result{}
}

// GetDefaultAssertions return default assertions for type dns
func (Executor) GetDefaultAssertions() *venom.StepAssertions {
	return &venom.StepAssertions{Assertions: []venom.Assertion{"result.err ShouldBeEmpty"}}
}

// Run execute TestStep of type dns
func (Executor) Run(ctx context.Context, step venom.TestStep) (interface{}, error) {
	var e Executor
	if err := mapstructure.Decode(step, &e); err != nil {
		return nil, err
	}

	// Server is mandatory
	if e.Server == "" {
		return nil, fmt.Errorf("server is mandatory for DNS executor")
	}

	// Set defaults
	if e.QType == "" {
		e.QType = "A"
	}
	if e.Timeout == 0 {
		e.Timeout = 5
	}

	result := Result{
		Query:  e.Query,
		QType:  e.QType,
		Server: e.Server,
	}
	start := time.Now()

	// Parse DNS record type
	qType, err := stringToQType(e.QType)
	if err != nil {
		result.Err = err.Error()
		elapsed := time.Since(start)
		result.TimeSeconds = elapsed.Seconds()
		return result, nil
	}

	// Create DNS message
	m := new(dns.Msg)
	m.SetQuestion(dns.Fqdn(e.Query), qType)
	m.RecursionDesired = true

	// Create DNS client
	client := new(dns.Client)
	timeout := time.Duration(e.Timeout) * time.Second
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Try UDP first
	client.Net = "udp"
	response, rtt, err := client.ExchangeContext(ctx, m, e.Server)
	if err != nil {
		result.Err = err.Error()
		elapsed := time.Since(start)
		result.TimeSeconds = elapsed.Seconds()
		return result, nil
	}

	// If response is truncated, retry with TCP
	if response.Truncated {
		// Create a new message for TCP retry (important: reset the message ID)
		mTCP := new(dns.Msg)
		mTCP.SetQuestion(dns.Fqdn(e.Query), qType)
		mTCP.RecursionDesired = true

		// Create a new TCP client
		tcpClient := new(dns.Client)
		tcpClient.Net = "tcp"
		tcpResponse, tcpRtt, tcpErr := tcpClient.ExchangeContext(ctx, mTCP, e.Server)
		if tcpErr != nil {
			// If TCP retry fails, return the truncated UDP response with error info
			result.Err = fmt.Sprintf("UDP response truncated, TCP retry failed: %v", tcpErr)
			// Continue to return the truncated UDP response so user can see what we got
		} else if tcpResponse.Truncated {
			// TCP response is also truncated (shouldn't happen, but handle it)
			result.Err = "UDP response truncated, TCP retry also returned truncated response"
		} else {
			// TCP retry succeeded - use the TCP response
			response = tcpResponse
			rtt = tcpRtt
		}
	}

	elapsed := time.Since(start)
	result.TimeSeconds = elapsed.Seconds()

	// Convert response code
	result.RCode = dns.RcodeToString[response.Rcode]

	// Convert full DNS message to JSON structure
	msgJSON, err := dnsMessageToJSON(response)
	if err != nil {
		result.Err = fmt.Sprintf("failed to convert DNS message to JSON: %v", err)
		return result, nil
	}
	result.Message = msgJSON
	result.SystemoutJSON = msgJSON

	// Build human-readable system output
	result.Systemout = fmt.Sprintf("DNS Query: %s %s\nServer: %s\nRCode: %s\nResponse Time: %v\n",
		e.Query, e.QType, e.Server, result.RCode, rtt)

	if len(response.Answer) > 0 {
		result.Systemout += "Answers:\n"
		for _, rr := range response.Answer {
			result.Systemout += fmt.Sprintf("  %s\n", rr.String())
		}
	}

	return result, nil
}

// stringToQType converts DNS record type string to dns.Type
func stringToQType(qtype string) (uint16, error) {
	switch qtype {
	case "A":
		return dns.TypeA, nil
	case "AAAA":
		return dns.TypeAAAA, nil
	case "MX":
		return dns.TypeMX, nil
	case "TXT":
		return dns.TypeTXT, nil
	case "CNAME":
		return dns.TypeCNAME, nil
	case "NS":
		return dns.TypeNS, nil
	case "PTR":
		return dns.TypePTR, nil
	case "SOA":
		return dns.TypeSOA, nil
	case "SRV":
		return dns.TypeSRV, nil
	case "CAA":
		return dns.TypeCAA, nil
	case "ANY":
		return dns.TypeANY, nil
	default:
		return 0, fmt.Errorf("unsupported DNS record type: %s", qtype)
	}
}

// dnsMessageToJSON converts a DNS message to a JSON-serializable structure
func dnsMessageToJSON(msg *dns.Msg) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	result["id"] = msg.Id
	result["response"] = msg.Response
	result["opcode"] = dns.OpcodeToString[msg.Opcode]
	result["authoritative"] = msg.Authoritative
	result["truncated"] = msg.Truncated
	result["recursion_desired"] = msg.RecursionDesired
	result["recursion_available"] = msg.RecursionAvailable
	result["zero"] = msg.Zero
	result["authenticated_data"] = msg.AuthenticatedData
	result["checking_disabled"] = msg.CheckingDisabled
	result["rcode"] = dns.RcodeToString[msg.Rcode]

	// Convert questions
	questions := make([]map[string]interface{}, 0, len(msg.Question))
	for _, q := range msg.Question {
		questions = append(questions, map[string]interface{}{
			"name":  q.Name,
			"type":  dns.TypeToString[q.Qtype],
			"class": dns.ClassToString[q.Qclass],
		})
	}
	result["question"] = questions

	// Convert answers
	answers := make([]map[string]interface{}, 0, len(msg.Answer))
	for _, rr := range msg.Answer {
		rrJSON := rrToJSON(rr)
		answers = append(answers, rrJSON)
	}
	result["answer"] = answers

	// Convert authority records
	authority := make([]map[string]interface{}, 0, len(msg.Ns))
	for _, rr := range msg.Ns {
		rrJSON := rrToJSON(rr)
		authority = append(authority, rrJSON)
	}
	result["authority"] = authority

	// Convert additional records
	additional := make([]map[string]interface{}, 0, len(msg.Extra))
	for _, rr := range msg.Extra {
		rrJSON := rrToJSON(rr)
		additional = append(additional, rrJSON)
	}
	result["additional"] = additional

	return result, nil
}

// rrToJSON converts a DNS resource record to JSON
func rrToJSON(rr dns.RR) map[string]interface{} {
	result := map[string]interface{}{
		"name":  rr.Header().Name,
		"type":  dns.TypeToString[rr.Header().Rrtype],
		"class": dns.ClassToString[rr.Header().Class],
		"ttl":   rr.Header().Ttl,
	}

	// Convert RR to string and add to result
	result["value"] = rr.String()

	// Add type-specific fields
	switch v := rr.(type) {
	case *dns.A:
		result["address"] = v.A.String()
	case *dns.AAAA:
		result["address"] = v.AAAA.String()
	case *dns.MX:
		result["preference"] = v.Preference
		result["mx"] = v.Mx
	case *dns.TXT:
		result["txt"] = v.Txt
	case *dns.CNAME:
		result["target"] = v.Target
	case *dns.NS:
		result["ns"] = v.Ns
	case *dns.PTR:
		result["ptr"] = v.Ptr
	case *dns.SOA:
		result["ns"] = v.Ns
		result["mbox"] = v.Mbox
		result["serial"] = v.Serial
		result["refresh"] = v.Refresh
		result["retry"] = v.Retry
		result["expire"] = v.Expire
		result["minttl"] = v.Minttl
	case *dns.SRV:
		result["priority"] = v.Priority
		result["weight"] = v.Weight
		result["port"] = v.Port
		result["target"] = v.Target
	case *dns.CAA:
		result["flag"] = v.Flag
		result["tag"] = v.Tag
		result["value"] = v.Value
	}

	return result
}
