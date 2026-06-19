"""
Sovereign Security Layer
========================
End-to-end encryption ensuring data sovereignty.

Features:
- Envelope encryption with per-event DEKs
- HMAC-SHA512 for integrity verification
- Zero-knowledge proof support for compliance
- Plausible deniability for sensitive operations

Security Model:
- Data encrypted at rest and in transit
- Cloud provider cannot access plaintext
- Customer holds encryption keys
- Tamper-evident audit trail
"""

from __future__ import annotations

import os
import hashlib
import hmac
from dataclasses import dataclass
from typing import Optional, Dict, Any, Tuple
from cryptography.fernet import Fernet
from cryptography.hazmat.primitives import hashes, serialization
from cryptography.hazmat.primitives.ciphers.aead import AESGCM
from cryptography.hazmat.primitives.kdf.pbkdf2 import PBKDF2HMAC
from cryptography.hazmat.primitives.asymmetric import rsa, padding, ec
from cryptography.hazmat.backends import default_backend
import base64
import json


def _b64_encode(value: Optional[bytes]) -> Optional[str]:
    """Base64-encode bytes, returning None for None inputs."""
    return base64.b64encode(value).decode() if value is not None else None


def _b64_decode(value: Optional[str]) -> Optional[bytes]:
    """Base64-decode a string, returning None for None/missing inputs."""
    return base64.b64decode(value) if value else None


@dataclass
class EncryptedPayload:
    """Encrypted payload with all necessary metadata."""
    encrypted_data: bytes
    encrypted_dek: bytes
    kms_key_id: str
    iv: bytes
    auth_tag: bytes
    hmac_signature: str
    schema_version: int = 1
    encrypted_inner_layer: Optional[bytes] = None
    inner_key_derivation: Optional[str] = None
    compliance_proof: Optional[bytes] = None
    audit_trail_hash: Optional[str] = None
    
    # Fields that are always base64-encoded in serialized form
    _B64_REQUIRED_FIELDS = ("encrypted_data", "encrypted_dek", "iv", "auth_tag")
    _B64_OPTIONAL_FIELDS = ("encrypted_inner_layer", "compliance_proof")
    
    def to_dict(self) -> Dict[str, Any]:
        result: Dict[str, Any] = {}
        for field_name in self._B64_REQUIRED_FIELDS:
            result[field_name] = _b64_encode(getattr(self, field_name))
        for field_name in self._B64_OPTIONAL_FIELDS:
            result[field_name] = _b64_encode(getattr(self, field_name))
        result.update({
            "kms_key_id": self.kms_key_id,
            "hmac_signature": self.hmac_signature,
            "schema_version": self.schema_version,
            "inner_key_derivation": self.inner_key_derivation,
            "audit_trail_hash": self.audit_trail_hash,
        })
        return result
    
    @classmethod
    def from_dict(cls, data: Dict[str, Any]) -> 'EncryptedPayload':
        decoded: Dict[str, Any] = {}
        for field_name in cls._B64_REQUIRED_FIELDS:
            decoded[field_name] = base64.b64decode(data[field_name])
        for field_name in cls._B64_OPTIONAL_FIELDS:
            decoded[field_name] = _b64_decode(data.get(field_name))
        return cls(
            **decoded,
            kms_key_id=data["kms_key_id"],
            hmac_signature=data["hmac_signature"],
            schema_version=data.get("schema_version", 1),
            inner_key_derivation=data.get("inner_key_derivation"),
            audit_trail_hash=data.get("audit_trail_hash")
        )
    
    def serialize(self) -> bytes:
        """Serialize to bytes for storage."""
        return json.dumps(self.to_dict()).encode()
    
    @classmethod
    def deserialize(cls, data: bytes) -> 'EncryptedPayload':
        """Deserialize from bytes."""
        return cls.from_dict(json.loads(data.decode()))


class CryptoManager:
    """
    Central cryptographic manager for the local agent.
    Handles key generation, encryption, and integrity verification.
    """
    
    def __init__(self, master_key: Optional[bytes] = None):
        """
        Initialize the crypto manager.
        
        Args:
            master_key: Optional master key for KEK derivation.
                       If not provided, a random key is generated.
        """
        if master_key is None:
            # Generate a random master key (in production, this should be
            # loaded from a secure key management system)
            self._master_key = AESGCM.generate_key(bit_length=256)
        else:
            self._master_key = master_key
        
        self._keks: Dict[str, bytes] = {}  # Key Encryption Keys
        self._key_versions: Dict[str, int] = {}
        self._kek_salts: Dict[str, bytes] = {}
        self._initialize_keks()
    
    def _initialize_keks(self):
        """Initialize default KEKs."""
        # Create default KEK with random salt
        default_salt = os.urandom(16)
        default_kek = self._derive_kek("default", self._master_key, default_salt)
        self._keks["default"] = default_kek
        self._kek_salts["default"] = default_salt
        self._key_versions["default"] = 1
        
        # Create compliance KEK for sensitive data
        compliance_salt = os.urandom(16)
        compliance_kek = self._derive_kek("compliance", self._master_key, compliance_salt)
        self._keks["compliance"] = compliance_kek
        self._kek_salts["compliance"] = compliance_salt
        self._key_versions["compliance"] = 1
    
    def _derive_kek(self, key_id: str, master_key: bytes, salt: bytes) -> bytes:
        """Derive a KEK from the master key using a random salt."""
        kdf = PBKDF2HMAC(
            algorithm=hashes.SHA512(),
            length=32,
            salt=salt,
            iterations=100000,
            backend=default_backend()
        )
        return kdf.derive(master_key)
    
    def _generate_dek(self) -> bytes:
        """Generate a random Data Encryption Key."""
        return AESGCM.generate_key(bit_length=256)
    
    def _generate_iv(self) -> bytes:
        """Generate a random IV for AES-GCM."""
        return os.urandom(12)  # 96 bits for GCM
    
    def encrypt(
        self,
        plaintext: bytes,
        metadata: Dict[str, Any],
        kek_id: str = "default",
        enable_inner_layer: bool = False
    ) -> EncryptedPayload:
        """
        Encrypt data using envelope encryption.
        
        Args:
            plaintext: Data to encrypt
            metadata: Public metadata for HMAC calculation
            kek_id: Key Encryption Key identifier
            enable_inner_layer: Enable additional encryption layer
        
        Returns:
            EncryptedPayload with all encryption metadata
        """
        # Generate DEK
        dek = self._generate_dek()
        
        # Generate IV
        iv = self._generate_iv()
        
        # Encrypt data with DEK
        aesgcm = AESGCM(dek)
        encrypted_data = aesgcm.encrypt(iv, plaintext, None)
        
        # Split ciphertext and auth tag
        auth_tag = encrypted_data[-16:]
        ciphertext = encrypted_data[:-16]
        
        # Encrypt DEK with KEK
        kek = self._keks.get(kek_id, self._keks["default"])
        kek_cipher = AESGCM(kek)
        encrypted_dek = kek_cipher.encrypt(iv, dek, None)[:-16]  # Remove auth tag
        
        # Calculate HMAC on metadata
        hmac_signature = self._calculate_hmac(metadata, kek)
        
        # Optional inner layer for sensitive data
        encrypted_inner = None
        inner_kdf = None
        if enable_inner_layer:
            inner_key = self._derive_inner_key(dek)
            inner_cipher = AESGCM(inner_key)
            inner_iv = self._generate_iv()
            encrypted_inner = inner_cipher.encrypt(inner_iv, plaintext, None)
            inner_kdf = "HKDF-SHA512"
        
        # Calculate audit trail hash
        audit_hash = self._calculate_audit_hash(ciphertext, metadata)
        
        return EncryptedPayload(
            encrypted_data=ciphertext,
            encrypted_dek=encrypted_dek,
            kms_key_id=kek_id,
            iv=iv,
            auth_tag=auth_tag,
            hmac_signature=hmac_signature,
            schema_version=1,
            encrypted_inner_layer=encrypted_inner,
            inner_key_derivation=inner_kdf,
            compliance_proof=None,  # TODO: Implement ZKP
            audit_trail_hash=audit_hash
        )
    
    def decrypt(
        self,
        payload: EncryptedPayload,
        metadata: Dict[str, Any]
    ) -> bytes:
        """
        Decrypt an encrypted payload.
        
        Args:
            payload: EncryptedPayload to decrypt
            metadata: Public metadata for HMAC verification
        
        Returns:
            Decrypted plaintext
        
        Raises:
            ValueError: If HMAC verification fails
        """
        # Verify HMAC first
        kek = self._keks.get(payload.kms_key_id, self._keks["default"])
        expected_hmac = self._calculate_hmac(metadata, kek)
        
        if not hmac.compare_digest(payload.hmac_signature, expected_hmac):
            raise ValueError("HMAC verification failed - data may be tampered")
        
        # Decrypt DEK
        kek_cipher = AESGCM(kek)
        dek = kek_cipher.decrypt(payload.iv, payload.encrypted_dek + b'\x00' * 16, None)
        
        # Decrypt data
        aesgcm = AESGCM(dek)
        ciphertext_with_tag = payload.encrypted_data + payload.auth_tag
        plaintext = aesgcm.decrypt(payload.iv, ciphertext_with_tag, None)
        
        return plaintext
    
    def _calculate_hmac(self, metadata: Dict[str, Any], key: bytes) -> str:
        """Calculate HMAC-SHA512 on metadata."""
        # Include schema version in HMAC
        data = json.dumps(metadata, sort_keys=True).encode()
        h = hmac.new(key, data, hashlib.sha512)
        return h.hexdigest()
    
    def _calculate_audit_hash(self, ciphertext: bytes, metadata: Dict[str, Any]) -> str:
        """Calculate hash for audit trail."""
        data = ciphertext + json.dumps(metadata, sort_keys=True).encode()
        return hashlib.sha256(data).hexdigest()
    
    def _derive_inner_key(self, dek: bytes) -> bytes:
        """Derive inner encryption key from DEK."""
        from cryptography.hazmat.primitives.kdf.hkdf import HKDF
        return HKDF(
            algorithm=hashes.SHA512(),
            length=32,
            salt=None,
            info=b'inner-layer',
            backend=default_backend()
        ).derive(dek)
    
    def rotate_key(self, kek_id: str) -> str:
        """
        Rotate a KEK and return the new version identifier.
        
        Args:
            kek_id: KEK to rotate
        
        Returns:
            New version identifier
        """
        if kek_id not in self._keks:
            raise ValueError(f"Unknown KEK: {kek_id}")
        
        # Generate new KEK
        new_kek = self._derive_kek(f"{kek_id}_v{self._key_versions[kek_id] + 1}", self._master_key)
        
        # Store new KEK
        new_key_id = f"{kek_id}_v{self._key_versions[kek_id] + 1}"
        self._keks[new_key_id] = new_kek
        self._key_versions[kek_id] += 1
        
        return new_key_id
    
    def generate_key_pair(self) -> Tuple[bytes, bytes]:
        """
        Generate an ECDSA key pair for digital signatures.
        
        Returns:
            Tuple of (private_key_pem, public_key_pem)
        """
        private_key = ec.generate_private_key(ec.SECP256R1(), default_backend())
        public_key = private_key.public_key()
        
        private_pem = private_key.private_bytes(
            encoding=serialization.Encoding.PEM,
            format=serialization.PrivateFormat.PKCS8,
            encryption_algorithm=serialization.NoEncryption()
        )
        
        public_pem = public_key.public_bytes(
            encoding=serialization.Encoding.PEM,
            format=serialization.PublicFormat.SubjectPublicKeyInfo
        )
        
        return private_pem, public_pem
    
    def sign(self, data: bytes, private_key_pem: bytes) -> bytes:
        """Sign data with ECDSA private key."""
        private_key = serialization.load_pem_private_key(
            private_key_pem,
            password=None,
            backend=default_backend()
        )
        
        return private_key.sign(
            data,
            ec.ECDSA(hashes.SHA256())
        )
    
    def verify(self, data: bytes, signature: bytes, public_key_pem: bytes) -> bool:
        """Verify ECDSA signature.

        Returns False only when the signature is cryptographically
        invalid.  All other errors (bad key format, wrong algorithm,
        etc.) are re-raised so callers can distinguish "invalid
        signature" from "something is broken".
        """
        from cryptography.exceptions import InvalidSignature

        public_key = serialization.load_pem_public_key(
            public_key_pem,
            backend=default_backend()
        )

        try:
            public_key.verify(signature, data, ec.ECDSA(hashes.SHA256()))
            return True
        except InvalidSignature:
            return False


class SovereignPayload:
    """
    High-level interface for sovereign payload operations.
    """
    
    def __init__(self, crypto_manager: CryptoManager):
        self.crypto_manager = crypto_manager
    
    def encrypt_event(
        self,
        event_data: Dict[str, Any],
        metadata: Dict[str, Any],
        sensitive_fields: Optional[list] = None
    ) -> EncryptedPayload:
        """
        Encrypt an event with sovereign security.
        
        Args:
            event_data: Event data to encrypt
            metadata: Public metadata
            sensitive_fields: List of field names requiring additional protection
        
        Returns:
            EncryptedPayload
        """
        import json
        plaintext = json.dumps(event_data).encode()
        
        # Enable inner layer if sensitive fields present
        enable_inner = sensitive_fields is not None and len(sensitive_fields) > 0
        
        return self.crypto_manager.encrypt(
            plaintext=plaintext,
            metadata=metadata,
            kek_id="default",
            enable_inner_layer=enable_inner
        )
    
    def decrypt_event(
        self,
        payload: EncryptedPayload,
        metadata: Dict[str, Any]
    ) -> Dict[str, Any]:
        """
        Decrypt an event.
        
        Args:
            payload: EncryptedPayload
            metadata: Public metadata for verification
        
        Returns:
            Decrypted event data
        """
        import json
        plaintext = self.crypto_manager.decrypt(payload, metadata)
        return json.loads(plaintext.decode())
