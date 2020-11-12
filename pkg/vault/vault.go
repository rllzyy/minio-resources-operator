package vault

import (
	"fmt"
	"os"

	"github.com/hashicorp/vault/api"
	"github.com/minio/minio/pkg/auth"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var vaultClient *api.Client
var log = logf.Log.WithName("vault")

func init() {

	var err error
	vaultClient, err = api.NewClient(nil)

	if err != nil {
		log.Error(err, err.Error())
		os.Exit(2)
	}

	// Check that Vault is responsive (todo: better check?)
	if health, err := vaultClient.Sys().Health(); err == nil {
		log.Info("Vault initialized", "Vault.version", health.Version)
	} else {
		log.Error(err, "Failed communicating with Vault")
		os.Exit(1)
	}

}

func hasCredentials(user string) bool {

	logic := vaultClient.Logical()
	s, err := logic.Read(fmt.Sprintf("minio/data/users/%s", user))

	if err != nil {
		panic(err)
	}

	if s != nil {
		return true
	}

	return false
}

// GetCredentials bleh bleh bleh
func GetCredentials(user string) (auth.Credentials, error) {
	path := fmt.Sprintf("minio/data/users/%s", user)

	if hasCredentials(user) {
		secret, err := vaultClient.Logical().Read(path)

		if err != nil {
			panic(err)
		}

		m, ok := secret.Data["data"].(map[string]interface{})

		if !ok {
			panic("failed to read secret data")
		}

		accessKey, ok := m["accessKey"].(string)

		if !ok {
			panic("no accesskey defined")
		}

		secretKey, ok := m["secretKey"].(string)

		if !ok {
			panic("no secretkey defined")
		}

		creds := auth.Credentials{
			AccessKey: accessKey,
			SecretKey: secretKey,
		}

		return creds, nil

	}

	creds, err := auth.GetNewCredentials()

	if err != nil {
		panic(err)
	}

	data := make(map[string]interface{})
	data["data"] = map[string]string{
		"accessKey": creds.AccessKey,
		"secretKey": creds.SecretKey,
	}

	_, err = vaultClient.Logical().Write(path, data)

	if err != nil {
		panic(err)
	}

	return creds, nil

}
