// Copyright 2020 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package secrets

import (
	"github.com/gardener/gardener/pkg/utils"
)

// DataKeyVPNTLSAuth is the key in a secret data holding the vpn tlsauth key.
const DataKeyVPNTLSAuth = "vpn.tlsauth"

// VPNTLSAuthConfig contains the specification for a to-be-generated vpn tls authentication secret.
// The key will be generated by the provided VPNTLSAuthKeyGenerator. By default the openvpn command is used to generate the key if no generator function is specified.
type VPNTLSAuthConfig struct {
	Name                   string
	VPNTLSAuthKeyGenerator func() ([]byte, error)
}

// VPNTLSAuth contains the name and the generated vpn tls authentication key.
type VPNTLSAuth struct {
	Name       string
	TLSAuthKey []byte
}

// GetName returns the name of the secret.
func (s *VPNTLSAuthConfig) GetName() string {
	return s.Name
}

// Generate implements ConfigInterface.
func (s *VPNTLSAuthConfig) Generate() (DataInterface, error) {
	key, err := s.generateKey()
	if err != nil {
		return nil, err
	}

	return &VPNTLSAuth{
		Name:       s.Name,
		TLSAuthKey: key,
	}, nil
}

func (s *VPNTLSAuthConfig) generateKey() (key []byte, err error) {
	if s.VPNTLSAuthKeyGenerator != nil {
		key, err = s.VPNTLSAuthKeyGenerator()
	} else {
		key, err = GenerateVPNKey()
	}
	return
}

// SecretData computes the data map which can be used in a Kubernetes secret.
func (v *VPNTLSAuth) SecretData() map[string][]byte {
	data := map[string][]byte{
		DataKeyVPNTLSAuth: v.TLSAuthKey,
	}
	return data
}

// generateVPNKey generates PSK for OpenVPN similar as generated by `openvpn --genkey` command.
func generateVPNKey() ([]byte, error) {
	allowedCharacters := "0123456789abcdef"
	keyString, err := utils.GenerateRandomStringFromCharset(512, allowedCharacters)
	if err != nil {
		return nil, err
	}

	var formattedKeyString string

	for i := 0; i < 16; i++ {
		formattedKeyString = formattedKeyString + keyString[i*32:((i+1)*32)] + "\n"
	}

	startString := "-----BEGIN OpenVPN Static key V1-----\n"
	endString := "-----END OpenVPN Static key V1-----"

	key := startString + formattedKeyString + endString

	return []byte(key), nil
}
