package config

import (
	"fmt"
	"os"
	"strconv"
)

// MinioServerConfiguration is minio server configuration
type MinioServerConfiguration struct {
	host      string
	port      int
	AccessKey string
	SecretKey string
	SSL       bool
}

// GetHostname return Minio Client compatible hostname
func (m *MinioServerConfiguration) GetHostname() string {
	return fmt.Sprintf("%s:%d", m.host, m.port)
}

// RuntimeServerConfiguration is active configuration
var RuntimeServerConfiguration *MinioServerConfiguration

func fail(msg string) {
	fmt.Println(msg)
	os.Exit(1)
}

func init() {
	// initialize RuntimeServerConfiguration from environment variable
	const (
		prefix                = "MINIO_"
		hostnameKey           = prefix + "HOSTNAME"
		portKey               = prefix + "PORT"
		accessKey             = prefix + "ACCESSKEY"
		secretKey             = prefix + "SECRETKEY"
		sslKey                = prefix + "SSL"
		errInvalidEnvironment = "Invalid environment %s"
		errMissingEnvironment = "Missing environment %s"
	)

	RuntimeServerConfiguration = &MinioServerConfiguration{
		host:      os.Getenv(hostnameKey),
		AccessKey: os.Getenv(accessKey),
		SecretKey: os.Getenv(secretKey),
	}

	var (
		err     error
		portStr = os.Getenv(prefix + "PORT")
		sslStr  = os.Getenv(prefix + "SSL")
	)

	if len(portStr) == 0 {
		RuntimeServerConfiguration.port = 9000
	} else {
		if RuntimeServerConfiguration.port, err = strconv.Atoi(portStr); err != nil {
			fail(fmt.Sprintf(errInvalidEnvironment, portKey))
		}
	}

	if len(sslStr) > 0 {
		if RuntimeServerConfiguration.SSL, err = strconv.ParseBool(sslStr); err != nil {
			fail(fmt.Sprintf(errInvalidEnvironment, sslKey))
		}
	}

	if len(RuntimeServerConfiguration.host) == 0 {
		fail(fmt.Sprintf(errMissingEnvironment, hostnameKey))
	}

	if len(RuntimeServerConfiguration.AccessKey) == 0 {
		fail(fmt.Sprintf(errMissingEnvironment, accessKey))
	}

	if len(RuntimeServerConfiguration.SecretKey) == 0 {
		fail(fmt.Sprintf(errMissingEnvironment, secretKey))
	}
}
