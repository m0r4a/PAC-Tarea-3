#!/usr/bin/env python3
import secrets

def generate_aes_key():
    """Genera una clave AES-128 aleatoria (16 bytes = 32 caracteres hex)"""
    key_bytes = secrets.token_bytes(16)
    key_hex = key_bytes.hex().upper()
    return key_hex

if __name__ == "__main__":
    key = generate_aes_key()
    print("=" * 60)
    print("  AES-128 Key Generator")
    print("=" * 60)
    print(f"\nGenerated Key: {key}")
    print(f"\nLength: {len(key)} characters (16 bytes)")
    print("\nUso del servidor:")
    print(f"  python3 server.py 4444 {key}")
    print("\nUso del cliente:")
    print(f"  .\\client.exe 192.168.100.1 4444 {key} --nopersist")
    print("=" * 60)
