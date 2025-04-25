package rsahelpers

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"hash"
	"io"
)

type OAEPDecoder struct {
	privateKey *rsa.PrivateKey
}

func NewOAEPDecoder(privatePem []byte) (*OAEPDecoder, error) {
	block, _ := pem.Decode(privatePem)
	if block == nil || block.Type != "PRIVATE KEY" {
		return nil, errors.New("invalid private key")
	}
	prv, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("invalid private key: %w", err)
	}
	return &OAEPDecoder{
		privateKey: prv.(*rsa.PrivateKey),
	}, nil
}

func (d *OAEPDecoder) Decode(chiper []byte) ([]byte, error) {
	return DecryptOAEP(sha256.New(), d.privateKey, chiper, nil)
}

func DecryptOAEP(hash hash.Hash, private *rsa.PrivateKey, msg []byte, label []byte) ([]byte, error) {
	msgLen := len(msg)
	step := private.PublicKey.Size()
	var decryptedBytes []byte

	for start := 0; start < msgLen; start += step {
		finish := start + step
		if finish > msgLen {
			finish = msgLen
		}

		decryptedBlockBytes, err := rsa.DecryptOAEP(hash, nil, private, msg[start:finish], label)
		if err != nil {
			return nil, fmt.Errorf("block decryption failed: %w", err)
		}

		decryptedBytes = append(decryptedBytes, decryptedBlockBytes...)
	}

	return decryptedBytes, nil
}

type OAEPEncoder struct {
	publicKey *rsa.PublicKey
}

func NewOAEPEncoder(publicPem []byte) (*OAEPEncoder, error) {
	block, _ := pem.Decode(publicPem)
	if block == nil || block.Type != "PUBLIC KEY" {
		return nil, errors.New("invalid public key")
	}
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("invalid public key: %w", err)
	}
	return &OAEPEncoder{
		publicKey: pub.(*rsa.PublicKey),
	}, nil
}

func (e *OAEPEncoder) Encode(plain []byte) ([]byte, error) {

	return EncryptOAEP(sha256.New(), rand.Reader, e.publicKey, plain, nil)
}

func EncryptOAEP(hash hash.Hash, random io.Reader, public *rsa.PublicKey, msg []byte, label []byte) ([]byte, error) {
	msgLen := len(msg)
	step := public.Size() - 2*hash.Size() - 2
	var encryptedBytes []byte

	for start := 0; start < msgLen; start += step {
		finish := start + step
		if finish > msgLen {
			finish = msgLen
		}
		encryptedBlockBytes, err := rsa.EncryptOAEP(hash, random, public, msg[start:finish], label)
		if err != nil {
			return nil, fmt.Errorf("block encryption failed: %w", err)
		}

		encryptedBytes = append(encryptedBytes, encryptedBlockBytes...)
	}

	return encryptedBytes, nil
}
