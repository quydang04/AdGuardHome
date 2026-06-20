&nbsp;
<p align="center">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="doc/adguard_home_darkmode.svg">
    <img alt="AdGuard Home" src="doc/adguard_home_lightmode.svg" width="300px">
  </picture>
</p>
<h3 align="center">Privacy protection center for you and your devices</h3>
<p align="center">
  Free and open source, powerful network-wide ads & trackers blocking DNS server.
</p>
<p align="center">
  <a href="https://github.com/quydang04/AdGuardHome/releases">
    <img src="https://img.shields.io/github/release/quydang04/AdGuardHome/all.svg" alt="Latest release"/>
  </a>
  <a href="https://github.com/quydang04/AdGuardHome/actions/workflows/build.yml">
    <img src="https://github.com/quydang04/AdGuardHome/actions/workflows/build.yml/badge.svg" alt="Build Status"/>
  </a>
  <a href="https://github.com/quydang04/AdGuardHome/actions/workflows/docker-simple.yml">
    <img src="https://github.com/quydang04/AdGuardHome/actions/workflows/docker-simple.yml/badge.svg" alt="Docker Build Status"/>
  </a>
  <a href="https://hub.docker.com/r/quydang04/adguardhome">
    <img alt="Docker Pulls" src="https://img.shields.io/docker/pulls/quydang04/adguardhome.svg?maxAge=604800"/>
  </a>
</p>
<br/>
<p align="center">
  <img src="https://cdn.adtidy.org/public/Adguard/Common/adguard_home.gif" width="800"/>
</p>
<hr/>

AdGuard Home is a network-wide software for blocking ads and tracking. After you set it up, it'll cover ALL your home devices, and you don't need any client-side software for that.

It operates as a DNS server that re-routes tracking domains to a "black hole", thus preventing your devices from connecting to those servers.

This is a fork maintained by [quydang04](https://github.com/quydang04), based on the original [AdGuardHome](https://github.com/AdguardTeam/AdGuardHome) by AdGuard Team.

- [Getting Started](#getting-started)
    - [Automated install (Linux/Unix/MacOS/FreeBSD/OpenBSD)](#automated-install-linux-and-mac)
    - [Docker](#docker)
    - [Manual installation](#manual-installation)
- [How to build from source](#how-to-build)
- [Contributing](#contributing)
- [Acknowledgments](#acknowledgments)

## <a href="#getting-started" id="getting-started" name="getting-started">Getting Started</a>

### <a href="#automated-install-linux-and-mac" id="automated-install-linux-and-mac" name="automated-install-linux-and-mac">Automated install (Linux/Unix/MacOS/FreeBSD/OpenBSD)</a>

To install with `curl` run the following command:

```sh
curl -s -S -L https://raw.githubusercontent.com/quydang04/AdGuardHome/master/scripts/install.sh | sh -s -- -v
```

To install with `wget` run the following command:

```sh
wget --no-verbose -O - https://raw.githubusercontent.com/quydang04/AdGuardHome/master/scripts/install.sh | sh -s -- -v
```

To install with `fetch` run the following command:

```sh
fetch -o - https://raw.githubusercontent.com/quydang04/AdGuardHome/master/scripts/install.sh | sh -s -- -v
```

The script also accepts some options:

- `-r` to reinstall AdGuard Home;
- `-u` to uninstall AdGuard Home;
- `-v` for verbose output.

Note that options `-r` and `-u` are mutually exclusive.

### <a href="#docker" id="docker" name="docker">Docker</a>

Docker images are automatically built and pushed to both **Docker Hub** and **GitHub Container Registry** on every push to `master`.

**Pull from Docker Hub:**

```sh
docker pull quydang04/adguardhome:latest
```

**Pull from GitHub Container Registry:**

```sh
docker pull ghcr.io/quydang04/adguardhome:latest
```

**Run:**

```sh
docker run -d \
  --name adguardhome \
  --restart unless-stopped \
  -v /opt/adguardhome/work:/opt/adguardhome/work \
  -v /opt/adguardhome/conf:/opt/adguardhome/conf \
  -p 53:53/tcp -p 53:53/udp \
  -p 3000:3000/tcp \
  quydang04/adguardhome:latest
```

Then open the setup wizard at `http://<your-server-ip>:3000`.

After the initial setup, the web interface will be available on the port you configured (default: 80).

**Docker Compose:**

```yaml
services:
  adguardhome:
    image: quydang04/adguardhome:latest
    container_name: adguardhome
    restart: unless-stopped
    ports:
      - "53:53/tcp"
      - "53:53/udp"
      - "3000:3000/tcp"
    volumes:
      - ./work:/opt/adguardhome/work
      - ./conf:/opt/adguardhome/conf
```

### <a href="#manual-installation" id="manual-installation" name="manual-installation">Manual installation</a>

Download the latest release from the [Releases page](https://github.com/quydang04/AdGuardHome/releases), extract it, and run:

```sh
./AdGuardHome -s install
```

Available platforms:

| Platform | Architecture | File |
|----------|--------------|------|
| Windows | x64 | `AdGuardHome_windows_amd64.zip` |
| Windows | ARM64 | `AdGuardHome_windows_arm64.zip` |
| Linux | x64 | `AdGuardHome_linux_amd64.tar.gz` |
| Linux | ARM64 | `AdGuardHome_linux_arm64.tar.gz` |
| Linux | ARMv7 | `AdGuardHome_linux_armv7.tar.gz` |
| macOS | x64 | `AdGuardHome_darwin_amd64.zip` |
| macOS | ARM64 (M1/M2) | `AdGuardHome_darwin_arm64.zip` |
| FreeBSD | x64 | `AdGuardHome_freebsd_amd64.tar.gz` |

## <a href="#how-to-build" id="how-to-build" name="how-to-build">How to build from source</a>

### Prerequisites

You will need:

- [Go](https://golang.org/dl/) v1.25 or later;
- [Node.js](https://nodejs.org/en/download/) v24.10.0 or later;
- [npm](https://www.npmjs.com/) v10.8 or later;

Run `make init` to prepare the development environment.

### Building

```sh
git clone https://github.com/quydang04/AdGuardHome
cd AdGuardHome
make
```

#### Building for a different platform

```sh
env GOOS='linux' GOARCH='arm64' make
```

#### Preparing releases

```sh
make build-release SIGN=0 VERSION='1.0.0'
```

#### Docker image

```sh
make build-docker
```

## <a href="#contributing" id="contributing" name="contributing">Contributing</a>

You are welcome to fork this repository, make your changes and [submit a pull request](https://github.com/quydang04/AdGuardHome/pulls).

### Report issues

If you run into any problem or have a suggestion, head to the [Issues page](https://github.com/quydang04/AdGuardHome/issues) and click on the "New issue" button.

## <a href="#acknowledgments" id="acknowledgments" name="acknowledgments">Acknowledgments</a>

This project is based on the original [AdGuard Home](https://github.com/AdguardTeam/AdGuardHome) by AdGuard Team.

This software wouldn't have been possible without:

- [Go](https://golang.org/dl/) and its libraries:
    - [gcache](https://github.com/bluele/gcache)
    - [miekg's dns](https://github.com/miekg/dns)
    - [go-yaml](https://github.com/go-yaml/yaml)
    - [service](https://godoc.org/github.com/kardianos/service)
    - [dnsproxy](https://github.com/AdguardTeam/dnsproxy)
    - [urlfilter](https://github.com/AdguardTeam/urlfilter)
- [Node.js](https://nodejs.org/) and its libraries:
    - [React.js](https://reactjs.org)
    - [Tabler](https://github.com/tabler/tabler)
    - And many more Node.js packages.
- [whotracks.me data](https://github.com/cliqz-oss/whotracks.me)
