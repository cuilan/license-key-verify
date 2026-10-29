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

func TestPEMRoundTrip(t *testing.T) {
	keyPair, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair() error = %v", err)
	}

	privateKeyPEM, err := keyPair.PrivateKeyToPEM()
	if err != nil {
		t.Fatalf("PrivateKeyToPEM() error = %v", err)
	}
	publicKeyPEM, err := keyPair.PublicKeyToPEM()
	if err != nil {
		t.Fatalf("PublicKeyToPEM() error = %v", err)
	}

	loadedPrivate, err := LoadPrivateKeyFromPEM(privateKeyPEM)
	if err != nil {
		t.Errorf("LoadPrivateKeyFromPEM() error = %v", err)
	}
	loadedPublic, err := LoadPublicKeyFromPEM(publicKeyPEM)
	if err != nil {
		t.Errorf("LoadPublicKeyFromPEM() error = %v", err)
	}

	if loadedPrivate.D.Sign() != keyPair.PrivateKey.D.Sign() {
		t.Error("loaded private key differs from original")
	}
	if loadedPublic.N.Sign() != keyPair.PublicKey.N.Sign() {
		t.Error("loaded public key differs from original")
	}
}

func TestRSASignVerify(t *testing.T) {
	keyPair, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair() error = %v", err)
	}

	data := []byte("This is test data for signing")

	signature, err := SignData(data, keyPair.PrivateKey)
	if err != nil {
		t.Errorf("SignData() error = %v", err)
		return
	}

	if len(signature) == 0 {
		t.Error("SignData() returned empty signature")
		return
	}

	if err := VerifySignature(data, signature, keyPair.PublicKey); err != nil {
		t.Errorf("VerifySignature() error = %v", err)
	}

	wrongData := []byte("Wrong data")
	if err := VerifySignature(wrongData, signature, keyPair.PublicKey); err == nil {
		t.Error("VerifySignature() should fail with wrong data")
	}
}
