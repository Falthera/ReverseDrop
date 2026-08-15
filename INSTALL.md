# Installation

Choose your platform below for step-by-step installation instructions.

---

## Windows

### Requirements

- Windows 10 or Windows 11
- Bluetooth adapter (built-in or USB)

### Install from MSI

1. Download the latest `reversedrop-windows-amd64.msi` from the [Releases page](https://github.com/Falthera/ReverseDrop/releases).
2. Double-click the `.msi` file to start the installer.
3. Follow the prompts:
   - Choose the installation folder (default is `C:\Program Files\ReverseDrop`)
   - Click **Install**
4. When finished, click **Finish**.
5. Launch ReverseDrop from the Start Menu or desktop shortcut.

### Uninstall

Open **Settings > Apps > Installed apps**, find ReverseDrop, and click **Uninstall**.

---

## macOS

### Requirements

- macOS 12 (Monterey) or later
- Bluetooth adapter (built-in on all modern Macs)

### Install from DMG

1. Download the latest `ReverseDrop.dmg` from the [Releases page](https://github.com/Falthera/ReverseDrop/releases).
2. Double-click the `.dmg` file to open it.
3. Drag the **ReverseDrop** app icon into the **Applications** folder.
4. Eject the disk image.
5. Launch ReverseDrop from Launchpad or the Applications folder.

### Install from PKG

1. Download the latest `ReverseDrop.pkg` from the [Releases page](https://github.com/Falthera/ReverseDrop/releases).
2. Double-click the `.pkg` file to start the installer.
3. Follow the prompts to install to `/Applications`.
4. Launch ReverseDrop from Launchpad or the Applications folder.

### First Run & Permissions

On macOS, the first time you run ReverseDrop, you may need to grant Bluetooth permissions:

1. Open **System Settings > Privacy & Security > Bluetooth**
2. Find ReverseDrop in the list and make sure it is enabled
3. If prompted, click **OK** or **Allow**

### Uninstall

Delete the `ReverseDrop.app` from the Applications folder. If you used the PKG installer, you can also run the PKG file again and select **Uninstall**.

---

## Linux

### Requirements

- A modern Linux distribution (Ubuntu 20.04+, Fedora 35+, Debian 11+, etc.)
- Bluetooth adapter
- For GUI: a desktop environment with X11 or Wayland

### Install from .deb (Debian / Ubuntu / Mint)

1. Download the latest `reversedrop_*.deb` from the [Releases page](https://github.com/Falthera/ReverseDrop/releases).
2. Install using your package manager:

   ```bash
   sudo dpkg -i reversedrop_*.deb
   sudo apt-get install -f
   ```

3. Launch ReverseDrop from your applications menu.

### Install from .rpm (Fedora / RHEL / openSUSE)

1. Download the latest `reversedrop-*.rpm` from the [Releases page](https://github.com/Falthera/ReverseDrop/releases).
2. Install using your package manager:

   ```bash
   sudo dnf install reversedrop-*.rpm
   # or on older systems:
   # sudo rpm -i reversedrop-*.rpm
   ```

3. Launch ReverseDrop from your applications menu.

### Uninstall

```bash
# Debian/Ubuntu
sudo apt-get remove reversedrop

# Fedora/RHEL
sudo dnf remove reversedrop
```

---

## FreeBSD / OpenBSD / NetBSD

These platforms are community-tested. You will need to build from source.

### Requirements

- Go 1.22 or later
- Bluetooth stack (e.g., BlueZ on Linux-compatible layers)
- X11 or Wayland libraries for GUI

### Build from Source

```bash
git clone https://github.com/Falthera/ReverseDrop.git
cd ReverseDrop
make build
```

See [CONTRIBUTING.md](CONTRIBUTING.md) for detailed build instructions.

---

## Building from Source (All Platforms)

If you prefer to build from source, or if your platform is not listed above:

```bash
git clone https://github.com/Falthera/ReverseDrop.git
cd ReverseDrop
make build
```

This produces `reversedrop` (CLI) and `reversedrop-gui` (GUI) in the project directory.

See [docs/development/building.md](docs/development/building.md) for details.
