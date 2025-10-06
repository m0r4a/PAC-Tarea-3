package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	dirPath := os.Args[1]
	decrypt := false
	keyFile := ""

	for i := 2; i < len(os.Args); i++ {
		switch os.Args[i] {
		case "-d":
			decrypt = true
			if i+1 < len(os.Args) {
				keyFile = os.Args[i+1]
				i++
			} else {
				fmt.Fprintln(os.Stderr, "error: -d requires key file or hex key")
				printUsage()
				os.Exit(1)
			}
		default:
			fmt.Fprintf(os.Stderr, "error: unknown argument: %s\n", os.Args[i])
			printUsage()
			os.Exit(1)
		}
	}

	if decrypt {
		if keyFile == "" {
			fmt.Fprintln(os.Stderr, "error: -d requires key file or hex key")
			printUsage()
			os.Exit(1)
		}
		if err := decryptDirectory(dirPath, keyFile); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
	} else {
		if err := encryptDirectory(dirPath); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
	}
}

func printUsage() {
	fmt.Fprintf(os.Stderr, "usage: %s <directory> [-d <keyfile|hexkey>]\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "  <directory>          directory to encrypt/decrypt\n")
	fmt.Fprintf(os.Stderr, "  -d <keyfile|hexkey>  decrypt mode with key file or hex key\n")
}

func encryptDirectory(dirPath string) error {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return fmt.Errorf("generating key: %w", err)
	}

	keyFile := "encryption.key"
	if err := os.WriteFile(keyFile, []byte(hex.EncodeToString(key)), 0600); err != nil {
		return fmt.Errorf("saving key: %w", err)
	}
	fmt.Printf("Key saved to: %s\n", keyFile)
	fmt.Printf("Key (hex): %s\n\n", hex.EncodeToString(key))

	hasher := sha256.New()
	fileCount := 0

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("reading %s: %w", path, err)
		}

		hasher.Write(data)

		encrypted, err := encryptAES(data, key)
		if err != nil {
			return fmt.Errorf("encrypting %s: %w", path, err)
		}

		encPath := path + ".enc"
		if err := os.WriteFile(encPath, encrypted, info.Mode()); err != nil {
			return fmt.Errorf("writing %s: %w", encPath, err)
		}

		if err := os.Remove(path); err != nil {
			return fmt.Errorf("removing %s: %w", path, err)
		}

		fileCount++
		fmt.Printf("Encrypted: %s\n", path)
		return nil
	})

	if err != nil {
		return err
	}

	hash := hex.EncodeToString(hasher.Sum(nil))
	fmt.Printf("\nFiles processed: %d\n", fileCount)
	fmt.Printf("SHA-256 hash of original content:\n%s\n", hash)
	return nil
}

func decryptDirectory(dirPath, keyInput string) error {
	var key []byte
	var err error

	key, err = hex.DecodeString(keyInput)
	if err != nil || len(key) != 32 {
		keyHex, fileErr := os.ReadFile(keyInput)
		if fileErr != nil {
			return fmt.Errorf("invalid key format and cannot read as file: %w", fileErr)
		}
		key, err = hex.DecodeString(string(keyHex))
		if err != nil {
			return fmt.Errorf("decoding key from file: %w", err)
		}
	}

	if len(key) != 32 {
		return fmt.Errorf("invalid key length: expected 32 bytes, got %d", len(key))
	}

	hasher := sha256.New()
	fileCount := 0

	err = filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || filepath.Ext(path) != ".enc" {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("reading %s: %w", path, err)
		}

		decrypted, err := decryptAES(data, key)
		if err != nil {
			return fmt.Errorf("decrypting %s: %w", path, err)
		}

		hasher.Write(decrypted)

		origPath := path[:len(path)-4]
		if err := os.WriteFile(origPath, decrypted, info.Mode()); err != nil {
			return fmt.Errorf("writing %s: %w", origPath, err)
		}

		if err := os.Remove(path); err != nil {
			return fmt.Errorf("removing %s: %w", path, err)
		}

		fileCount++
		fmt.Printf("Decrypted: %s\n", origPath)
		return nil
	})

	if err != nil {
		return err
	}

	hash := hex.EncodeToString(hasher.Sum(nil))
	fmt.Printf("\nFiles processed: %d\n", fileCount)
	fmt.Printf("SHA-256 hash of decrypted content:\n%s\n", hash)

	keyPath := "encryption.key"
	if _, err := os.Stat(keyPath); err == nil {
		if err := os.Remove(keyPath); err != nil {
			fmt.Fprintf(os.Stderr, "warning: could not remove %s: %v\n", keyPath, err)
		} else {
			fmt.Printf("Removed: %s\n", keyPath)
		}
	}

	return nil
}

func encryptAES(data, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	return gcm.Seal(nonce, nonce, data, nil), nil
}

func decryptAES(data, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return nil, fmt.Errorf("encrypted data too short")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	return gcm.Open(nil, nonce, ciphertext, nil)
}
