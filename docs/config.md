# Bifrost Configuration

## ServiceSettings

ServiceSettings is the configuration related to the web server.

### Host

*string*

The hostname and port that a Mattermost instance will use as S3 endpoint. The port number is required.

### HealthHost

*string*

The hostname and port that will be used by K8s so that it can ping this endpoint to know the state of the server.

### TLSCertFile

*string*

The path to the certificate file to use for TLS connection security.

### TLSKeyFile

*string*

The path to the TLS key file to use for TLS connection security.

### MaxConnsPerHost

*int*

Maximum allowed number of connections per host.

### ResponseHeaderTimeout

*int*

Specifies the amount of time to wait for a server's response headers after fully writing the request.

## S3Settings

Settings related to S3-compatible object storage instance.

### AccessKeyId

*string*

This is required for access the S3 instance.

### SecretAccessKey

*string*

The secret access key associated with your S3 Access Key ID.

### Bucket

*string*

The name of the bucket for the S3 instance.

### Region

*string*

The AWS region you selected when creating your S3 bucket.

### Endpoint

*string*

Hostname of your S3 instance.

### Scheme

*string*

Protocol scheme with the S3 instance.

## LogSettings

### EnableConsole

*bool*

If true, the server outputs log messages to the console based on ConsoleLevel option.

### ConsoleLevel

*string*

Level of detail at which log events are written to the console.

### ConsoleJson

*bool*

When true, logged events are written in a machine readable JSON format. Otherwise they are printed as plain text.

### EnableFile

*bool*

When true, logged events are written to the file specified by the `FileLocation` setting.

### FileLevel

*string*

Level of detail at which log events are written to log files.

### FileJson

*bool*

When true, logged events are written in a machine readable JSON format. Otherwise they are printed as plain text.

### FileLocation

*string*

The location of the log file.
