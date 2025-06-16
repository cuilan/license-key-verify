package crypto

import (
	"testing"
)

func TestGenerateKeyPair(t *testing.T) {
	keyPair, err := GenerateKeyPair()
	if err != nil {
		t.Errorf("GenerateKeyPair() error = %v", err)
		return
	}

	if keyPair.PrivateKey == nil {
		t.Error("GenerateKeyPair() returned nil private key")
	}
	if keyPair.PublicKey == nil {
		t.Error("GenerateKeyPair() returned nil public key")
	}
}

func TestGenerateAESKey(t *testing.T) {
	key, err := GenerateAESKey()
	if err != nil {
		t.Errorf("GenerateAESKey() error = %v", err)
		return
	}

	if len(key) != 32 {
		t.Errorf("GenerateAESKey() returned key with wrong length: got %d, want 32", len(key))
	}
}

func TestAESEncryptDecrypt(t *testing.T) {
	key, err := GenerateAESKey()
	if err != nil {
		t.Fatalf("GenerateAESKey() error = %v", err)
	}

	plaintext := []byte("Hello, World! This is a test message.")

	// 测试加密
	ciphertext, err := EncryptAES(plaintext, key)
	if err != nil {
		t.Errorf("EncryptAES() error = %v", err)
		return
	}

	if len(ciphertext) == 0 {
		t.Error("EncryptAES() returned empty ciphertext")
		return
	}

	// 测试解密
	decrypted, err := DecryptAES(ciphertext, key)
	if err != nil {
		t.Errorf("DecryptAES() error = %v", err)
		return
	}

	if string(decrypted) != string(plaintext) {
		t.Errorf("DecryptAES() = %s, want %s", string(decrypted), string(plaintext))
	}
}

func TestRSASignVerify(t *testing.T) {
	keyPair, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair() error = %v", err)
	}

	data := []byte("This is test data for signing")

	// 测试签名
	signature, err := SignData(data, keyPair.PrivateKey)
	if err != nil {
		t.Errorf("SignData() error = %v", err)
		return
	}

	if len(signature) == 0 {
		t.Error("SignData() returned empty signature")
		return
	}

	// 测试验证
	err = VerifySignature(data, signature, keyPair.PublicKey)
	if err != nil {
		t.Errorf("VerifySignature() error = %v", err)
	}

	// 测试错误数据验证
	wrongData := []byte("Wrong data")
	err = VerifySignature(wrongData, signature, keyPair.PublicKey)
	if err == nil {
		t.Error("VerifySignature() should fail with wrong data")
	}
}
