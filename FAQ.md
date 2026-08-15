# Frequently Asked Questions

## What is ReverseDrop?

ReverseDrop is a genuine reverse-engineered implementation of Apple's AirDrop protocol. It lets Windows, Mac, and Linux devices share files using the same protocols that Apple AirDrop uses.

## Is ReverseDrop compatible with Apple AirDrop?

Yes. ReverseDrop uses the exact same BLE advertisements, mDNS service discovery, TLS transport, HTTP API, and archive formats as Apple AirDrop. It can discover and communicate with macOS and iOS devices running AirDrop.

## Do I need an Apple ID to use ReverseDrop?

No. ReverseDrop does not require an Apple ID, iCloud account, or any other account. It works completely standalone.

## Does ReverseDrop work offline?

Yes. All file transfers happen directly between devices using Bluetooth and local Wi-Fi. No internet connection is required.

## Is ReverseDrop secure?

ReverseDrop uses TLS 1.2/1.3 for all transfers, just like real AirDrop. However, it does not implement Apple's full PKI certificate validation, so it should only be used on trusted networks.

## Why did you reverse-engineer AirDrop?

Apple's AirDrop protocol is proprietary and only works between Apple devices. By reverse-engineering it, we can make it work across all platforms. The protocol details are publicly available through academic research.

## Can ReverseDrop send files to iPhones?

Yes, if the iPhone has AirDrop enabled and is discoverable. ReverseDrop broadcasts the same BLE advertisements and mDNS records that AirDrop uses.

## Can iPhones send files to ReverseDrop?

iPhones can discover ReverseDrop devices via mDNS if they are on the same network. However, iPhones may not see ReverseDrop in their BLE scan results because ReverseDrop does not have an Apple ID to hash.

## What devices can I share files with?

- Other ReverseDrop instances (Windows, Mac, Linux)
- macOS devices with AirDrop enabled
- iOS devices with AirDrop enabled

## How big of files can I transfer?

There is no hard limit, but large files (100MB+) may transfer slowly, especially on Windows. The transfer speed depends on your Wi-Fi Direct or local network speed.

## Does ReverseDrop work over the internet?

No. ReverseDrop only works over local networks (Wi-Fi Direct or mDNS multicast). It does not route traffic over the internet.

## What happens if the transfer fails?

If a transfer fails, the file is not saved on the receiving device. You can simply retry the transfer.

## Can I use ReverseDrop on my phone?

Currently, ReverseDrop only supports desktop platforms (Windows, Mac, Linux). Mobile support may be added in the future.

## Is ReverseDrop open source?

Yes. ReverseDrop is released under the GNU General Public License v3.0. You can inspect, modify, and redistribute the code.

## Who maintains ReverseDrop?

ReverseDrop is maintained by the open-source community. Contributions are welcome!

## I found a bug. How do I report it?

Please open an issue on GitHub with:
- Your operating system and version
- The version of ReverseDrop you are using
- A description of what happened
- Any error messages you saw
