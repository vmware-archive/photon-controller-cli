// Copyright (c) 2016 VMware, Inc. All Rights Reserved.
//
// This product is licensed to you under the Apache License, Version 2.0 (the "License").
// You may not use this product except in compliance with the License.
//
// This product may include a number of subcomponents with separate copyright notices and
// license terms. Your use of these subcomponents is subject to the terms and conditions
// of the subcomponent's license, as noted in the LICENSE file.

package command

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"

	cf "github.com/vmware/photon-controller-cli/photon/configuration"
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
