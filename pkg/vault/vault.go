package vault

import (
	"errors"
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

// GetCredentials bleh bleh bleh
func GetCredentials(user string) (auth.Credentials, error) {
	path := fmt.Sprintf("minio/data/users/%s", user)
	secret, err := vaultClient.Logical().Read(path)

	if err != nil {
		return auth.Credentials{}, err
	}

	if secret != nil {

		m, ok := secret.Data["data"].(map[string]interface{})

		if !ok {
			return auth.Credentials{}, errors.New("failed to read secret data")
		}

		accessKey, ok := m["accessKey"].(string)

		if !ok {
			return auth.Credentials{}, errors.New("no accesskey defined")
		}

		secretKey, ok := m["secretKey"].(string)

		if !ok {
			return auth.Credentials{}, errors.New("no secretkey defined")
		}

		creds := auth.Credentials{
			AccessKey: accessKey,
			SecretKey: secretKey,
		}

		return creds, nil

	}

	creds, err := auth.GetNewCredentials()

	if err != nil {
		return auth.Credentials{}, err
	}

	data := make(map[string]interface{})
	data["data"] = map[string]string{
		"accessKey": creds.AccessKey,
		"secretKey": creds.SecretKey,
	}

	_, err = vaultClient.Logical().Write(path, data)

	if err != nil {
		return auth.Credentials{}, err
	}

	return creds, nil

}

// GetServerCredentials bleh bleh bleh
func GetServerCredentials(server string) (auth.Credentials, error) {
	path := fmt.Sprintf("minio/data/servers/%s", server)
	secret, err := vaultClient.Logical().Read(path)

	if err != nil {
		return auth.Credentials{}, err
	}

	if secret != nil {

		m, ok := secret.Data["data"].(map[string]interface{})

		if !ok {
			return auth.Credentials{}, errors.New("failed to read secret data")
		}

		accessKey, ok := m["accessKey"].(string)

		if !ok {
			return auth.Credentials{}, errors.New("no accesskey defined")
		}

		secretKey, ok := m["secretKey"].(string)

		if !ok {
			return auth.Credentials{}, errors.New("no secretkey defined")
		}

		creds := auth.Credentials{
			AccessKey: accessKey,
			SecretKey: secretKey,
		}

		return creds, nil

	}

	return auth.Credentials{}, errors.New("no credentials for server")
}
