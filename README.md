# dnoxy

A DNS-over-HTTPS client proxy and server with Cloudflare compatible interfaces,
_dnoxy_ (pronounced "d-NOX-y") is a collection of services for running a
DNS-over-HTTPS server, and a local network DNS proxy for those servers.

**Note:** This is proof of concept code, and should not be relied upon for
production use. If you're interested in communicating with existing
DNS-over-HTTPS servers—such as those run by [Cloudflare][cdns] or
[Google][gdns]—you should look at [secure-operator][secop].

[secop]: https://github.com/fardog/secureoperator
[gdns]: https://developers.google.com/speed/public-dns/docs/dns-over-https
[cdns]: https://developers.cloudflare.com/1.1.1.1/dns-over-https/

Right now, _dnoxy_ has two components:

- `dnoxy-http` – an HTTP server which implements Cloudflare's [DNS-over-HTTPS
  DNS Wireformat][dns-wireformat], and looks up answers against plain DNS.
- `dnoxy-dns` – a DNS server which can perform lookups against a DNS-over-HTTPS
  server, such as `dnoxy-http` or Cloudflare DNS.

[dns-wireformat]:
  https://developers.cloudflare.com/1.1.1.1/dns-over-https/wireformat/

A simplified deployment would be:

```
           dns req                | http req |                 dns req
+--------+         +-----------+  |          |  +------------+         +------------+
| client | ------> | dnoxy-dns | -------------> | dnoxy-http | ------> | dns server |
+--------+         +-----------+  |          |  +------------+         +------------+
        Local Network             | Internet |             Remote Network
```

Of course, that's no better than current DNS since it's unencrypted, and no
caching would be performed; but these services are meant to be no more than
building blocks. You would pair `dnoxy-dns` with a caching DNS server like
[dnsmasq][], and `dnoxy-http` with an HTTPS terminator proxy like [nginx][].

[dnsmasq]: http://www.thekelleys.org.uk/dnsmasq/doc.html
[nginx]: https://nginx.org/

## Building

Dockerfiles are includes for the DNS and HTTP components; to build:

```
# dns component
docker build -t dnoxy-dns:latest -f Dockerfile-dns
# http component
docker build -t dnoxy-http:latest -f Dockerfile-http
```

Dependencies are managed with Go 1.11+ modules; to install without Docker:

```
go mod download
go install -v ./...
```

## License

```
   Copyright 2019 Nathan Wittstock

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0
```
