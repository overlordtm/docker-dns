# docker-dns

Simple dns server to serve Docker container IP addresses. Had too much coffe and could not sleep. Useful for me, maybe even for you.

## Setup
Generate config:
```
./docker-dns --createConfig
```
If you feel like, you can edit config

Add this line to your local dnsmasq resolver
```
server=/yourtld/127.0.0.1#8053
```
Then run it
```
./docker-dns --config=/path/to/conf.json
```
Check if it works:
```
dig @127.0.0.1 containerName.tld A
```

## Usage
```
./bin/docker-dns -h
  -config="": Path to config file
  -createConfig=false: Print default config file to stdout
  -docker="unix:///var/run/docker.sock": Address for docker client (HTTP or Unix)
  -listen="127.0.0.1:8053": Listen address
  -loglevel="info": Logrus loglevel
  -tld="docker.": TLD to serve
  -ttl=60: Default TT
```

## systemd service

```
cp server /usr/local/bin/docker-dns
sudo cp docker-dns.service /etc/systemd/system/docker-dns.service
sudo systemctl daemon-reload
sudo systemctl enable docker-dns
sudo systemctl start docker-dns.service
sudo systemctl status docker-dns.service
```

## Use it with NetworkManager
Make sure `dns=dnsmasq` in `/etc/NetworkManager/dnsmasq.d/docker-dns`

```
[main]
plugins=ifupdown,keyfile,dnsmasq
dns=dnsmasq

[ifupdown]
managed=false
```

Create file `/etc/NetworkManager/dnsmasq.d/docker-dns`
```
server=/docker/127.0.0.1#8053
```

```
sudo service network-manager restart
```

