//go:build !cgo
// +build !cgo

package choria

import (
	"fmt"

	"github.com/choria-io/go-choria/inter"
)

func (fw *Framework) setupPKCS11(_ inter.RequestSigner) (err error) {
	return fmt.Errorf("pkcs11 is not supported in this build")
}
