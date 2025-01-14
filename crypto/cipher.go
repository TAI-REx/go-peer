package crypto

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
)

var (
	_ ICipher = &sCipher{}
)

const (
	SymmKeyType = "go-peer\\aes"
)

type sCipher struct {
	fKey []byte
}

func NewCipher(key []byte) ICipher {
	switch HashSize {
	case 16, 24, 32: // AES keys
	default:
		return nil
	}
	return &sCipher{
		fKey: NewHasher(key).Bytes(),
	}
}

func (cph *sCipher) Encrypt(msg []byte) []byte {
	block, err := aes.NewCipher(cph.fKey)
	if err != nil {
		return nil
	}
	blockSize := block.BlockSize()
	msg = paddingPKCS5(msg, blockSize)
	cipherText := make([]byte, blockSize+len(msg))
	iv := cipherText[:blockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil
	}
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(cipherText[blockSize:], msg)
	return cipherText
}

func (cph *sCipher) Decrypt(msg []byte) []byte {
	block, err := aes.NewCipher(cph.fKey)
	if err != nil {
		return nil
	}
	blockSize := block.BlockSize()
	if len(msg) < blockSize {
		return nil
	}
	iv := msg[:blockSize]
	msg = msg[blockSize:]
	if len(msg)%blockSize != 0 {
		return nil
	}
	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(msg, msg)
	return unpaddingPKCS5(msg)
}

func (cph *sCipher) String() string {
	return fmt.Sprintf("Key(%s){%X}", SymmKeyType, cph.Bytes())
}

func (cph *sCipher) Bytes() []byte {
	return cph.fKey
}

func (cph *sCipher) Type() string {
	return SymmKeyType
}

func (cph *sCipher) Size() uint64 {
	return HashSize
}

func paddingPKCS5(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

func unpaddingPKCS5(origData []byte) []byte {
	length := len(origData)
	if length == 0 {
		return nil
	}
	unpadding := int(origData[length-1])
	if length < unpadding {
		return nil
	}
	return origData[:(length - unpadding)]
}
