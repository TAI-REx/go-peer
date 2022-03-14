package crypto

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"fmt"
	"strings"
)

var (
	_ PrivKey = &privKeyT{}
	_ PubKey  = &pubKeyT{}
)

const (
	AsymmKeyType = "go-peer\\rsa"
)

/*
 * PRIVATE KEY
 */

type privKeyT struct {
	priv *rsa.PrivateKey
}

// Create private key by number of bits.
func NewPrivKey(bits uint64) PrivKey {
	priv, err := rsa.GenerateKey(rand.Reader, int(bits))
	if err != nil {
		return nil
	}
	return &privKeyT{priv}
}

func LoadPrivKey(pbytes []byte) PrivKey {
	return &privKeyT{bytesToPrivateKey(pbytes)}
}

func LoadPrivKeyByString(pstring string) PrivKey {
	var (
		prefix = fmt.Sprintf("Priv(%s){", AsymmKeyType)
		suffix = "}"
	)
	if !strings.HasPrefix(pstring, prefix) {
		return nil
	}
	if !strings.HasSuffix(pstring, suffix) {
		return nil
	}
	pstring = strings.TrimPrefix(pstring, prefix)
	pstring = strings.TrimSuffix(pstring, suffix)
	pbytes, err := hex.DecodeString(pstring)
	if err != nil {
		return nil
	}
	return LoadPrivKey(pbytes)
}

func (key *privKeyT) Decrypt(msg []byte) []byte {
	return decryptRSA(key.priv, msg)
}

func (key *privKeyT) Sign(msg []byte) []byte {
	return sign(key.priv, NewHasher(msg).Bytes())
}

func (key *privKeyT) PubKey() PubKey {
	return &pubKeyT{&key.priv.PublicKey}
}

func (key *privKeyT) Bytes() []byte {
	return privateKeyToBytes(key.priv)
}

func (key *privKeyT) String() string {
	return fmt.Sprintf("Priv(%s){%X}", AsymmKeyType, key.Bytes())
}

func (key *privKeyT) Type() string {
	return AsymmKeyType
}

func (key *privKeyT) Size() uint64 {
	return key.PubKey().Size()
}

// Used PKCS1.
func bytesToPrivateKey(privData []byte) *rsa.PrivateKey {
	priv, err := x509.ParsePKCS1PrivateKey(privData)
	if err != nil {
		return nil
	}
	return priv
}

// Used RSA(OAEP).
func decryptRSA(priv *rsa.PrivateKey, data []byte) []byte {
	data, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, priv, data, nil)
	if err != nil {
		return nil
	}
	return data
}

// Used PKCS1.
func privateKeyToBytes(priv *rsa.PrivateKey) []byte {
	return x509.MarshalPKCS1PrivateKey(priv)
}

func sign(priv *rsa.PrivateKey, hash []byte) []byte {
	signature, err := rsa.SignPSS(rand.Reader, priv, crypto.SHA256, hash, nil)
	if err != nil {
		return nil
	}
	return signature
}

/*
 * PUBLIC KEY
 */

type pubKeyT struct {
	pub *rsa.PublicKey
}

func LoadPubKey(pbytes []byte) PubKey {
	return &pubKeyT{bytesToPublicKey(pbytes)}
}

func LoadPubKeyByString(pstring string) PubKey {
	var (
		prefix = fmt.Sprintf("Pub(%s){", AsymmKeyType)
		suffix = "}"
	)
	if !strings.HasPrefix(pstring, prefix) {
		return nil
	}
	if !strings.HasSuffix(pstring, suffix) {
		return nil
	}
	pstring = strings.TrimPrefix(pstring, prefix)
	pstring = strings.TrimSuffix(pstring, suffix)
	pbytes, err := hex.DecodeString(pstring)
	if err != nil {
		return nil
	}
	return LoadPubKey(pbytes)
}

func (key *pubKeyT) Encrypt(msg []byte) []byte {
	return encryptRSA(key.pub, msg)
}

func (key *pubKeyT) Address() string {
	return NewHasher(key.Bytes()).String()
}

func (key *pubKeyT) Verify(msg []byte, sig []byte) bool {
	return verify(key.pub, NewHasher(msg).Bytes(), sig) == nil
}

func (key *pubKeyT) Bytes() []byte {
	return publicKeyToBytes(key.pub)
}

func (key *pubKeyT) String() string {
	return fmt.Sprintf("Pub(%s){%X}", AsymmKeyType, key.Bytes())
}

func (key *pubKeyT) Type() string {
	return AsymmKeyType
}

func (key *pubKeyT) Size() uint64 {
	return uint64(key.pub.N.BitLen())
}

// Used RSA(OAEP).
func encryptRSA(pub *rsa.PublicKey, data []byte) []byte {
	data, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, pub, data, nil)
	if err != nil {
		return nil
	}
	return data
}

// Used PKCS1.
func bytesToPublicKey(pubData []byte) *rsa.PublicKey {
	pub, err := x509.ParsePKCS1PublicKey(pubData)
	if err != nil {
		return nil
	}
	return pub
}

// Used PKCS1.
func publicKeyToBytes(pub *rsa.PublicKey) []byte {
	return x509.MarshalPKCS1PublicKey(pub)
}

// Used RSA(PSS).
func verify(pub *rsa.PublicKey, hash, sign []byte) error {
	return rsa.VerifyPSS(pub, crypto.SHA256, hash, sign, nil)
}
