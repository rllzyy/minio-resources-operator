package vault

import (
	"os"

	"github.com/hashicorp/vault/api"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var vaultClient api.Client
var log = logf.Log.WithName("vault")

func init() {

	if vaultClient, err := api.NewClient(nil); err != nil {
		log.Error(err, err.Error())
		os.Exit(1)
	}

	log.Info("Vault initialized")

}
