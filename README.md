# DNS resolver

This is a simple DNS resolver.  Sends DNS queries to an upstream resolver
and displays the result.  Minimal interpretation or post-processing on the
response.

```
Usage: ./ddnsr [options] hostname1 hostname2 ...
  -server string
    	Upstream DNS server (default "1.1.1.1")
  -timeout uint
    	Request timeout, in seconds (default 3)
```
