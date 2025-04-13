package rsahelpers

import (
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
)

type OAEPDecoder struct {
	privateKey *rsa.PrivateKey
}

func NewOAEPDecoder(privateKey []byte) (*OAEPDecoder, error) {
	prv, err := x509.ParsePKCS1PrivateKey(privateKey)
	if err != nil {
		return nil, err
	}
	return &OAEPDecoder{
		privateKey: prv,
	}, nil
}

func (d *OAEPDecoder) Decode(chiper []byte) ([]byte, error) {
	return rsa.DecryptOAEP(sha256.New(), nil, d.privateKey, chiper, nil)
}

type OAEPEncoder struct {
	publicKey *rsa.PublicKey
}

func NewOAEPEncoder(publicKey []byte) (*OAEPEncoder, error) {
	pub, err := x509.ParsePKCS1PublicKey(publicKey)
	if err != nil {
		return nil, err
	}
	return &OAEPEncoder{
		publicKey: pub,
	}, nil
}

func (e *OAEPEncoder) Encode(plain []byte) ([]byte, error) {
	return rsa.EncryptOAEP(sha256.New(), nil, e.publicKey, plain, nil)
}
