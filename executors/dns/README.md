# Venom - Executor DNS

Step to execute DNS (Domain Name System) queries

Use case: your software needs to perform DNS lookups and verify DNS responses.

## Input

The following inputs are available:
- `server` (mandatory): DNS server address and port (e.g., "8.8.8.8:53")
- `query` (mandatory): Domain name to query
- `qtype` (optional): DNS record type, default: "A"
- `timeout` (optional): Query timeout in seconds, default: 5

### Supported DNS Record Types

- `A`: IPv4 address record
- `AAAA`: IPv6 address record
- `MX`: Mail exchange record
- `TXT`: Text record
- `CNAME`: Canonical name record
- `NS`: Name server record
- `PTR`: Pointer record
- `SOA`: Start of Authority record
- `SRV`: Service record
- `CAA`: Certificate Authority Authorization record
- `ANY`: Any record type

## Example

```yaml
name: DNS testsuite
vars:
  dns_server: "8.8.8.8:53"
  test_domain: "example.com"

testcases:
- name: dns_a_query
  steps:
  - type: dns
    server: "{{.dns_server}}"
    query: "{{.test_domain}}"
    qtype: A
    timeout: 5
    assertions:
    - result.err ShouldBeEmpty
    - result.rcode ShouldEqual "NOERROR"
    - result.messagejson.answer ShouldNotBeEmpty
    vars:
      dns_response_code: "{{.result.rcode}}"

- name: dns_mx_query
  steps:
  - type: dns
    server: "{{.dns_server}}"
    query: "{{.test_domain}}"
    qtype: MX
    timeout: 5
    assertions:
    - result.err ShouldBeEmpty
    - result.rcode ShouldEqual "NOERROR"
    - result.messagejson.answer ShouldNotBeEmpty

- name: dns_txt_query
  steps:
  - type: dns
    server: "{{.dns_server}}"
    query: "{{.test_domain}}"
    qtype: TXT
    timeout: 5
    assertions:
    - result.err ShouldBeEmpty
    - result.rcode ShouldEqual "NOERROR"

- name: dns_aaaa_query
  steps:
  - type: dns
    server: "{{.dns_server}}"
    query: "{{.test_domain}}"
    qtype: AAAA
    timeout: 5
    assertions:
    - result.err ShouldBeEmpty
    - result.rcode ShouldEqual "NOERROR"

- name: dns_ns_query
  steps:
  - type: dns
    server: "{{.dns_server}}"
    query: "{{.test_domain}}"
    qtype: NS
    timeout: 5
    assertions:
    - result.err ShouldBeEmpty
    - result.rcode ShouldEqual "NOERROR"
    - result.messagejson.authority ShouldNotBeEmpty
```

## Output

The executor returns a result object that contains the DNS response information.

- `result.query`: The domain name that was queried
- `result.qtype`: The DNS record type that was queried
- `result.server`: The DNS server that was used
- `result.rcode`: DNS response code (NOERROR, NXDOMAIN, SERVFAIL, etc.)
- `result.message`: Full DNS message in JSON structure
- `result.messagejson`: JSON representation of the DNS message (accessible for assertions)
- `result.timeseconds`: Query execution duration in seconds
- `result.systemout`: Human-readable output
- `result.systemoutjson`: JSON representation of the response (same as message)
- `result.err`: Error message if the query failed

### DNS Message Structure

The `result.message` (and `result.messagejson`) contains the full DNS message with the following structure:

```json
{
  "id": 12345,
  "response": true,
  "opcode": "QUERY",
  "authoritative": false,
  "truncated": false,
  "recursion_desired": true,
  "recursion_available": true,
  "zero": false,
  "authenticated_data": false,
  "checking_disabled": false,
  "rcode": "NOERROR",
  "question": [
    {
      "name": "example.com.",
      "type": "A",
      "class": "IN"
    }
  ],
  "answer": [
    {
      "name": "example.com.",
      "type": "A",
      "class": "IN",
      "ttl": 3600,
      "value": "example.com.\t3600\tIN\tA\t93.184.216.34",
      "address": "93.184.216.34"
    }
  ],
  "authority": [...],
  "additional": [...]
}
```

### Record Type-Specific Fields

Different record types expose specific fields in the JSON structure:

- **A records**: `address` - IPv4 address
- **AAAA records**: `address` - IPv6 address
- **MX records**: `preference`, `mx` - Mail exchange preference and hostname
- **TXT records**: `txt` - Text data array
- **CNAME records**: `target` - Canonical name target
- **NS records**: `ns` - Name server hostname
- **PTR records**: `ptr` - Pointer target
- **SOA records**: `ns`, `mbox`, `serial`, `refresh`, `retry`, `expire`, `minttl`
- **SRV records**: `priority`, `weight`, `port`, `target`
- **CAA records**: `flag`, `tag`, `value`

## Default Assertion

```yaml
result.err ShouldBeEmpty
```

## Notes

- The DNS executor uses UDP for communication with the DNS server by default
- If a UDP response is truncated (too large for UDP), the executor automatically retries with TCP to get the full response
- The `server` parameter is mandatory - the test will fail if not provided
- Timeout is applied to the entire query/response cycle (including TCP retry if needed)
- The executor returns the full DNS message structure, allowing comprehensive assertions on DNS responses
- All DNS record types are parsed and exposed with type-specific fields for easy assertion
- Response codes include: NOERROR, NXDOMAIN, SERVFAIL, NOTIMP, REFUSED, and others as defined in RFC 1035
