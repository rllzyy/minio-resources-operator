package vault

import (
	"os"

	"github.com/hashicorp/vault/api"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var vaultClient api.Client
var log = logf.Log.WithName("vault")

func init() {

	vaultClient, err := api.NewClient(nil)

	if err != nil {
		log.Error(err, err.Error())
		os.Exit(1)
	}

	if health, err := vaultClient.Sys().Health(); err == nil {
		log.Info("Vault initialized, version: %s", health.Version)
	} else {
		log.Error(err, "Buttbutt")
		os.Exit(1)
	}

}
