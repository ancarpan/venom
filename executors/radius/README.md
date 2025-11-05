# Venom - Executor RADIUS

Step to execute RADIUS (Remote Authentication Dial-In User Service) requests

Use case: your software needs to authenticate users or send accounting data to a RADIUS server.

## Input

The following inputs are available:
- `server` (optional): RADIUS server address and port, default: "localhost:1812"
- `secret` (optional): RADIUS shared secret, default: "secret"
- `code` (optional): RADIUS packet type, default: "Access-Request"
- `timeout` (optional): Request timeout in seconds, default: 5
- `attributes`: Map of RADIUS attributes to include in the packet

### Supported RADIUS Packet Types

- `Access-Request`: Authentication request
- `Access-Accept`: Authentication accepted
- `Access-Reject`: Authentication rejected
- `Access-Challenge`: Authentication challenge
- `Accounting-Request`: Accounting data request
- `Accounting-Response`: Accounting data response
- `Status-Server`: Status server request
- `Status-Client`: Status client request
- `Disconnect-Request`: Disconnect request
- `Disconnect-ACK`: Disconnect acknowledgment
- `Disconnect-NAK`: Disconnect negative acknowledgment
- `CoA-Request`: Change of Authorization request
- `CoA-ACK`: Change of Authorization acknowledgment
- `CoA-NAK`: Change of Authorization negative acknowledgment

### Supported RADIUS Attributes

#### Authentication Attributes (RFC 2865)
- `User-Name`: Username for authentication
- `User-Password`: Password for authentication
- `NAS-IP-Address`: Network Access Server IP address
- `NAS-Port`: Network Access Server port
- `Service-Type`: Type of service requested
- `Framed-Protocol`: Protocol to use for framing
- `Framed-IP-Address`: IP address to assign to user
- `Framed-IP-Netmask`: IP netmask for user
- `Framed-Routing`: Routing information
- `Filter-Id`: Filter identifier
- `Framed-MTU`: Maximum Transmission Unit
- `Reply-Message`: Reply message from server
- `Callback-Number`: Callback number
- `Callback-Id`: Callback identifier
- `Framed-Route`: Framed route
- `State`: State attribute
- `Class`: Class attribute
- `Session-Timeout`: Session timeout
- `Idle-Timeout`: Idle timeout
- `Termination-Action`: Termination action
- `Called-Station-Id`: Called station identifier
- `Calling-Station-Id`: Calling station identifier
- `NAS-Identifier`: NAS identifier
- `Proxy-State`: Proxy state
- `Login-LAT-Service`: Login LAT service
- `Login-LAT-Node`: Login LAT node
- `Login-LAT-Group`: Login LAT group
- `Framed-AppleTalk-Zone`: Framed AppleTalk zone

#### Accounting Attributes (RFC 2866)
- `Acct-Status-Type`: Accounting status type (Start, Stop, Interim-Update, etc.)
- `Acct-Session-Id`: Accounting session identifier
- `Acct-Authentic`: Accounting authentication method
- `Acct-Session-Time`: Accounting session time
- `Acct-Input-Octets`: Accounting input octets
- `Acct-Output-Octets`: Accounting output octets
- `Acct-Input-Packets`: Accounting input packets
- `Acct-Output-Packets`: Accounting output packets
- `Acct-Terminate-Cause`: Accounting termination cause
- `Acct-Multi-Session-Id`: Accounting multi-session identifier
- `Acct-Link-Count`: Accounting link count

## Example

```yaml
name: RADIUS testsuite
vars:
  radius_server: "localhost:1812"
  radius_secret: "secret"
  test_user: "testuser"
  test_password: "testpass"

testcases:
- name: radius_access_request
  steps:
  - type: radius
    server: "{{.radius_server}}"
    secret: "{{.radius_secret}}"
    code: "Access-Request"
    timeout: 5
    attributes:
      User-Name: "{{.test_user}}"
      User-Password: "{{.test_password}}"
      NAS-IP-Address: "192.168.1.1"
      NAS-Port: "0"
    assertions:
    - result.err ShouldBeEmpty
    - result.code ShouldNotBeEmpty
    vars:
      radius_response_code: "{{.result.code}}"

- name: radius_accounting_request
  steps:
  - type: radius
    server: "{{.radius_server}}"
    secret: "{{.radius_secret}}"
    code: "Accounting-Request"
    timeout: 5
    attributes:
      User-Name: "{{.test_user}}"
      Acct-Status-Type: "Start"
      Acct-Session-Id: "12345"
      NAS-IP-Address: "192.168.1.1"
    assertions:
    - result.err ShouldBeEmpty
    - result.code ShouldNotBeEmpty

- name: radius_with_service_type
  steps:
  - type: radius
    server: "{{.radius_server}}"
    secret: "{{.radius_secret}}"
    code: "Access-Request"
    timeout: 5
    attributes:
      User-Name: "{{.test_user}}"
      User-Password: "{{.test_password}}"
      Service-Type: "Framed-User"
      Framed-Protocol: "PPP"
      Framed-IP-Address: "192.168.1.100"
      NAS-IP-Address: "192.168.1.1"
    assertions:
    - result.err ShouldBeEmpty
    - result.code ShouldNotBeEmpty
```

## Output

The executor returns a result object that contains the RADIUS response information.

- `result.code`: RADIUS response code (Access-Accept, Access-Reject, etc.)
- `result.attributes`: Map of attributes returned by the RADIUS server
- `result.request`: Information about the request sent
- `result.timeseconds`: Execution duration in seconds
- `result.systemout`: Human-readable output
- `result.systemoutjson`: JSON representation of the response
- `result.err`: Error message if the request failed

### Service-Type Values
- `Login-User`: Login user service
- `Framed-User`: Framed user service
- `Callback-Login-User`: Callback login user service
- `Callback-Framed-User`: Callback framed user service
- `Outbound-User`: Outbound user service
- `Administrative-User`: Administrative user service
- `NAS-Prompt-User`: NAS prompt user service
- `Call-Check-User`: Call check user service
- `Callback-Administrative-User`: Callback administrative user service

### Framed-Protocol Values
- `PPP`: Point-to-Point Protocol
- `SLIP`: Serial Line Internet Protocol
- `ARAP`: AppleTalk Remote Access Protocol
- `Gandalf`: Gandalf protocol
- `Xylogics`: Xylogics protocol
- `X.75-Synchronous`: X.75 Synchronous protocol

### Acct-Status-Type Values
- `Start`: Accounting start
- `Stop`: Accounting stop
- `Interim-Update`: Interim update
- `Accounting-On`: Accounting on
- `Accounting-Off`: Accounting off

## Default Assertion

```yaml
result.err ShouldBeEmpty
```

## Notes

- The RADIUS executor uses UDP for communication with the RADIUS server
- Timeout is applied to the entire request/response cycle
- All attribute values are treated as strings and converted to appropriate RADIUS attribute types
- IP addresses should be provided in dotted decimal notation (e.g., "192.168.1.1")
- Port numbers should be provided as integers
- The executor supports both authentication and accounting RADIUS operations
