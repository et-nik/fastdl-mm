![](fastdl-metamod.png)

[FastDL Metamod plugin](https://github.com/et-nik/fastdl-mm) for GoldSrc ([Half-Life 1](https://github.com/ValveSoftware/halflife), CS 1.6) servers. 
This plugin allows you to download files from a web server to the client's computer.

This plugin is safe and does not allow downloading files from forbidden directories or files with forbidden extensions (cfg, ini, ...).

Plugin written with [Metamod-Go](https://github.com/et-nik/metamod-go) library.

## Features

- File caching to reduce the load on the web server.
- Serve only precached files. If enabled, you can't download files that are not in the precache list.
- Secure file downloading. The plugin does not allow downloading files from forbidden directories or files with forbidden extensions.
- Rate limiting. Plugin can block IP addresses that download files too often.
- IP Blocklist. You can block IP addresses or IP subnet.

## Installation

1. Download the latest release from the [releases page](https://github.com/et-nik/fastdl-mm/releases)
2. Copy the `fastdl_386.so` file to the `addons/fastdl` directory of your game server
3. Open `addons/metamod/plugins.ini` file and add the following line:
```
linux addons/fastdl/fastdl_386.so
```

4. Restart the game server

### Optional

5. Create a `fastdl.yaml` file in the `addons/fastdl` directory and configure the plugin (see [Configuration](#configuration))

## Configuration

You can configure the plugin using the `fastdl.yaml` file. The file can be located in game directory or in the `addons/fastdl` directory.

### Example

```yaml
# fastdl.yaml

# The host of the FastDL HTTP server. 
# Leave it empty if you want to use the same IP as the game server.
# host: "127.0.0.1"

# The port of the FastDL HTTP server. 
# Leave it empty if you want to use random port.
# port: 13080

# The range of random ports for the FastDL HTTP server.
# If the port is not specified, the plugin will use a random port from this range.
# If the port is specified, the plugin will use the specified port, ignoring this range.
#portRange: 40000-50000

# Serve only precached files. 
# If enabled, the plugin will not allow downloading files that are not in the precache list.
servePrecached: false

# Generate auto index page for directories. 
# It allows to see the list of files in the directory.
autoIndexEnabled: true

# Cache size for downloaded files. 
# The plugin will delete the oldest files if the cache is full.
cacheSize: 50MB

# Forbidden files and directories by regular expressions.
forbiddenRegexp:
  - mapcycle.*
  - .*textscheme.*
    
# Allowed file extensions. 
# Files with extensions not in this list can not be downloaded.
allowedExtensions:
  - bmp
  - bsp
  - gif
  - jpeg
  - jpg
  - lmp
  - lst
  - mdl
  - mp3
  - png
  - res
  - spr
  - tga
  - txt
  - wad
  - wav
  - zip

# Allowed paths. 
# Files from directories not in this list can not be downloaded.
allowedPaths:
  - gfx
  - maps
  - media
  - models
  - overviews
  - sound
  - sprites

# Block IP addresses by single IP, IP subnet or IP range.
# It supports IPv4 and IPv6.
#blockListIP:
#  - 192.0.2.1
#  - 10.80.0.0/24
#  - 172.12.132.1-172.12.132.20

# Rate limiting for IP addresses.
rateLimits:

  - limit: 5
    period: 1s
    
  - limit: 100
    period: 1m
```

### Configuration options

#### host

The host of the FastDL server. This is the IP address. 
Leave it empty if you want to use the same IP as the game server.

#### port

The port of the FastDL server. Leave it empty if you want to use random port.

#### portRange

The range of random ports for the FastDL server. 
If the port is not specified, the plugin will use a random port from this range. 
If the port is specified, the plugin will use the specified port, ignoring this range.

#### servePrecached

Serve only precached files. If enabled, the plugin will not allow downloading 
files that are not in the precache list.

#### autoIndexEnabled

If enabled, the plugin will generate an index file for each directory. 
The index file will contain a list of files in the directory.

#### cacheSize

The size of the cache for downloaded files. 
The plugin will delete the oldest files if the cache is full. 
The size can be specified in bytes, kilobytes, megabytes, or gigabytes.
Example values: `50MB`, `1GB`.

#### allowedExtensions

A list of allowed file extensions. 
Files with extensions not in this list will not be downloaded.

#### forbiddenExtensions

A list of forbidden file extensions. 
Files with extensions in this list will not be downloaded.

#### allowedPaths

A list of allowed paths. 
Files from directories not in this list will not be downloaded.

#### forbiddenPaths

A list of forbidden paths. 
Files from directories in this list will not be downloaded.

#### customDownloadURL

A custom download URL. 
If specified, the plugin will use this URL to download files.
Example: `http://example.com:14080/`

#### blockListIP

Block IP addresses by single IP, IP subnet, or IP range. 
It supports IPv4 and IPv6.

Example values IPv4:
- `192.0.2.1`
- `10.80.0.0/24`
- `172.12.132.1-172.12.132.20`


Example values IPv6:
- `2001:db8::1`
- `2001:db8::/32`
- `2001:db8::1-2001:db8::20`

#### rateLimits

Rate limiting for IP addresses. 
You can specify multiple rate limits.
The plugin can block IP addresses that download files too often.

Example 1. Limit 5 requests per second per IP (5 rps per IP):
```yaml
rateLimits:
  - limit: 5
    period: 1s
```

Example 2. Limit 100 requests per minute per IP (100 rpm per IP):
```yaml
rateLimits:
  - limit: 100
    period: 1m
```

Example 3. Two rate limits. 100 rpm and 5 rps per IP:
```yaml
rateLimits:
  - limit: 5
    period: 1s
    
  - limit: 100
    period: 1m
```