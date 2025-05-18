# scantopl

An updated fork of celedhrim's original work.

Automatically send [scanservjs](https://github.com/sbs20/scanservjs) scanned document to [paperless-ngx](https://github.com/paperless-ngx/paperless-ngx)

## What is updated

3rd party libs
go language
docker base container

## How to configure

```
Usage of /usr/bin/scantopl:
  -config string
        path to config file
  -pltoken string
        Paperless auth token , generated through admin (default "xxxxxxxxxxxxxxxxxx")
  -plurl string
        The paperless instance URL without trailing / (default "http://localhost:8080")
  -scandir string
        Scanserjs ouput directory (default "/home/scanservjs/output")
```

or you can use envvar : SCANDIR, PLTOKEN, PLURL

provide the paperless-ngx url , the paperless-ngx token and the scanservjs output dir ( or bind to /output in docker) 

## How to use it

* Scan something
* if you want to send it to paperless-ngx , go in the scanservjs file section and rename file to add prefix **pl_** ( test_scan.pdf -> pl_test_scan.pdf)
* the file is submitted with name "test_scan" ( remove prefix and extension automatically) then remove source file is deleted 

## How it work

* listen for file creation in the scanservjs output dir
* if a newly created file start with **pl_** , upload it to paperless 
* If uploaded succefully, remove file from scanservjs output

## Install

### Docker

```
$ docker run --rm \
  -v /your/host/scanservjs/output:/output \
  -e PLURL=https://paperless.yourdomain.instance \
  -e PLTOKEN=XXXXXXXXXXXX \
  ghcr.io/starkzarn/scantopl:master
```

### Docker Compose

Example with scanservjs container.

Paperless webserver is reachable on a network called plnet
So scan2pl is connected to this network.

Your local network is repesented via homelan.
Scanservjs is maybe connected here. This is yust an example.


```
version: "3"
services:
  scanservjs:
    container_name: scanservjs
    environment:
      - SANED_NET_HOSTS="<IP of your scanner>"
    volumes:
      - ./scans:/var/lib/scanservjs/output
      - ./config:/etc/scanservjs
      - /var/run/dbus:/var/run/dbus
    ports:
      - 8080:8080
    restart: unless-stopped
    image: sbs20/scanservjs:latest
    networks:
      homelan:
        # Define a static ip for the container. The containter can be accessible by others devices on the LAN network with this IP.
        ipv4_address: <containers ip addr from your local network>
    
  
  scan2pl:
    container_name: scan2pl
    environment:
      - PLURL=http://paperless-ngx-webserver-1:8000
      - PLTOKEN=${PLTOKEN}
    image: ghcr.io/sidey79/scantopl:master
    networks:
      - plnet
    volumes:
      - ./scans:/output

networks:
  homelan:
    name: network_homelan
    external: true

  plnet:
    name: paperless_internal_network
    external: true
     
```
