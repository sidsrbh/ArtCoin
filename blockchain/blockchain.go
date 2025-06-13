package blockchain

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
)

func VerifySignature(transaction string, signatureBase64 string, publicKeyPEM string) (bool, error) {
	// Decode the PEM-encoded public key

	fmt.Println("Transaction:", transaction)
	block, _ := pem.Decode([]byte(publicKeyPEM))
	if block == nil {
		return false, errors.New("failed to decode PEM block containing public key")
	}

	publicKey, err := x509.ParsePKCS1PublicKey(block.Bytes)
	if err != nil {
		return false, err
	}
	// Compute the SHA-256 hash of the transaction
	hashed := sha256.Sum256([]byte(transaction))

	// Decode the Base64-encoded signature
	signature, err := base64.StdEncoding.DecodeString(signatureBase64)
	if err != nil {
		fmt.Println(err.Error())
		return false, err
	}
	fmt.Println("Signature: ", signature)
	// Verify the signature
	err = rsa.VerifyPKCS1v15(publicKey, crypto.SHA256, hashed[:], signature)
	if err != nil {
		return false, err
	}
	return true, nil
}
