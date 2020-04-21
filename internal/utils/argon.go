package utils

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"runtime"
	"strconv"
	"strings"

	"golang.org/x/crypto/argon2"
)

const (
	passwordType = "argon2id"
	saltLen      = 32
	argon2KeyLen = 32
	argon2Time   = 1
	argon2Memory = 64 * 1024
)

var argon2Threads = uint8(runtime.NumCPU())

//GenerateHash for salted passwords with argon2
func GenerateHash(inString string) string {
	salt, _ := generateSalt(saltLen)
	byteHash := argon2.IDKey([]byte(inString), []byte(salt), argon2Time, argon2Memory, argon2Threads, argon2KeyLen)
	encHash := base64.StdEncoding.EncodeToString(byteHash)
	hash := fmt.Sprintf("%s$%d$%d$%d$%d$%s$%s",
		passwordType, argon2Time, argon2Memory, argon2Threads, argon2KeyLen, salt, encHash)
	return hash
}

//CompareHash compare a argon2 hash string with a plain string
func CompareHash(hash string, inString string) (bool, error) {
	hashParts := strings.Split(hash, "$")
	if len(hashParts) != 7 {
		return false, errors.New("Invalid Hash Segments")
	}

	hashType := hashParts[0]
	time, _ := strconv.Atoi((hashParts[1]))
	memory, _ := strconv.Atoi(hashParts[2])
	threads, _ := strconv.Atoi(hashParts[3])
	keyLen, _ := strconv.Atoi(hashParts[4])
	salt := []byte(hashParts[5])
	key, _ := base64.StdEncoding.DecodeString(hashParts[6])

	var calculatedKey []byte
	switch hashType {
	case "argon2id":
		calculatedKey = argon2.IDKey([]byte(inString), salt, uint32(time), uint32(memory), uint8(threads), uint32(keyLen))
	case "argon2i", "argon2":
		calculatedKey = argon2.Key([]byte(inString), salt, uint32(time), uint32(memory), uint8(threads), uint32(keyLen))
	default:
		return false, errors.New("Invalid Hash Type")
	}

	if subtle.ConstantTimeCompare(key, calculatedKey) != 1 {
		return false, nil
	}
	return true, nil
}

func generateSalt(len int) (string, error) {
	bSalt := make([]byte, len)
	if _, err := rand.Read(bSalt); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(bSalt), nil
}
