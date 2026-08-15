# Troubleshooting

This guide helps you solve common problems with ReverseDrop.

---

## ReverseDrop won't start

### Windows
- Make sure you have Windows 10 or later.
- Try running the installer again (right-click the `.msi` file and select **Repair**).
- Check that you have enough disk space.
- Restart your computer and try again.

### Mac
- Make sure you have macOS 12 (Monterey) or later.
- If you see a security warning, go to **System Settings > Privacy & Security** and click **Open Anyway**.
- Make sure you have granted Bluetooth permissions (see below).

### Linux
- Make sure you have a supported desktop environment (GNOME, KDE, XFCE, etc.).
- Try running `reversedrop-gui` from a terminal to see error messages.
- Make sure you have the required libraries installed (see below).

---

## Bluetooth issues

### "Bluetooth not available" or "Bluetooth disabled"

1. Make sure Bluetooth is turned on:
   - **Windows:** Settings > Bluetooth & devices > Bluetooth
   - **Mac:** System Settings > Bluetooth
   - **Linux:** Settings > Bluetooth, or run `bluetoothctl` and check status

2. Make sure your device has a Bluetooth adapter. Most modern laptops and phones have built-in Bluetooth. Desktop PCs may need a USB Bluetooth adapter.

### "Permission denied" or "Bluetooth access denied"

#### macOS
1. Open **System Settings > Privacy & Security > Bluetooth**
2. Find ReverseDrop in the list
3. Make sure the toggle is turned **ON**
4. Restart ReverseDrop

#### Linux
You may need to grant your user access to Bluetooth:

```bash
sudo usermod -aG bluetooth $USER
```

Log out and back in for the change to take effect.

---

## Can't find nearby devices

1. **Make sure both devices have Bluetooth turned on.**
2. **Make sure both devices are running ReverseDrop.**
3. **Move devices closer together.** Bluetooth works best within 5-10 meters (15-30 feet).
4. **Remove obstacles.** Walls, metal objects, and interference from other devices can reduce range.
5. **Restart Bluetooth** on both devices.
6. **Restart ReverseDrop** on both devices.

---

## File transfer is slow or stuck

1. **Move closer together.** Signal strength drops with distance.
2. **Remove interference.** Other Bluetooth devices, microwaves, and WiFi routers can cause interference.
3. **Keep devices awake.** On laptops, make sure the screen does not turn off during transfer.
4. **Try a smaller file first** to confirm the connection works.
5. **Restart the transfer** if it gets stuck.

---

## "Permission denied" when saving files

1. Make sure you have write access to the folder you are saving to.
2. Try saving to your home folder or Desktop instead.
3. On Linux, make sure you are not trying to save to a system folder without `sudo`.

---

## ReverseDrop crashes or freezes

1. **Restart the app.**
2. **Restart your computer.**
3. **Update to the latest version** from the Releases page.
4. **Check for system updates** (Windows Update, macOS Software Update, or your Linux package manager).

---

## Linux-specific issues

### "cannot open display" or GUI won't start

Make sure you have a graphical desktop environment running. ReverseDrop requires X11 or Wayland.

If you are using SSH, make sure X11 forwarding is enabled, or run the app locally.

### Missing libraries

On some Linux distributions, you may need to install additional packages:

**Ubuntu / Debian:**
```bash
sudo apt-get install libbluetooth-dev libwayland-dev
```

**Fedora:**
```bash
sudo dnf install bluez-libs-devel wayland-devel
```

**Arch:**
```bash
sudo pacman -S bluez-libs wayland
```

---

## Still need help?

If you are still having trouble:

1. Check the [FAQ.md](FAQ.md) for more answers.
2. Search [existing issues](https://github.com/Falthera/ReverseDrop/issues) to see if others have had the same problem.
3. Open a [new issue](https://github.com/Falthera/ReverseDrop/issues/new) and include:
   - Your operating system and version
   - The version of ReverseDrop you are using
   - A description of what happened
   - Any error messages you saw
