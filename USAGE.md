# Usage Guide

This guide explains how to use ReverseDrop for everyday file sharing.

---

## Starting ReverseDrop

### Windows

Click **Start**, type `ReverseDrop`, and press Enter. Or find it in your Start Menu.

### Mac

Open **Launchpad** or go to **Applications** and double-click **ReverseDrop**.

### Linux

Open your applications menu and find **ReverseDrop**, or run `reversedrop-gui` from a terminal.

---

## Sharing a File

1. **Make sure Bluetooth is turned on** on both devices.
2. **Open ReverseDrop** on both devices.
3. **On the sending device:**
   - You will see a list of nearby devices appear automatically.
   - Click the device you want to send to.
   - A file picker will open. Select the file you want to share.
   - Click **Send**.
4. **On the receiving device:**
   - You will see a pop-up asking if you want to accept the file.
   - Click **Accept**.
   - Choose where to save the file.
   - The file will be transferred.

---

## Sharing a Link

1. **Copy the link** you want to share (Ctrl+C / Cmd+C).
2. **Open ReverseDrop** on both devices.
3. **On the sending device:**
   - Click the device you want to send to.
   - Select **Share Link** (or paste the link in the text field).
   - Click **Send**.
4. **On the receiving device:**
   - Click **Accept**.
   - The link will open in your default browser.

---

## Receiving Files

When someone sends you a file:

1. You will see a notification or pop-up in ReverseDrop.
2. Click **Accept** to receive the file.
3. Choose a save location if prompted.
4. Wait for the transfer to complete.

You can also click **Decline** to reject the file.

---

## Device List

The main window shows nearby devices that are also running ReverseDrop.

- **Green dot**: device is available and ready.
- **Yellow dot**: device is busy or connecting.
- **Red dot**: device is unavailable or out of range.

If you do not see a device you expect:
- Make sure both devices have Bluetooth turned on.
- Make sure both devices are running ReverseDrop.
- Move the devices closer together (within about 10 meters / 30 feet).

---

## Settings

Click the **Settings** or **Preferences** button in the app to adjust:

- **Device name**: what other people see when you appear in their list
- **Auto-accept**: automatically accept files from trusted devices
- **Save location**: default folder for received files
- **Notifications**: show or hide transfer notifications

---

## Offline Use

ReverseDrop works completely offline. No internet connection is required. All transfers happen directly between devices using Bluetooth and local WiFi.

---

## Using the Command Line (Advanced)

If you prefer the command line, ReverseDrop also includes a CLI tool.

```bash
# Scan for nearby devices
./reversedrop scan

# Send a file to a specific device
./reversedrop send --target AA:BB:CC:DD:EE:FF --file photo.jpg

# Receive files (listens for incoming transfers)
./reversedrop receive
```

See `./reversedrop --help` for all available options.

---

## Tips

- Keep your device awake and the app open while transferring large files.
- For best performance, stay within 5 meters (15 feet) of the other device.
- If a transfer is slow, try moving closer or removing obstacles between devices.
- You can share multiple files one at a time, or select a folder to share all files inside it.
