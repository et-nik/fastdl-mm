# [Metamod FastDL Plugin](https://github.com/et-nik/fastdl-mm)

FastDL Metamod plugin for GoldSrc (Half-Life 1, CS 1.6) servers. 
This plugin allows you to download files from a web server to the client's computer.

Plugin written with [Metamod-Go](https://github.com/et-nik/metamod-go) library.

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

fastDLHost: "127.0.0.1"      # Leave it empty if you want to use the same IP as the game server
fastDLPort: 13080             # Leave it empty if you want to use random port
autoIndexEnabled: true
forbiddenRegexp:
  - mapcycle.*
  - .*textscheme.*
allowedExtentions:
  - lmp
  - lst
  - wad
  - bmp
  - tga
  - jpg
  - jpeg
  - png
  - gif
  - txt
  - zip
  - bsp
  - res
  - wav
  - mp3
  - spr
allowedPaths:
  - gfx
  - maps
  - media
  - models
  - sound
  - sprites
```

### Configuration options

#### fastDLHost

The host of the FastDL server. This is the IP address. Leave it empty if you want to use the same IP as the game server.

#### fastDLPort

The port of the FastDL server. Leave it empty if you want to use random port.

#### autoIndexEnabled

If enabled, the plugin will generate an index file for each directory. The index file will contain a list of files in the directory.

#### allowedExtentions

A list of allowed file extensions. Files with extensions not in this list will not be downloaded.

#### forbiddenExtentions

A list of forbidden file extensions. Files with extensions in this list will not be downloaded.

#### allowedPaths

A list of allowed paths. Files from directories not in this list will not be downloaded.

#### forbiddenPaths

A list of forbidden paths. Files from directories in this list will not be downloaded.