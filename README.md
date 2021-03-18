# DNS resolver

This is a simple DNS resolver.  Sends DNS queries to an upstream DNS server
and displays the result.  This is mostly just a protocol engine; there is
minimal interpretation or post-processing on the responses, etc.


## Known issues
- Assumes all queries + replies are CLASS IN (Internet).
- Decoding of some Resource Records is incomplete.
- No support for EDNS (RFC 6891).


## Build
```
dan@dan-desktop:~/src/ddnsr$ make
dan@dan-desktop:~/src/ddnsr$ make test
=== RUN   TestMessageHeaderPacking
--- PASS: TestMessageHeaderPacking (0.00s)
=== RUN   TestMessagePacking
--- PASS: TestMessagePacking (0.00s)
=== RUN   TestNameCompression
--- PASS: TestNameCompression (0.00s)
=== RUN   TestNamePacking
=== RUN   TestNamePacking/a
=== RUN   TestNamePacking/amazon
--- PASS: TestNamePacking (0.00s)
    --- PASS: TestNamePacking/a (0.00s)
    --- PASS: TestNamePacking/amazon (0.00s)
=== RUN   TestResourceRecordPacking
--- PASS: TestResourceRecordPacking (0.00s)
=== RUN   TestQuestionPacking
--- PASS: TestQuestionPacking (0.00s)
PASS
coverage: 50.0% of statements
ok      ddnsr   0.004s
dan@dan-desktop:~/src/ddnsr$ make vet
```

## Usage
```
Usage: ./ddnsr [options] hostname1 hostname2 ...
  -raw
        Show the raw packet bytes?
  -recursive
        Send a recursive DNS query? (default true)
  -rtype string
        DNS record type (A, ALL, CNAME, MX, PTR, SOA, TXT, etc) (default "A")
  -server string
        IP address of upstream DNS server (default "1.1.1.1")
  -timeout uint
        Request timeout, in seconds (default 3)
```

## Examples
```
dan@dan-desktop:~/src/ddnsr$ ./ddnsr amazon.com
H:  flags 0x8180 (QR RD RA), QD 1, AN 3, NS 0, AR 0
Q:  amazon.com (A)
A:  amazon.com (A), TTL 23: 205.251.103.103
A:  amazon.com (A), TTL 23: 176.32.205.205
A:  amazon.com (A), TTL 23: 54.239.85.85

dan@dan-desktop:~/src/ddnsr$ ./ddnsr -rtype MX google.com
H:  flags 0x8180 (QR RD RA), QD 1, AN 5, NS 0, AR 0
Q:  google.com (MX)
A:  google.com (MX), TTL 600: alt2.aspmx.l.google.com
A:  google.com (MX), TTL 600: aspmx.l.google.com
A:  google.com (MX), TTL 600: alt4.aspmx.l.google.com
A:  google.com (MX), TTL 600: alt3.aspmx.l.google.com
A:  google.com (MX), TTL 600: alt1.aspmx.l.google.com
```
