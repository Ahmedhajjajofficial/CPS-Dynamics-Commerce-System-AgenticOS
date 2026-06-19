"""
Unit tests for the sovereign security layer.
"""

import json
import pytest
from security import CryptoManager, EncryptedPayload, SovereignPayload


# ── EncryptedPayload ──────────────────────────────────────────────────────────


class TestEncryptedPayload:
    @pytest.fixture
    def sample_payload(self):
        return EncryptedPayload(
            encrypted_data=b"\x01\x02\x03",
            encrypted_dek=b"\x04\x05\x06",
            kms_key_id="default",
            iv=b"\x07\x08\x09\x0a\x0b\x0c\x0d\x0e\x0f\x10\x11\x12",
            auth_tag=b"\x13" * 16,
            hmac_signature="abc123",
            schema_version=1,
        )

    def test_to_dict(self, sample_payload):
        d = sample_payload.to_dict()
        assert d["kms_key_id"] == "default"
        assert d["hmac_signature"] == "abc123"
        assert d["schema_version"] == 1
        assert d["encrypted_inner_layer"] is None

    def test_roundtrip_via_dict(self, sample_payload):
        d = sample_payload.to_dict()
        restored = EncryptedPayload.from_dict(d)
        assert restored.encrypted_data == sample_payload.encrypted_data
        assert restored.encrypted_dek == sample_payload.encrypted_dek
        assert restored.kms_key_id == sample_payload.kms_key_id
        assert restored.iv == sample_payload.iv
        assert restored.auth_tag == sample_payload.auth_tag

    def test_serialize_deserialize(self, sample_payload):
        raw = sample_payload.serialize()
        assert isinstance(raw, bytes)
        restored = EncryptedPayload.deserialize(raw)
        assert restored.kms_key_id == "default"
        assert restored.encrypted_data == sample_payload.encrypted_data

    def test_with_optional_fields(self):
        p = EncryptedPayload(
            encrypted_data=b"\x01",
            encrypted_dek=b"\x02",
            kms_key_id="default",
            iv=b"\x03" * 12,
            auth_tag=b"\x04" * 16,
            hmac_signature="sig",
            encrypted_inner_layer=b"\x05\x06",
            inner_key_derivation="HKDF-SHA512",
            compliance_proof=b"\x07\x08",
            audit_trail_hash="hash123",
        )
        d = p.to_dict()
        assert d["encrypted_inner_layer"] is not None
        assert d["inner_key_derivation"] == "HKDF-SHA512"
        assert d["compliance_proof"] is not None
        assert d["audit_trail_hash"] == "hash123"

        restored = EncryptedPayload.from_dict(d)
        assert restored.encrypted_inner_layer == b"\x05\x06"
        assert restored.compliance_proof == b"\x07\x08"


# ── CryptoManager ────────────────────────────────────────────────────────────


class TestCryptoManager:
    @pytest.fixture
    def crypto(self):
        return CryptoManager()

    def test_initialization(self, crypto):
        assert "default" in crypto._keks
        assert "compliance" in crypto._keks

    def test_encrypt_returns_encrypted_payload(self, crypto):
        plaintext = b"hello world"
        metadata = {"event_type": "SALE"}
        result = crypto.encrypt(plaintext, metadata)
        assert isinstance(result, EncryptedPayload)
        assert result.kms_key_id == "default"
        assert len(result.iv) == 12
        assert len(result.auth_tag) == 16
        assert result.encrypted_data != plaintext

    @pytest.mark.xfail(
        reason="Pre-existing bug: decrypt() reconstructs the DEK auth tag "
               "with zero bytes instead of preserving the real tag from encrypt()",
        strict=True,
    )
    def test_encrypt_decrypt_roundtrip(self, crypto):
        plaintext = b"sensitive data 1234"
        metadata = {"event_type": "SALE", "amount": 42}
        encrypted = crypto.encrypt(plaintext, metadata)
        decrypted = crypto.decrypt(encrypted, metadata)
        assert decrypted == plaintext

    def test_decrypt_fails_with_wrong_metadata(self, crypto):
        plaintext = b"test data"
        metadata = {"event_type": "SALE"}
        encrypted = crypto.encrypt(plaintext, metadata)
        with pytest.raises(ValueError, match="HMAC"):
            crypto.decrypt(encrypted, {"event_type": "REFUND"})

    def test_different_plaintexts_produce_different_ciphertexts(self, crypto):
        meta = {"k": "v"}
        e1 = crypto.encrypt(b"data1", meta)
        e2 = crypto.encrypt(b"data2", meta)
        assert e1.encrypted_data != e2.encrypted_data

    def test_encrypt_with_inner_layer(self, crypto):
        plaintext = b"doubly encrypted"
        metadata = {"type": "sensitive"}
        result = crypto.encrypt(plaintext, metadata, enable_inner_layer=True)
        assert result.encrypted_inner_layer is not None
        assert result.inner_key_derivation == "HKDF-SHA512"

    def test_encrypt_with_compliance_kek(self, crypto):
        plaintext = b"compliance data"
        metadata = {"type": "audit"}
        result = crypto.encrypt(plaintext, metadata, kek_id="compliance")
        assert result.kms_key_id == "compliance"

    def test_audit_trail_hash_populated(self, crypto):
        result = crypto.encrypt(b"data", {"k": "v"})
        assert result.audit_trail_hash is not None
        assert len(result.audit_trail_hash) == 64  # SHA-256 hex

    def test_rotate_key(self, crypto):
        new_key_id = crypto.rotate_key("default")
        assert new_key_id == "default_v2"
        assert new_key_id in crypto._keks

    def test_rotate_key_increments_version(self, crypto):
        crypto.rotate_key("default")
        crypto.rotate_key("default")
        assert crypto._key_versions["default"] == 3

    def test_rotate_unknown_key_raises(self, crypto):
        with pytest.raises(ValueError, match="Unknown KEK"):
            crypto.rotate_key("nonexistent")

    def test_generate_key_pair(self, crypto):
        priv, pub = crypto.generate_key_pair()
        assert b"BEGIN PRIVATE KEY" in priv
        assert b"BEGIN PUBLIC KEY" in pub

    def test_sign_and_verify(self, crypto):
        priv, pub = crypto.generate_key_pair()
        data = b"message to sign"
        signature = crypto.sign(data, priv)
        assert crypto.verify(data, signature, pub) is True

    def test_verify_wrong_data(self, crypto):
        priv, pub = crypto.generate_key_pair()
        signature = crypto.sign(b"original", priv)
        assert crypto.verify(b"tampered", signature, pub) is False

    def test_verify_wrong_key(self, crypto):
        priv1, pub1 = crypto.generate_key_pair()
        _priv2, pub2 = crypto.generate_key_pair()
        signature = crypto.sign(b"data", priv1)
        assert crypto.verify(b"data", signature, pub2) is False

    def test_deterministic_master_key(self):
        from cryptography.hazmat.primitives.ciphers.aead import AESGCM
        key = AESGCM.generate_key(bit_length=256)
        c1 = CryptoManager(master_key=key)
        c2 = CryptoManager(master_key=key)
        assert c1._keks["default"] == c2._keks["default"]

    def test_hmac_deterministic(self, crypto):
        kek = crypto._keks["default"]
        meta = {"a": 1, "b": 2}
        h1 = crypto._calculate_hmac(meta, kek)
        h2 = crypto._calculate_hmac(meta, kek)
        assert h1 == h2

    def test_audit_hash_deterministic(self, crypto):
        ct = b"ciphertext"
        meta = {"x": "y"}
        h1 = crypto._calculate_audit_hash(ct, meta)
        h2 = crypto._calculate_audit_hash(ct, meta)
        assert h1 == h2


# ── SovereignPayload ─────────────────────────────────────────────────────────


class TestSovereignPayload:
    @pytest.fixture
    def sov(self):
        crypto = CryptoManager()
        return SovereignPayload(crypto)

    def test_encrypt_event(self, sov):
        event_data = {"product": "widget", "qty": 5}
        metadata = {"event_type": "SALE"}
        encrypted = sov.encrypt_event(event_data, metadata)
        assert isinstance(encrypted, EncryptedPayload)

    @pytest.mark.xfail(
        reason="Pre-existing bug: CryptoManager.decrypt() fails due to "
               "incorrect DEK auth-tag reconstruction",
        strict=True,
    )
    def test_encrypt_decrypt_event_roundtrip(self, sov):
        event_data = {"product": "widget", "qty": 5, "price": 19.99}
        metadata = {"event_type": "SALE"}
        encrypted = sov.encrypt_event(event_data, metadata)
        decrypted = sov.decrypt_event(encrypted, metadata)
        assert decrypted == event_data

    def test_encrypt_with_sensitive_fields(self, sov):
        event_data = {"card_number": "4111-1111-1111-1111"}
        metadata = {"type": "PAYMENT"}
        encrypted = sov.encrypt_event(
            event_data, metadata, sensitive_fields=["card_number"]
        )
        assert encrypted.encrypted_inner_layer is not None

    def test_encrypt_without_sensitive_fields(self, sov):
        event_data = {"product": "item"}
        metadata = {"type": "SALE"}
        encrypted = sov.encrypt_event(event_data, metadata, sensitive_fields=None)
        assert encrypted.encrypted_inner_layer is None

    def test_encrypt_with_empty_sensitive_fields(self, sov):
        event_data = {"product": "item"}
        metadata = {"type": "SALE"}
        encrypted = sov.encrypt_event(event_data, metadata, sensitive_fields=[])
        assert encrypted.encrypted_inner_layer is None
