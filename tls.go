package gotalk

import (
  "crypto/x509"
  "fmt"
  "io/ioutil"
)

var tlsCertPool *x509.CertPool

func init() {
  tlsCertPool, _ = x509.SystemCertPool()
  if tlsCertPool == nil {
    tlsCertPool = x509.NewCertPool()
  }
}

// TLSCertPool returns the root CA pool.
// This is normally the same as returned by crypto/x509.SystemCertPool
// and can be modified, i.e. by adding your own development CA certs.
// All gotalk TLS functions that creates tls.Config uses this.
//
func TLSCertPool() *x509.CertPool {
  return tlsCertPool
}

// TLSAddRootCerts is a convenience for adding root (CA) certificates from
// a PEM file to the cert pool used by gotalk's TLS functions and returned by TLSCertPool()
//
func TLSAddRootCerts(certFile string) error {
	buf, err := ioutil.ReadFile(certFile)
	if err != nil {
		return err
	}
  if !tlsCertPool.AppendCertsFromPEM(buf) {
		return fmt.Errorf("failed to load X.509 certificate file %q", certFile)
	}
	return nil
}
