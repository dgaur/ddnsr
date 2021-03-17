# DNS resolver

This is a simple DNS resolver.  Sends DNS queries to an upstream DNS server
and displays the result.  This is mostly just a protocol engine; there is
minimal interpretation or post-processing on the responses, etc.

```
Usage: ./ddnsr [options] hostname1 hostname2 ...
  -recursive
    	Recursive DNS query? (default true)
  -server string
    	IP address of upstream DNS server (default "1.1.1.1")
  -timeout uint
    	Request timeout, in seconds (default 3)
```
