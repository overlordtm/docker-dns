# docker-dns

Simple dns server to serve Docker container IP addresses. Had too much coffe and could not sleep. Useful for me, maybe even for you.

# Setup
Generate config:
```
./docker-dns --createConfig
```
If you feel like, you can edit config

# Usage
Add this line to your local dnsmasq resolver
```
server=/yourtld/127.0.0.1#8053
```



