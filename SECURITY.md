# Security Policy

## Supported Versions

| Version | Supported |
|---|---|
| 1.x | Yes |

## Reporting a Vulnerability

Please report security vulnerabilities privately via GitHub Security Advisories.

## Release Signing

GitHub Releases are signed with GPG to ensure authenticity. The public key is available in the repository.

### Verify Release Signatures

1. Import the public key:
   ```bash
   gpg --keyserver keyserver.ubuntu.com --recv-keys <KEY_ID>
   ```

2. Verify a release asset:
   ```bash
   gpg --verify reversedrop-linux-amd64.tar.gz.sig reversedrop-linux-amd64.tar.gz
   ```

### Checksums

Each release includes SHA256 checksums for all assets. Verify with:

```bash
sha256sum -c SHA256SUMS.txt
```

## Security Model

ReverseDrop implements the same security model as Apple AirDrop:

### Transport Security
- All data transfers use **TLS 1.2 or TLS 1.3**
- Certificates are self-signed — no PKI or CA required
- No hostname verification is performed
- TLS provides encryption and integrity, not peer authentication

### Authentication
- Contact verification happens at the application layer via bplist messages
- The `/Discover` endpoint exchanges sender identity information
- Contact matching is based on truncated SHA-256 hashes of contact identifiers
- **No Apple ID or iCloud account is required**

### Privacy
- No telemetry, no tracking, no analytics
- Files never pass through any server
- Open source — all code can be audited

## Known Limitations

- Self-signed certificates are not pinned. Any device can impersonate another.
- Contact hash matching is simplified compared to Apple's full PKI implementation.
- No protection against man-in-the-middle attacks on untrusted networks.
