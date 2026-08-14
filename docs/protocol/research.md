# Protocol Research

## Observed Behavior

- Apple devices broadcast BLE advertisements with manufacturer ID 0x004C
- Advertisements may include service data that identifies AirDrop-capable devices
- Device names are commonly included in advertisement local names

## Inferred Behavior

- AirDrop likely uses a record-based protocol within BLE service data
- Record identifiers may be derived from service data contents
- Device capabilities may be encoded in advertisement metadata

## Experimental Behavior

- The current parser treats any Apple manufacturer advertisement with service data as potentially AirDrop-capable
- Exact record semantics are not yet verified

## Unknown Behavior

- Exact AirDrop record format
- Capability encoding
- Connection handshake protocol
- Transfer protocol
