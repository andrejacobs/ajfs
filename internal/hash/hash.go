package hash

import (
	"crypto"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"hash"
)

// Type of hashing algorithm used.
type Algo uint8

const (
	AlgoSHA1   Algo = 1 + iota // SHA-1
	AlgoSHA256                 // SHA-256
	AlgoSHA512                 // SHA-512
)

const (
	DefaultAlgo = AlgoSHA256 // The default hash algorithm is SHA-256
)

var (
	// Used to write out zeroes for an uncalculated hash wihtout doing an alloc
	// NOTE: Don't create duplicate buffers, e.g. next 64 byte algo can just point to the same SHA512 one
	AlgoSHA1Zero   = make([]byte, sha1.Size)   // 20 bytes
	AlgoSHA256Zero = make([]byte, sha256.Size) // 32 bytes
	AlgoSHA512Zero = make([]byte, sha512.Size) // 64 bytes
)

// Return the size of bytes that a digest for the hashing algorithm uses.
func (h Algo) Size() int {
	return h.cryptoHash().Size()
}

func (h Algo) cryptoHash() crypto.Hash {
	switch h {
	case AlgoSHA1:
		return crypto.SHA1
	case AlgoSHA256:
		return crypto.SHA256
	case AlgoSHA512:
		return crypto.SHA512
	default:
		panic("not yet implemented!")
	}
}

// Stringer implementation.
func (h Algo) String() string {
	switch h {
	case AlgoSHA1:
		return "SHA-1"
	case AlgoSHA256:
		return "SHA-256"
	case AlgoSHA512:
		return "SHA-512"
	default:
		return "unknown"
	}
}

// Return the hash (as a string) for when zero bytes are hashed.
func (h Algo) HashedStringForZeroBytes() string {
	switch h {
	case AlgoSHA1:
		// shasum -a 1 /dev/null
		return "da39a3ee5e6b4b0d3255bfef95601890afd80709"
	case AlgoSHA256:
		// shasum -a 256 /dev/null
		return "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
	case AlgoSHA512:
		// shasum -a 512 /dev/null
		return "cf83e1357eefb8bdf1542850d66d8007d620e4050b5715dc83f4a921d36ce9ce47d0d13c5d85f2b0ff8318d2877eec2f63b931bd47417a81a538327af927da3e"
	default:
		return ""
	}
}

// Return the hasher to be used for the algorithm.
func (h Algo) Hasher() hash.Hash {
	switch h {
	case AlgoSHA1:
		return sha1.New()
	case AlgoSHA256:
		return sha256.New()
	case AlgoSHA512:
		return sha512.New()
	default:
		panic("unknown hashing algorithm")
	}
}

// Return true if all the bytes in the slice are zero.
func AllZeroBytes(buf []byte) bool {
	for _, b := range buf {
		if b != 0 {
			return false
		}
	}
	return true
}
