package common

import (
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
)

func EncryptRSAPrivateKey(prikey []byte, passwd string) ([]byte, error) {

	skBlock, _ := pem.Decode(prikey)
	encPriKey, err := x509.EncryptPEMBlock(rand.Reader, "RSA PRIVATE KEY", skBlock.Bytes, []byte(passwd), x509.PEMCipherAES256)
	if err != nil {
		return nil, err
	}
	return pem.EncodeToMemory(encPriKey), nil
}

func DecryptRSAPrivateKey(encPrikey []byte, passwd string) ([]byte, error) {

	skBlock, _ := pem.Decode(encPrikey)
	prikey, err := x509.DecryptPEMBlock(skBlock, []byte(passwd))
	if err != nil {
		return nil, err
	}
	sk, err := x509.ParsePKCS8PrivateKey(prikey)
	if err != nil {
		return nil, err
	}
	privateKeyBytes, err := x509.MarshalPKCS8PrivateKey(sk)
	if err != nil {
		return nil, err
	}
	return pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	}), nil
}
