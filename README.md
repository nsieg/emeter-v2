# Home Energy Meter

## Setup RaspberryPi
1. Install RaspberryPi-OS 64bit Lite
1. Create local directories
    ```bash
    mkdir -p /opt/energymeter/backup
    ```
1. Setup InfluxDB
    1. Install InfluxDB according to website docs
    1. Create Influx API user
        1. Copy token into emeter.env
        1. Copy token into vzlogger.conf
1. Install vzlogger dependencies
    ```bash
    apt-get update
    apt-get install -y libcurl4 libgnutls30 libsasl2-2  libuuid1 libssl1.1 libgcrypt20  libmicrohttpd12 libltdl7 libatomic1 libjson-c3 liblept5 libmosquitto1 libunistring2 
    ```

## Build
```bash
cd systemd
./build.sh
```

## Deploy
```bash
cd systemd
./deploy.sh
```

## Manually build vzlogger
```bash
cd /tmp
git clone https://github.com/volkszaehler/vzlogger.git
cd vzlogger
docker build -t vzlogger .
```
