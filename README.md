# wsproxy

A pair of proxies (proxy-local and proxy-remote) that allow directing traffic from local applications to Internet services in a controlled way. The protocol runs over WebSocket connections. The `Destination:` header tells the remote proxy the final destination hostname:port for a connection.

This tool requires another system (hostname and local IP manager) to map each desired public Internet host name (e.g. www.google.com) to a local address, e.g. 192.168.2.105. 

Example configuration file:

```
remoteProxyUrl: ws://my-remote-proxy.cloud.com:8088
services:
- name: google
  localAddress: 192.168.2.105:443
  destinationAddress: www.google.com:443
- name: facebook
  localAddress: 192.168.2.106:443
  destinationAddress: www.facebook.com:443
```

