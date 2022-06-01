# Attested TLS (aTLS)

In a confidential computing (CC) environment, attested TLS (aTLS) can be used to establish secure connections between two parties utilizing the remote attestation features of the CC components.

aTLs modifies the TLS handshake by embedding an attestation statement into the TLS certificate.
Instead of relying on a Certificate Authority, aTLS uses this attestation statement to establish trust in the certificate.

The protocol can be used by clients to verify a server certificate, by a server to verify a client certificate, or for mutual verification (mutual aTLS).

## Client side verification

1. The client sends a ClientHello message, setting ServerName to a random nonce.

2. The server generates an attestation statement using the clients nonce and its CC capabilities.
    * The attestation is embedded in the server certificate using x509 certificate extensions with an object identifier (OID) to identify the CC attestation type. See [OID](../oid/oid.go) for implementation details.

3. The client verifies the attestation statement.

4. If successful the client can trust the server to be running the expected configuration, and finish the TLS handshake.

```mermaid
sequenceDiagram
    participant Client
    participant Server
    Client->>Server: ClientHello(nonce)
    Server->>Client: ServerCertificate(AttestationStatement), ServerHelloDone
    Note over Client: Verify Attestation
    Client->>Server: ClientKeyExchange
    Client->>Server: ChangeCipherSpec, Finished
    Server->>Client: 
```

## Server side verification

1. The client sends a ClientHello message

2. The server sends back a certificate with a random nonce. The nonce is embedded using x509 certificate extensions with the OID `1.3.9900.0.1`.

3. The client does not verify the servers certificate, but uses the embedded nonce to generate an attestation based on its CC capabilities.
    * The attestation is embedded in the client certificate using x509 certificate extensions with an OID to identify the CC attestation type.

4. The server verifies the client's attestation statement.

5. If successful the server can trust the client to be running the expected configuration, and finish the TLS handshake.

```mermaid
sequenceDiagram
    participant Client
    participant Server
    Client->>Server: ClientHello
    Server->>Client: ServerCertificate(nonce), ServerHelloDone
    Client->>Server: ClientKeyExchange, ClientCertificate(AttestationStatement)
    Client->>Server: ChangeCipherSpec, Finished
    Note over Server: Verify Attestation
    Server->>Client: ChangeCipherSpec, Finished
```

## Mutual aTLS

1. The client sends a ClientHello message, setting ServerName to a random nonce.

2. The server generates an attestation statement using the clients nonce and its CC capabilities.
    * The attestation is embedded in the server certificate using x509 certificate extensions with an OID to identify the attestation type.
    * A nonce is embedded using x509 certificate extensions with the OID `1.3.9900.0.1`.

3. The client verifies the attestation statement.

4. The client uses the nonce embedded in the server's certificate to generate an attestation based on its CC capabilities.
    * The attestation is embedded in the client certificate using x509 certificate extensions with an OID to identify the CC attestation type.

5. The server verifies the client's attestation statement.

6. If all verifications were successful, mutual trust in each others configuration is established, and the TLS handshake can be finished.

```mermaid
sequenceDiagram
    participant Client
    participant Server
    Client->>Server: ClientHello(nonce)
    Server->>Client: ServerCertificate(AttestationStatement, nonce), ServerHelloDone
    Note over Client: Verify Attestation
    Client->>Server: ClientKeyExchange, ClientCertificate(AttestationStatement)
    Client->>Server: ChangeCipherSpec, Finished
    Note over Server: Verify Attestation
    Server->>Client: ChangeCipherSpec, Finished
```
