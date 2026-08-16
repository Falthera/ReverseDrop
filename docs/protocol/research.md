# Protocol Research

## Reverse-Engineering Sources

This implementation is based on the following public reverse-engineering work:

- Stute et al., *"A Billion Open Interfaces for Eve and Mallory"*, USENIX Security '19
- Heinrich et al., *"Discontinued Privacy: Personal Data Leaks in Apple BLE Continuity Protocols"*, PoPETs 2020
- Ebrahim & Tippenhauer, *"Protocol Prying: AirDrop & Quick Share"*, arXiv:2606.26967, 2026

## Observed Behavior

- Apple devices broadcast BLE advertisements with manufacturer ID `0x004C`
- AirDrop sub-type is `0x05` within the manufacturer data
- Advertisements contain truncated SHA-256 hashes (first 2 bytes = 16 bits) of contact identifiers
- Device names are NOT in BLE advertisements; they appear in mDNS TXT records

## Inferred Behavior

- AirDrop uses a two-phase discovery: BLE wake-up → mDNS over AWDL/Wi-Fi Direct
- Service type: `_airdrop._tcp.local.`
- TCP port: 8770
- All message bodies use Apple Binary Property Lists (bplist00)
- File transfers use CPIO newc archives (`070701` magic) compressed with DVZip or gzip

## Implemented Behavior

ReverseDrop implements the complete AirDrop protocol stack:

1. **BLE Advertisements**: Generates and parses real AirDrop BLE advertisements with Apple company ID `0x004C` and sub-type `0x05`
2. **mDNS Discovery**: Publishes and browses `_airdrop._tcp.local.` with AirDrop TXT records
3. **TLS Transport**: Self-signed certificates, TLS 1.2/1.3, no hostname verification
4. **HTTP API**: `/Discover`, `/Ask`, `/Upload` endpoints with bplist bodies
5. **Archive Format**: CPIO newc + DVZip/gzip compression matching AirDrop's exact format

## Protocol Constants

| Item | Value |
|------|-------|
| BLE Apple Company ID | `0x004C` |
| BLE AirDrop Sub-Type | `0x05` |
| BLE PDU Type | `ADV_NONCONN_IND` |
| Hash truncation | First **16 bits** of SHA-256 |
| mDNS Service Type | `_airdrop._tcp.local.` |
| TCP Port | **8770** |
| TLS Version | TLS 1.2 / 1.3 |
| HTTP Endpoints | `/Discover`, `/Ask`, `/Upload`, `/Error` |
| Upload Format | DVZip (`application/x-dvzip`) or gzip + CPIO newc |
| Archive Magic | `070701` (CPIO newc) |
| Archive Terminator | `TRAILER!!!` |

## DVZip Compression

ReverseDrop uses `gzip.BestCompression` with 256KB chunks for DVZip archives. This provides better compression ratios at the cost of increased CPU usage and slightly slower transfer times. For faster transfers with lower compression, consider adjusting the compression level and chunk size in `internal/transfer/transfer.go`.

## Open Questions

- Exact AirDrop record format for all TXT record keys
- Capability encoding details
- Full DVZip adaptive compression behavior
- Apple ID Validation Record (VR) certificate chain format
- Auto-accept behavior for trusted contacts
