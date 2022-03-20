package vault

import (
	"errors"
	"fmt"

	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/Walkbase/minio-resources-operator/auth"
	"github.com/hashicorp/vault/api"
)

var vaultClient *api.Client
var log = ctrl.Log.WithName("vault")
var vaultPath = "minio/"

func ConnectVault(vaultpath string) error {

	// set vaultPath and make sure it ends with a slash
	if vaultpath[len(vaultpath)] == '/' {
		vaultPath = vaultpath
	} else {
		vaultPath = vaultpath + "/"
	}

	log.Info("Connecting to vault")
	vaultClient, err := api.NewClient(nil)
	if err != nil {
		log.Error(err, "Failed to create Vault client")
		return err
	}

	// Check that Vault is responsive (todo: better check?)
	if health, err := vaultClient.Sys().Health(); err == nil {
		log.Info("Vault initialized and healthy", "Vault.version", health.Version)
	} else {
		log.Error(err, "Vault is unhealthy")
		return err
	}

	return nil
}

// GetCredentials fetches minio credentails for a user from Vault or generates
// new ones and saves them to Vault if they do not yet exist.
func GetCredentials(user string) (auth.Credentials, error) {
	path := fmt.Sprintf("%sdata/users/%s", vaultPath, user)
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

// GetServerCredentials fetches minio credentials for a server from Vault
func GetServerCredentials(server string) (auth.Credentials, error) {
	path := fmt.Sprintf("%sdata/servers/%s", vaultPath, server)
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
