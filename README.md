# DumbDNS (With AdBlock)

DumbDNS is a stupid simple DNS proxy with Ad Blocking written in [Go](https://go.dev/). It compiles to a single Go binary and is exceptionally easy to run. It's not designed to be feature rich or complete.

Life's too short to be setting up PiHole and maintaining it. You can start using DumbDNS with a few easy commands.

DumbDNS currently comes with the following features:

- Ad blocking
- Cached lookups (15 min TTL)
- Block list refreshing
- White list (bypass any blocked domain)
- Rejects external IPs
- Misses out 90% of the DNS spec (:

### Use cases

I've been running a WireGuard server with DumbDNS on both my laptop and phone for a few months now - and it works great.

### Getting started (Ubuntu)

Stop the system DNS service and free up port 53

```bash
service systemd-resolved stop
```

Set the system default DNS to 1.1.1.1 (CloudFlare) or 8.8.8.8 (Google) so we can download the blocklists.

```bash
nano /etc/resolv.conf
nameserver 8.8.8.8
```

Start the service in the background
```bash
./dumbdns &
```

**Note**: External non-private IPs are rejected and the service will bind to port 53.

### Project Roadmap

- Config file
- A simple way to add domains to the whitelist
- IPv6 support
- DNS over HTTPS (one day)

### Who built this & licenses.

This DNS Proxy is created by [Patrick Socha](https://psocha.co.uk) and is licensed under the [MIT License](LICENSE).

It makes use of the [miekg/dns](https://github.com/miekg/dns) package, which is licensed under [BSD 3-Clause "New" or "Revised" License](https://github.com/miekg/dns/blob/master/LICENSE).
