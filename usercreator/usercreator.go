package usercreator

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"indicartcoin/sqldatabase"
	"io"
	"log"
	"net/http"
)

type LoginSignUpResponse struct {
	Message    string `json:"message"`
	PrivateKey string `json:"privateKey"`
	PublicKey  string `json:"publicKey"`
}

// Simulate database
var Database map[string][]string

type PrivateKey struct {
	PEM string `json:"pem"`
	P   string `json:"p"`
	Q   string `json:"q"`
}

func generateKeyPair(bits int) (string, string, error) {
	// Generate RSA key pair
	privateKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return "", "", err
	}

	// Encode private key to PEM format
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyPEM := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: privateKeyBytes,
		},
	)

	// Encode public key to PEM format
	publicKeyBytes := x509.MarshalPKCS1PublicKey(&privateKey.PublicKey)
	publicKeyPEM := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PUBLIC KEY",
			Bytes: publicKeyBytes,
		},
	)

	// Serialize the private key, including p and q
	privateKeySerialized, err := json.Marshal(PrivateKey{
		PEM: string(privateKeyPEM),
		P:   privateKey.Primes[0].String(),
		Q:   privateKey.Primes[1].String(),
	})
	if err != nil {
		return "", "", err
	}

	return string(privateKeySerialized), string(publicKeyPEM), nil
}

func decrypt(ciphertextHex string, passphrase string) (string, error) {
	if len(passphrase) != 16 && len(passphrase) != 24 && len(passphrase) != 32 {
		return "", fmt.Errorf("invalid passphrase length: %d", len(passphrase))
	}

	// Decode the hexadecimal string to get the ciphertext
	ciphertext, err := hex.DecodeString(ciphertextHex)
	if err != nil {
		return "", fmt.Errorf("failed to decode hex string: %v", err)
	}

	// Create a new cipher block from the passphrase
	block, err := aes.NewCipher([]byte(passphrase))
	if err != nil {
		return "", fmt.Errorf("failed to create cipher block: %v", err)
	}

	// Check if the ciphertext is too short
	if len(ciphertext) < aes.BlockSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	// Separate the IV from the ciphertext
	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	// Decrypt the ciphertext using CBC mode
	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(ciphertext, ciphertext)

	// Unpad the plaintext
	plaintext := pkcs7Unpad(ciphertext)
	if plaintext == nil {
		return "", fmt.Errorf("failed to unpad plaintext")
	}
	fmt.Println(string(plaintext))
	return string(plaintext), nil
}

func pkcs7Unpad(data []byte) []byte {
	length := len(data)
	unpadding := int(data[length-1])

	if unpadding > length {
		return nil
	}

	return data[:(length - unpadding)]
}

func storeInDatabase(username string, encryptedPrivateKey string, publicKey string) {
	Database[username] = []string{string(encryptedPrivateKey), string(publicKey)}
	sqldatabase.AddBalances(string(publicKey), 0)
	//update sql database
	sqldatabase.AddUser(username, []string{string(encryptedPrivateKey), string(publicKey)})
}

func retrieveFromDatabase(username string) (string, string) {
	if data, found := Database[username]; found {
		return data[0], data[1]
	}
	return "", ""
}

func randomString(length int) string {
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		log.Fatal(err)
	}
	return hex.EncodeToString(bytes)
}

func encrypt(plainText string, passphrase string) (string, error) {
	if len(passphrase) != 16 && len(passphrase) != 24 && len(passphrase) != 32 {
		return "", fmt.Errorf("invalid passphrase length: %d", len(passphrase))
	}

	block, err := aes.NewCipher([]byte(passphrase))
	if err != nil {
		return "", err
	}

	paddedText := pkcs7Pad([]byte(plainText), block.BlockSize())
	ciphertext := make([]byte, aes.BlockSize+len(paddedText))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}

	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext[aes.BlockSize:], paddedText)

	return hex.EncodeToString(ciphertext), nil
}

func pkcs7Pad(b []byte, blocksize int) []byte {
	padding := blocksize - len(b)%blocksize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(b, padtext...)
}

func SignupHandler(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("username")
	passphrase := r.URL.Query().Get("passphrase")

	if len(passphrase) != 16 && len(passphrase) != 24 && len(passphrase) != 32 {
		http.Error(w, fmt.Sprintf("Invalid passphrase length: %d", len(passphrase)), http.StatusBadRequest)
		return
	}

	privateKey, publicKey, err := generateKeyPair(2048)
	if err != nil {
		http.Error(w, "Key generation failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	encryptedPrivateKey, err := encrypt(privateKey, passphrase)
	if err != nil {
		http.Error(w, "Encryption failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	storeInDatabase(username, encryptedPrivateKey, publicKey)
	response := LoginSignUpResponse{
		Message:    "SignUp successful",
		PrivateKey: string(encryptedPrivateKey),
		PublicKey:  string(publicKey),
	}

	jsonResponse, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "JSON marshaling failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResponse)
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("username")
	passphrase := r.URL.Query().Get("passphrase")

	encryptedPrivateKey, publicKey := retrieveFromDatabase(username)
	if encryptedPrivateKey == "" || publicKey == "" {
		fmt.Println("No User:", username)
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	_, err := decrypt(encryptedPrivateKey, passphrase)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	response := LoginSignUpResponse{
		Message:    "Login successful",
		PrivateKey: string(encryptedPrivateKey),
		PublicKey:  string(publicKey),
	}

	jsonResponse, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "JSON marshaling failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResponse)
}
