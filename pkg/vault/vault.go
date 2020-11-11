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

	// Check that Vault is responsive (todo: better check?)
	if health, err := vaultClient.Sys().Health(); err == nil {
		log.Info("Vault initialized", "Vault.version", health.Version)
	} else {
		log.Error(err, "Failed communicating with Vault")
		os.Exit(1)
	}

}

// Hello just prints a message and forces and init
func Hello() {
	fmt.Println("Hello")
}
