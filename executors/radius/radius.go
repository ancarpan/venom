package radius

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/mitchellh/mapstructure"
	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
	"layeh.com/radius/rfc2866"

	"github.com/ovh/venom"
)

// Name of executor
const Name = "radius"

// New returns a new Executor
func New() venom.Executor {
	return &Executor{}
}

// Executor represents a Test Exec
type Executor struct {
	Server     string            `json:"server,omitempty" yaml:"server,omitempty"`
	Secret     string            `json:"secret,omitempty" yaml:"secret,omitempty"`
	Code       string            `json:"code,omitempty" yaml:"code,omitempty"`
	Timeout    int               `json:"timeout,omitempty" yaml:"timeout,omitempty"`
	Attributes map[string]string `json:"attributes,omitempty" yaml:"attributes,omitempty"`
}

// Result represents a step result
type Result struct {
	Systemout     string                 `json:"systemout,omitempty" yaml:"systemout,omitempty"`
	SystemoutJSON interface{}            `json:"systemoutjson,omitempty" yaml:"systemoutjson,omitempty"`
	Systemerr     string                 `json:"systemerr,omitempty" yaml:"systemerr,omitempty"`
	SystemerrJSON interface{}            `json:"systemerrjson,omitempty" yaml:"systemerrjson,omitempty"`
	Err           string                 `json:"err,omitempty" yaml:"err,omitempty"`
	Code          string                 `json:"code,omitempty" yaml:"code,omitempty"`
	TimeSeconds   float64                `json:"timeseconds,omitempty" yaml:"timeseconds,omitempty"`
	Attributes    map[string]interface{} `json:"attributes,omitempty" yaml:"attributes,omitempty"`
	Request       map[string]interface{} `json:"request,omitempty" yaml:"request,omitempty"`
}

// ZeroValueResult return an empty implementation of this executor result
func (Executor) ZeroValueResult() interface{} {
	return Result{}
}

// GetDefaultAssertions return default assertions for type exec
func (Executor) GetDefaultAssertions() *venom.StepAssertions {
	return &venom.StepAssertions{Assertions: []venom.Assertion{"result.err ShouldBeEmpty"}}
}

// Run execute TestStep of type exec
func (Executor) Run(ctx context.Context, step venom.TestStep) (interface{}, error) {
	var e Executor
	if err := mapstructure.Decode(step, &e); err != nil {
		return nil, err
	}

	// Set defaults
	if e.Server == "" {
		e.Server = "localhost:1812"
	}
	if e.Secret == "" {
		e.Secret = "secret"
	}
	if e.Code == "" {
		e.Code = "Access-Request"
	}
	if e.Timeout == 0 {
		e.Timeout = 5
	}

	result := Result{}
	start := time.Now()

	// Parse RADIUS code
	var code radius.Code
	switch e.Code {
	case "Access-Request":
		code = radius.CodeAccessRequest
	case "Access-Accept":
		code = radius.CodeAccessAccept
	case "Access-Reject":
		code = radius.CodeAccessReject
	case "Accounting-Request":
		code = radius.CodeAccountingRequest
	case "Accounting-Response":
		code = radius.CodeAccountingResponse
	case "Access-Challenge":
		code = radius.CodeAccessChallenge
	case "Status-Server":
		code = radius.CodeStatusServer
	case "Status-Client":
		code = radius.CodeStatusClient
	case "Disconnect-Request":
		code = radius.CodeDisconnectRequest
	case "Disconnect-ACK":
		code = radius.CodeDisconnectACK
	case "Disconnect-NAK":
		code = radius.CodeDisconnectNAK
	case "CoA-Request":
		code = radius.CodeCoARequest
	case "CoA-ACK":
		code = radius.CodeCoAACK
	case "CoA-NAK":
		code = radius.CodeCoANAK
	default:
		return nil, fmt.Errorf("unsupported RADIUS code: %s", e.Code)
	}

	// Create RADIUS packet
	packet := radius.New(code, []byte(e.Secret))

	// Add attributes
	for attrName, attrValue := range e.Attributes {
		if err := addAttribute(packet, attrName, attrValue); err != nil {
			result.Err = fmt.Sprintf("failed to add attribute %s: %v", attrName, err)
			return result, nil
		}
	}

	// Store request info
	result.Request = map[string]interface{}{
		"server":     e.Server,
		"code":       e.Code,
		"attributes": e.Attributes,
	}

	// Send packet
	timeout := time.Duration(e.Timeout) * time.Second
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	response, err := radius.Exchange(ctx, packet, e.Server)
	if err != nil {
		result.Err = err.Error()
		elapsed := time.Since(start)
		result.TimeSeconds = elapsed.Seconds()
		return result, nil
	}

	// Parse response
	result.Code = response.Code.String()
	result.Attributes = make(map[string]interface{})

	// Extract common attributes
	if username := rfc2865.UserName_GetString(response); username != "" {
		result.Attributes["User-Name"] = username
	}
	if replyMessage := rfc2865.ReplyMessage_GetString(response); replyMessage != "" {
		result.Attributes["Reply-Message"] = replyMessage
	}
	if acctSessionID := rfc2866.AcctSessionID_GetString(response); acctSessionID != "" {
		result.Attributes["Acct-Session-Id"] = acctSessionID
	}

	// Build system output
	result.Systemout = fmt.Sprintf("RADIUS Response: %s", result.Code)
	if len(result.Attributes) > 0 {
		result.Systemout += "\nAttributes:"
		for key, value := range result.Attributes {
			result.Systemout += fmt.Sprintf("\n  %s: %v", key, value)
		}
	}

	// Parse as JSON for systemoutjson
	result.SystemoutJSON = map[string]interface{}{
		"code":       result.Code,
		"attributes": result.Attributes,
	}

	elapsed := time.Since(start)
	result.TimeSeconds = elapsed.Seconds()

	return result, nil
}

// addAttribute adds a RADIUS attribute to the packet
func addAttribute(packet *radius.Packet, name, value string) error {
	switch name {
	case "User-Name":
		rfc2865.UserName_SetString(packet, value)
	case "User-Password":
		rfc2865.UserPassword_SetString(packet, value)
	case "NAS-IP-Address":
		// Parse IP address and set
		rfc2865.NASIPAddress_Set(packet, []byte(value))
	case "NAS-Port":
		if port, err := strconv.Atoi(value); err == nil {
			rfc2865.NASPort_Set(packet, rfc2865.NASPort(port))
		} else {
			return fmt.Errorf("invalid NAS-Port value: %s", value)
		}
	case "Service-Type":
		// Map service type strings to values
		var serviceType uint32
		switch value {
		case "Login-User":
			serviceType = 1
		case "Framed-User":
			serviceType = 2
		case "Callback-Login-User":
			serviceType = 3
		case "Callback-Framed-User":
			serviceType = 4
		case "Outbound-User":
			serviceType = 5
		case "Administrative-User":
			serviceType = 6
		case "NAS-Prompt-User":
			serviceType = 7
		case "Call-Check-User":
			serviceType = 8
		case "Callback-Administrative-User":
			serviceType = 9
		default:
			return fmt.Errorf("unsupported Service-Type: %s", value)
		}
		rfc2865.ServiceType_Set(packet, rfc2865.ServiceType(serviceType))
	case "Framed-Protocol":
		// Map protocol strings to values
		var protocol uint32
		switch value {
		case "PPP":
			protocol = 1
		case "SLIP":
			protocol = 2
		case "ARAP":
			protocol = 3
		case "Gandalf":
			protocol = 4
		case "Xylogics":
			protocol = 5
		case "X.75-Synchronous":
			protocol = 6
		default:
			return fmt.Errorf("unsupported Framed-Protocol: %s", value)
		}
		rfc2865.FramedProtocol_Set(packet, rfc2865.FramedProtocol(protocol))
	case "Framed-IP-Address":
		// Parse IP address and set
		rfc2865.FramedIPAddress_Set(packet, []byte(value))
	case "Framed-IP-Netmask":
		// Parse IP netmask and set
		rfc2865.FramedIPNetmask_Set(packet, []byte(value))
	case "Framed-Routing":
		// Map routing strings to values
		var routing uint32
		switch value {
		case "None":
			routing = 0
		case "Broadcast":
			routing = 1
		case "Listen":
			routing = 2
		case "Broadcast-Listen":
			routing = 3
		default:
			return fmt.Errorf("unsupported Framed-Routing: %s", value)
		}
		rfc2865.FramedRouting_Set(packet, rfc2865.FramedRouting(routing))
	case "Filter-Id":
		rfc2865.FilterID_SetString(packet, value)
	case "Framed-MTU":
		if mtu, err := strconv.Atoi(value); err == nil {
			rfc2865.FramedMTU_Set(packet, rfc2865.FramedMTU(mtu))
		} else {
			return fmt.Errorf("invalid Framed-MTU value: %s", value)
		}
	case "Reply-Message":
		rfc2865.ReplyMessage_SetString(packet, value)
	case "Callback-Number":
		rfc2865.CallbackNumber_SetString(packet, value)
	case "Callback-Id":
		rfc2865.CallbackID_SetString(packet, value)
	case "Framed-Route":
		rfc2865.FramedRoute_SetString(packet, value)
	case "Framed-IPX-Network":
		// Parse IPX network and set
		rfc2865.FramedIPXNetwork_Set(packet, []byte(value))
	case "State":
		rfc2865.State_SetString(packet, value)
	case "Class":
		rfc2865.Class_SetString(packet, value)
	case "Session-Timeout":
		if timeout, err := strconv.Atoi(value); err == nil {
			rfc2865.SessionTimeout_Set(packet, rfc2865.SessionTimeout(timeout))
		} else {
			return fmt.Errorf("invalid Session-Timeout value: %s", value)
		}
	case "Idle-Timeout":
		if timeout, err := strconv.Atoi(value); err == nil {
			rfc2865.IdleTimeout_Set(packet, rfc2865.IdleTimeout(timeout))
		} else {
			return fmt.Errorf("invalid Idle-Timeout value: %s", value)
		}
	case "Termination-Action":
		// Map action strings to values
		var action uint32
		switch value {
		case "Default":
			action = 0
		case "RADIUS-Request":
			action = 1
		default:
			return fmt.Errorf("unsupported Termination-Action: %s", value)
		}
		rfc2865.TerminationAction_Set(packet, rfc2865.TerminationAction(action))
	case "Called-Station-Id":
		rfc2865.CalledStationID_SetString(packet, value)
	case "Calling-Station-Id":
		rfc2865.CallingStationID_SetString(packet, value)
	case "NAS-Identifier":
		rfc2865.NASIdentifier_SetString(packet, value)
	case "Proxy-State":
		rfc2865.ProxyState_SetString(packet, value)
	case "Login-LAT-Service":
		rfc2865.LoginLATService_SetString(packet, value)
	case "Login-LAT-Node":
		rfc2865.LoginLATNode_SetString(packet, value)
	case "Login-LAT-Group":
		rfc2865.LoginLATGroup_SetString(packet, value)
	case "Framed-AppleTalk-Zone":
		rfc2865.FramedAppleTalkZone_SetString(packet, value)
	case "Acct-Status-Type":
		// Map status strings to values
		var status uint32
		switch value {
		case "Start":
			status = 1
		case "Stop":
			status = 2
		case "Interim-Update":
			status = 3
		case "Accounting-On":
			status = 7
		case "Accounting-Off":
			status = 8
		default:
			return fmt.Errorf("unsupported Acct-Status-Type: %s", value)
		}
		rfc2866.AcctStatusType_Set(packet, rfc2866.AcctStatusType(status))
	case "Acct-Session-Id":
		rfc2866.AcctSessionID_SetString(packet, value)
	case "Acct-Authentic":
		// Map authentic strings to values
		var authentic uint32
		switch value {
		case "RADIUS":
			authentic = 1
		case "Local":
			authentic = 2
		case "Remote":
			authentic = 3
		default:
			return fmt.Errorf("unsupported Acct-Authentic: %s", value)
		}
		rfc2866.AcctAuthentic_Set(packet, rfc2866.AcctAuthentic(authentic))
	case "Acct-Session-Time":
		if sessionTime, err := strconv.Atoi(value); err == nil {
			rfc2866.AcctSessionTime_Set(packet, rfc2866.AcctSessionTime(sessionTime))
		} else {
			return fmt.Errorf("invalid Acct-Session-Time value: %s", value)
		}
	case "Acct-Input-Octets":
		if octets, err := strconv.Atoi(value); err == nil {
			rfc2866.AcctInputOctets_Set(packet, rfc2866.AcctInputOctets(octets))
		} else {
			return fmt.Errorf("invalid Acct-Input-Octets value: %s", value)
		}
	case "Acct-Output-Octets":
		if octets, err := strconv.Atoi(value); err == nil {
			rfc2866.AcctOutputOctets_Set(packet, rfc2866.AcctOutputOctets(octets))
		} else {
			return fmt.Errorf("invalid Acct-Output-Octets value: %s", value)
		}
	case "Acct-Input-Packets":
		if packets, err := strconv.Atoi(value); err == nil {
			rfc2866.AcctInputPackets_Set(packet, rfc2866.AcctInputPackets(packets))
		} else {
			return fmt.Errorf("invalid Acct-Input-Packets value: %s", value)
		}
	case "Acct-Output-Packets":
		if packets, err := strconv.Atoi(value); err == nil {
			rfc2866.AcctOutputPackets_Set(packet, rfc2866.AcctOutputPackets(packets))
		} else {
			return fmt.Errorf("invalid Acct-Output-Packets value: %s", value)
		}
	case "Acct-Terminate-Cause":
		// Map cause strings to values
		var cause uint32
		switch value {
		case "User-Request":
			cause = 1
		case "Lost-Carrier":
			cause = 2
		case "Lost-Service":
			cause = 3
		case "Idle-Timeout":
			cause = 4
		case "Session-Timeout":
			cause = 5
		case "Admin-Reset":
			cause = 6
		case "Admin-Reboot":
			cause = 7
		case "Port-Error":
			cause = 8
		case "NAS-Error":
			cause = 9
		case "NAS-Request":
			cause = 10
		case "NAS-Reboot":
			cause = 11
		case "Port-Unneeded":
			cause = 12
		case "Port-Preempted":
			cause = 13
		case "Port-Suspended":
			cause = 14
		case "Service-Unavailable":
			cause = 15
		case "Callback":
			cause = 16
		case "User-Error":
			cause = 17
		case "Host-Request":
			cause = 18
		default:
			return fmt.Errorf("unsupported Acct-Terminate-Cause: %s", value)
		}
		rfc2866.AcctTerminateCause_Set(packet, rfc2866.AcctTerminateCause(cause))
	case "Acct-Multi-Session-Id":
		rfc2866.AcctMultiSessionID_SetString(packet, value)
	case "Acct-Link-Count":
		if count, err := strconv.Atoi(value); err == nil {
			rfc2866.AcctLinkCount_Set(packet, rfc2866.AcctLinkCount(count))
		} else {
			return fmt.Errorf("invalid Acct-Link-Count value: %s", value)
		}
	default:
		return fmt.Errorf("unsupported attribute: %s", name)
	}

	return nil
}
