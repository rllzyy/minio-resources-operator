package v1

import "fmt"

// GetHostname return a minio client compatible hostname
func (ms *MinioServerSpec) GetHostname() string {
	return fmt.Sprintf("%s:%d", ms.Hostname, ms.Port)
}
