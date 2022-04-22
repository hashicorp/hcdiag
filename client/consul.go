package client

// https://www.consul.io/api-docs

import (
	"errors"
	"os"
	"path/filepath"
	"strconv"
)

const (
	DefaultConsulAddr = "http://127.0.0.1:8500"

	EnvConsulCaCert        = "CONSUL_CACERT"
	EnvConsulCaPath        = "CONSUL_CAPATH"
	EnvConsulClientCert    = "CONSUL_CLIENT_CERT"
	EnvConsulClientKey     = "CONSUL_CLIENT_KEY"
	EnvConsulHttpSslVerify = "CONSUL_HTTP_SSL_VERIFY"
	EnvConsulTlsServerName = "CONSUL_TLS_SERVER_NAME"
)

// NewConsulAPI returns an APIClient for Consul.
func NewConsulAPI() (*APIClient, error) {
	addr := os.Getenv("CONSUL_HTTP_ADDR")
	if addr == "" {
		addr = DefaultConsulAddr
	}

	headers := map[string]string{}
	token := os.Getenv("CONSUL_TOKEN")
	if token != "" {
		headers["X-Consul-Token"] = token
	}

	t, err := NewConsulTLSConfig()
	if err != nil {
		return nil, err
	}

	apiClient, err := NewAPIClient("consul", addr, headers, *t)
	if err != nil {
		return nil, err
	}

	return apiClient, nil
}

//func NewConsulHTTPClient() (*http.Client, error) {
//	t, err := NewConsulTLSConfig()
//	if err != nil {
//		return nil, err
//	}
//
//	client := http.DefaultClient
//
//	// We don't need to configure TLS if the TLSConfig struct is default.
//	if reflect.DeepEqual(*t, TLSConfig{}) {
//		return client, nil
//	}
//
//	if client.Transport == nil {
//		client.Transport = &http.Transport{
//			TLSClientConfig: &tls.Config{},
//		}
//	}
//	clientTransport := client.Transport.(*http.Transport)
//
//	if clientTransport.TLSClientConfig == nil {
//		clientTransport.TLSClientConfig = &tls.Config{}
//	}
//	clientTLSConfig := clientTransport.TLSClientConfig
//
//	if t.CACert != "" || len(t.CACertBytes) != 0 || t.CAPath != "" {
//		rootConfig := &rootcerts.Config{
//			CAFile:        t.CACert,
//			CACertificate: t.CACertBytes,
//			CAPath:        t.CAPath,
//		}
//		if err := rootcerts.ConfigureTLS(clientTLSConfig, rootConfig); err != nil {
//			return nil, err
//		}
//	}
//
//	clientTLSConfig.InsecureSkipVerify = t.Insecure
//
//	if t.TLSServerName != "" {
//		clientTLSConfig.ServerName = t.TLSServerName
//	}
//
//	var clientCert tls.Certificate
//	foundClientCert := false
//
//	switch {
//	case t.ClientCert != "" && t.ClientKey != "":
//		var err error
//		clientCert, err = tls.LoadX509KeyPair(t.ClientCert, t.ClientKey)
//		if err != nil {
//			return nil, err
//		}
//		foundClientCert = true
//	case t.ClientCert != "" || t.ClientKey != "":
//		return nil, fmt.Errorf("both client cert and client key must be provided")
//	}
//
//	if foundClientCert {
//		// We use `GetClientCertificate` here because it works with Vault, along with other products. In Vault
//		// client authentication can use a different CA than the one used for Vault's certificate. However, it
//		// will indicate that its CA should be used when sending the client certificate request. If the client
//		// certificate does not share this CA, Go will fail with a "remote error: tls: bad certificate" error.
//		// By providing an override for `GetClientCertificate`, we ensure that we send the client certificate
//		// anytime one is requested, even if the CA does not match.
//		//
//		// See GitHub issue https://github.com/hashicorp/vault/issues/2946 for more context on why Vault
//		// uses this mechanism when building their own clients.
//		clientTLSConfig.GetClientCertificate = func(*tls.CertificateRequestInfo) (*tls.Certificate, error) {
//			return &clientCert, nil
//		}
//	}
//
//	return client, nil
//}

// NewConsulTLSConfig returns a *TLSConfig object, using
// default environment variables to build up the object.
func NewConsulTLSConfig() (*TLSConfig, error) {
	shouldVerify := true
	if v := os.Getenv(EnvConsulHttpSslVerify); v != "" {
		// The semantics Consul uses to indicate whether verification should be done
		// is the opposite of the `tlsConfig.Insecure` field; we negate the value indicated by
		// EnvConsulHttpSslVerify environment variable here to align them.
		var err error
		shouldVerify, err = strconv.ParseBool(v)
		if err != nil {
			return nil, err
		}
	}

	return &TLSConfig{
		CACert:        os.Getenv(EnvConsulCaCert),
		CAPath:        os.Getenv(EnvConsulCaPath),
		ClientCert:    os.Getenv(EnvConsulClientCert),
		ClientKey:     os.Getenv(EnvConsulClientKey),
		TLSServerName: os.Getenv(EnvConsulTlsServerName),
		Insecure:      !shouldVerify,
	}, nil
}

func GetConsulLogPath(api *APIClient) (string, error) {
	path, err := api.GetStringValue(
		"/v1/agent/self",
		// format ~ {"DebugConfig": {"Logging": {"LogFilePath": "/some/path"}}}
		"DebugConfig", "Logging", "LogFilePath",
	)

	if err != nil {
		return "", err
	}
	if path == "" {
		return "", errors.New("empty Consul LogFilePath")
	}

	// account for variable behavior depending on destination type
	if _, file := filepath.Split(path); file == "" {
		// this is a directory
		path = path + "consul-*"
	} else {
		// this is a "prefix"
		path = path + "-*"
	}

	return path, nil
}
