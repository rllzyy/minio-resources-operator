package vault

import (
	"fmt"
	"os"

	"github.com/hashicorp/vault/api"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var vaultClient api.Client
var log = logf.Log.WithName("vault")

func VaultInit() {

	vaultClient, err := api.NewClient(nil)

	if err != nil {
		log.Error(err, err.Error())
		os.Exit(2)
	}

	if health, err := vaultClient.Sys().Health(); err == nil {
		log.Info("Vault initialized, version: %s", health.Version)
	} else {
		log.Error(err, "Buttbutt")
		os.Exit(3)
	}

}

// Hello just prints a message and forces and init
func Hello() {
	fmt.Println("Hello")
}
