# SES Exporter
Prometheus exporter for Amazon Simple Email Service (SES) metrics, written in Go.

## Installation and Usage
The `ses_exporter` listens on HTTP port 9101 by default. See the `--help` output for more options.

The `ses_exporter` uses the AWS SDK for Go which requires credentials (an access key and secret access key) to sign requests to AWS. You can specify your credentials in several different locations, depending on your particular use case. For information about obtaining credentials, see [Setting Up](https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/setting-up.html).

## Development building and running
Prerequisites:

* [Go compiler](https://golang.org/dl/)

Building:

```bash
git clone https://github.com/warpnet/ses_exporter.git
cd ses_exporter
go build
./ses_exporter <flags>
```

To see all available configuration flags:

```bash
./ses_exporter --help
```

## Docker Compose Usage
Below is an example `docker-compose.yml` file that starts `ses_exporter` with port `9101` exposed. It also mounts your local `~/.aws` directory into the container so it can access your AWS credentials:
