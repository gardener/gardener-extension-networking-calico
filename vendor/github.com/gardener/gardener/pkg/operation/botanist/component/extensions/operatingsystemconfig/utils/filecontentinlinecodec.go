// Copyright (c) 2019 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
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

package utils

import (
	"github.com/gardener/gardener/extensions/pkg/controller/operatingsystemconfig/oscommon/cloudinit"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"

	"github.com/pkg/errors"
)

// FileContentInlineCodec contains methods for encoding and decoding byte slices
// to and from *extensionsv1alpha1.FileContentInline.
type FileContentInlineCodec interface {
	// Encode encodes the given byte slice into a *extensionsv1alpha1.FileContentInline.
	Encode([]byte, string) (*extensionsv1alpha1.FileContentInline, error)
	// Decode decodes a byte slice from the given *extensionsv1alpha1.FileContentInline.
	Decode(*extensionsv1alpha1.FileContentInline) ([]byte, error)
}

// NewFileContentInlineCodec creates an returns a new FileContentInlineCodec.
func NewFileContentInlineCodec() FileContentInlineCodec {
	return &fileContentInlineCodec{}
}

type fileContentInlineCodec struct{}

// Encode encodes the given byte slice into a *extensionsv1alpha1.FileContentInline.
func (c *fileContentInlineCodec) Encode(data []byte, encoding string) (*extensionsv1alpha1.FileContentInline, error) {
	// Initialize file codec
	fileCodec, err := getFileCodec(encoding)
	if err != nil {
		return nil, err
	}

	// Encode data using the file codec, if needed
	if fileCodec != nil {
		if data, err = fileCodec.Encode(data); err != nil {
			return nil, errors.Wrap(err, "could not encode data using file codec")
		}
	}

	return &extensionsv1alpha1.FileContentInline{
		Encoding: encoding,
		Data:     string(data),
	}, nil
}

// Decode decodes a byte slice from the given *extensionsv1alpha1.FileContentInline.
func (c *fileContentInlineCodec) Decode(fci *extensionsv1alpha1.FileContentInline) ([]byte, error) {
	data := []byte(fci.Data)

	// Initialize file codec
	fileCodec, err := getFileCodec(fci.Encoding)
	if err != nil {
		return nil, err
	}

	// Decode data using the file codec, if needed
	if fileCodec != nil {
		if data, err = fileCodec.Decode(data); err != nil {
			return nil, errors.Wrap(err, "could not decode data using file codec")
		}
	}

	return data, nil
}

func getFileCodec(encoding string) (cloudinit.FileCodec, error) {
	if encoding == "" {
		return nil, nil
	}
	fileCodecID, err := cloudinit.ParseFileCodecID(encoding)
	if err != nil {
		return nil, errors.Wrapf(err, "could not parse file codec ID '%s'", encoding)
	}
	return cloudinit.FileCodecForID(fileCodecID), nil
}
