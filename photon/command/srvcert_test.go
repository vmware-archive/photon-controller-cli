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
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"log"
	"math/big"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	cf "github.com/vmware/photon-controller-cli/photon/configuration"
)

//Generates a self signed test root cert for the test server
func genTestRootCert() (cert_b, priv_b []byte) {
	//Generate a short lived cert
	//valid for 1 day to be used with server hosted
	ca := &x509.Certificate{
		SerialNumber: big.NewInt(1653),
		Subject: pkix.Name{
			Country:            []string{"Test"},
			Organization:       []string{"Test"},
			OrganizationalUnit: []string{"Test"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(0, 0, 1),
		SubjectKeyId:          []byte{1, 2, 3, 4, 5},
		BasicConstraintsValid: true,
		IsCA:        true,
		DNSNames:    []string{"localhost"},
		IPAddresses: []net.IP{net.ParseIP("127.0.0.1")},
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:    x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
	}

	priv, _ := rsa.GenerateKey(rand.Reader, 1024)
	pub := &priv.PublicKey
	ca_b, err := x509.CreateCertificate(rand.Reader, ca, ca, pub, priv)
	if err != nil {
		log.Println("create ca failed", err)
		return
	}
	priv_b = x509.MarshalPKCS1PrivateKey(priv)
	return ca_b, priv_b
}

func TestServerTrustUtils(t *testing.T) {

	//Launch Test Server
	cert_b, priv_b := genTestRootCert()

	//Certificate for the TLS connectiona and private key are the args
	cert, _ := x509.ParseCertificate(cert_b)
	priv, _ := x509.ParsePKCS1PrivateKey(priv_b)

	pool := x509.NewCertPool()
	pool.AddCert(cert)

	tls_cert := tls.Certificate{
		Certificate: [][]byte{cert_b},
		PrivateKey:  priv,
	}

	config := tls.Config{
		ClientAuth:   tls.NoClientCert,
		Certificates: []tls.Certificate{tls_cert},
	}

	//Launch a server with TLS end point
	ts := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, client")
	}))
	defer ts.Close()
	ts.TLS = &config
	ts.StartTLS()

	//At this point the appropriate test root cert doesn't exist
	//Test the case when we dont have root cert for the server
	u, err := url.Parse(ts.URL)
	if err != nil {
		t.Error("Failed to parse URL")
		return
	}

	bServerTrusted, err := isServerTrusted(u.Host)
	if err != nil || bServerTrusted == true {
		fmt.Println(err)
		t.Error("Failed to check server trust")
		return
	}

	//Get the remote server's root cert and add it to our trust list
	cert, err = getServerCert(u.Host)
	if err != nil {
		t.Error("Failed to get server cert")
		return
	}

	err = cf.AddCertToLocalStore(cert)
	if err != nil {
		t.Error("Failed to Add server cert to local store")
		return
	}

	//At this point we should have added the root cert of the remote server
	//trust should be established already
	bServerTrusted, err = isServerTrusted(u.Host)
	if err != nil || bServerTrusted == false {
		t.Error("Failed to check server trust")
	}

	err = cf.RemoveCertFromLocalStore(cert)
	if err != nil {
		t.Error("Failed to Add server cert to local store")
	}
}
