package command

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"

	cf "github.com/vmware/photon-controller-cli/photon/cli/configuration"
)

func isServerTrusted(server string) (bool, error) {
	bServerTrusted := false

	roots, err := cf.GetCertsFromLocalStore()

	if err != nil {
		return bServerTrusted, err
	}

	//Try connecting securely to the server
	config := tls.Config{RootCAs: roots, InsecureSkipVerify: false}
	conn, err := tls.Dial("tcp", server, &config)

	if err == nil {
		bServerTrusted = true
		_ = conn.Close()
	} else {
		switch err.(type) {
		case x509.UnknownAuthorityError:
			bServerTrusted = false
			err = nil
		}
	}

	return bServerTrusted, err
}

func getServerCert(server string) (*x509.Certificate, error) {
	config := tls.Config{InsecureSkipVerify: true}
	conn, err := tls.Dial("tcp", server, &config)

	//Ensure we can connect to the server
	if err == nil {
		cert := new(x509.Certificate)
		state := conn.ConnectionState()
		//return 1st in the cert list (leaf cert)
		if state.PeerCertificates != nil {
			cert = state.PeerCertificates[0]
		}
		_ = conn.Close()
		return cert, nil
	}
	fmt.Println(err)
	return nil, err
}
