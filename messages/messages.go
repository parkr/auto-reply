// Remove this when https://github.com/google/go-github/pull/294 is merged
// Copyright 2016 The go-github AUTHORS. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file provides functions for validating payloads from GitHub Webhooks.
// GitHub docs: https://developer.github.com/webhooks/securing/#validating-payloads-from-github

package messages

import (
	"crypto/hmac"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"errors"
	"fmt"
	"hash"
	"io/ioutil"
	"net/http"
	"strings"
)

const (
	// sha1Prefix is the prefix used by Github before the HMAC hexdigest.
	sha1Prefix = "sha1"
	// sha256Prefix and sha512Prefix are provided for future compatibility.
	sha256Prefix = "sha256"
	sha512Prefix = "sha512"
	// eventHeader is the Github header key used to pass the event type.
	eventHeader = "X-GitHub-Event"
	// signatureHeader is the Github header key used to pass the HMAC hexdigest.
	signatureHeader = "X-Hub-Signature"
	// deliveryHeader is the Gihub header key used to pass the globally-unique Github Event ID.
	deliveryHeader = "X-GitHub-Delivery"
)

// genMAC generates the HMAC signature for a message provided the secret key and hashFunc.
// If hashFunc is nil, sha1.New is used.
func genMAC(message, key []byte, hashFunc func() hash.Hash) []byte {
	if hashFunc == nil {
		hashFunc = sha1.New
	}
	mac := hmac.New(hashFunc, key)
	mac.Write(message)
	return mac.Sum(nil)
}

// checkMAC reports whether messageMAC is a valid HMAC tag for message.
// If hashFunc is nil, sha1.New is used.
func checkMAC(message, messageMAC, key []byte, hashFunc func() hash.Hash) bool {
	expectedMAC := genMAC(message, key, hashFunc)
	return hmac.Equal(messageMAC, expectedMAC)
}

// messageMAC returns the hex-decoded HMAC tag from the signature and its
// corresponding hash function.
func messageMAC(signature string) ([]byte, func() hash.Hash, error) {
	if signature == "" {
		return nil, nil, errors.New("missing signature")
	}
	sigParts := strings.SplitN(signature, "=", 2)
	if len(sigParts) != 2 {
		return nil, nil, fmt.Errorf("error parsing signature %q", signature)
	}

	var hashFunc func() hash.Hash
	switch sigParts[0] {
	case sha1Prefix:
		hashFunc = sha1.New
	case sha256Prefix:
		hashFunc = sha256.New
	case sha512Prefix:
		hashFunc = sha512.New
	default:
		return nil, nil, fmt.Errorf("unknown hash type prefix: %q", sigParts[0])
	}

	buf, err := hex.DecodeString(sigParts[1])
	if err != nil {
		return nil, nil, fmt.Errorf("error decoding signature %q: %v", signature, err)
	}
	return buf, hashFunc, nil
}

// ValidatedPayload validates an incoming Github Webhook event request
// and returns the (JSON) payload.
// secretKey is the GitHub Webhook secret message.
func ValidatedPayload(r *http.Request, secretKey []byte) (payload []byte, err error) {
	payload, err = ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	sig := r.Header.Get(signatureHeader)
	if err := ValidateSignature(sig, payload, secretKey); err != nil {
		return nil, err
	}
	return payload, nil
}

// ValidateSignature validates the signature for the given payload.
// signature is the GitHub hash signature delivered in the X-Hub-Signature header.
// payload is the JSON payload sent by GitHub Webhooks.
// secretKey is the GitHub Webhook secret message.
// GitHub docs: https://developer.github.com/webhooks/securing/#validating-payloads-from-github
func ValidateSignature(signature string, payload, secretKey []byte) error {
	messageMAC, hashFunc, err := messageMAC(signature)
	if err != nil {
		return err
	}
	if !checkMAC(payload, messageMAC, secretKey, hashFunc) {
		return errors.New("payload signature check failed")
	}
	return nil
}
