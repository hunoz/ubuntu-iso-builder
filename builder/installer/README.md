# USB Installer for mkosi Image

This creates a bootable USB that automatically installs your mkosi image to a target disk.

## What This Does

1. Boot from USB
2. Auto-login to installer
3. Select target disk interactively
4. Write image to disk
5. Shutdown when complete

## Quick Start

### 1. Build Your Main Image First

```bash
cd ../mkosi-ubuntu-project
sudo mkosi build
```

### 2. Build the Installer

```bash
./build-installer.sh
```

This will:
- Copy your main image into the installer
- Build a bootable installer USB image
- Output: `mkosi.output/ubuntu-installer_1.0.raw`

### 3. Write to USB

```bash
# Find your USB device
lsblk

# Write installer to USB (ERASES USB!)
sudo dd if=mkosi.output/ubuntu-installer_1.0.raw of=/dev/sdX bs=4M status=progress
sync
```

### 4. Use the Installer

1. Boot from USB
2. Installer starts automatically
3. Select target disk
4. Type "yes" to confirm
5. Wait for installation
6. System shuts down when complete
7. Remove USB and boot from installed disk

## Customization

### Change Auto-behavior

Edit `install-image.sh` and set environment variables:

**Fully automatic (no prompts):**
```bash
export AUTO_INSTALL=true
export AUTO_TARGET_DISK=/dev/nvme0n1
export AUTO_SHUTDOWN=true
```

**Manual installation:**
Remove the auto-installer service and run `/root/install-image.sh` manually.

### Add Verification

Modify `install-image.sh` to verify the image after writing:

```bash
# After dd command
EXPECTED_MD5=$(md5sum "$IMAGE_FILE" | cut -d' ' -f1)
ACTUAL_MD5=$(dd if="$TARGET_DISK" bs=4M count=... | md5sum | cut -d' ' -f1)
```

## How It Works

### Boot Sequence

```
USB Boot
  ↓
systemd-boot starts
  ↓
Ubuntu minimal system boots
  ↓
Auto-login to root on tty1
  ↓
auto-installer.service runs
  ↓
/root/install-image.sh executes
  ↓
User selects target disk
  ↓
Image written with dd
  ↓
System shuts down
```

### Components

**auto-installer.service** - Systemd service that runs installer on boot
**install-image.sh** - Interactive installer script
**mkosi.postinst** - Configures auto-login and enables service

## Troubleshooting

### Installer doesn't auto-start

Boot from USB and check:
```bash
systemctl status auto-installer.service
journalctl -u auto-installer.service
```

### Can't select disk

Make sure the disk is:
- Not the boot device (USB itself)
- Not mounted
- Accessible to root

### Image not found

Verify the image was copied:
```bash
ls -lh /root/custom-ubuntu.img
```

### Build fails

Make sure you built the main image first:
```bash
cd ../mkosi-ubuntu-project
sudo mkosi build
```

## Advanced Usage

### Compress Image (Save Space)

```bash
# Compress the image before copying
gzip -9 mkosi.output/custom-ubuntu_1.0.raw

# Modify install-image.sh to decompress:
gunzip -c custom-ubuntu.img.gz | dd of=$TARGET_DISK bs=4M status=progress
```

### Network-based Installation

Instead of including the image on USB, download it:

```bash
# In install-image.sh, replace image file with:
IMAGE_URL="http://your-server/custom-ubuntu.img"
curl -L "$IMAGE_URL" | dd of=$TARGET_DISK bs=4M status=progress
```

### Multi-image Installer

Create menu to select from multiple images:

```bash
# Copy multiple images to /root/
# Modify install-image.sh to show menu
```

## Tips

- **Test in VM first** using `mkosi qemu`
- **Label USB clearly** to avoid confusion
- **Keep backup** of original image
- **Verify checksums** after writing
- **Use `pv`** for better progress display:
  ```bash
  pv custom-ubuntu.img | dd of=/dev/sdX bs=4M
  ```
