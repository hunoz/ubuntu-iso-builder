#!/bin/bash
set -e

# Automatic Image Installer
# This script runs on boot and writes the mkosi image to the target disk

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
IMAGE_FILE="${SCRIPT_DIR}/custom-ubuntu.img"
LOG_FILE="/var/log/auto-installer.log"

# Logging function
log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $*" | tee -a "$LOG_FILE"
}

# Display banner
clear
cat << 'EOF'
╔════════════════════════════════════════════════════════╗
║                                                        ║
║          AUTOMATIC UBUNTU IMAGE INSTALLER              ║
║                                                        ║
╚════════════════════════════════════════════════════════╝

This installer will write a pre-configured Ubuntu image
to your target disk.

⚠️  WARNING: ALL DATA ON THE TARGET DISK WILL BE ERASED! ⚠️

EOF

log "Installer started"

# Check if image exists
if [ ! -f "$IMAGE_FILE" ]; then
    log "ERROR: Image file not found: $IMAGE_FILE"
    echo "ERROR: Image file not found!"
    echo "Expected: $IMAGE_FILE"
    exit 1
fi

# Get image size
IMAGE_SIZE=$(du -h "$IMAGE_FILE" | cut -f1)
log "Image file found: $IMAGE_FILE (Size: $IMAGE_SIZE)"

# List available disks (excluding current boot device and loop devices)
echo ""
echo "Available disks:"
echo "================"
lsblk -d -n -o NAME,SIZE,TYPE,MODEL | grep disk | grep -v loop | nl
echo ""

# Detect current boot device
BOOT_DEVICE=$(mount | grep "on / " | cut -d' ' -f1 | sed 's/[0-9]*$//')
log "Boot device detected: $BOOT_DEVICE"

# Function to select target disk
select_disk() {
    while true; do
        echo ""
        read -p "Enter the target disk (e.g., sda, nvme0n1) or 'list' to show disks again: " DISK_NAME
        
        if [ "$DISK_NAME" = "list" ]; then
            echo ""
            lsblk -d -n -o NAME,SIZE,TYPE,MODEL | grep disk | grep -v loop | nl
            continue
        fi
        
        TARGET_DISK="/dev/$DISK_NAME"
        
        # Check if disk exists
        if [ ! -b "$TARGET_DISK" ]; then
            echo "ERROR: Disk $TARGET_DISK does not exist!"
            continue
        fi
        
        # Prevent writing to boot device
        if [ "$TARGET_DISK" = "$BOOT_DEVICE" ]; then
            echo "ERROR: Cannot write to boot device ($BOOT_DEVICE)!"
            echo "Please select a different disk."
            continue
        fi
        
        # Show disk info
        echo ""
        echo "Selected disk: $TARGET_DISK"
        lsblk "$TARGET_DISK"
        echo ""
        
        # Confirm
        read -p "⚠️  Write image to $TARGET_DISK? This will ERASE ALL DATA! (yes/no): " CONFIRM
        
        if [ "$CONFIRM" = "yes" ]; then
            break
        else
            echo "Cancelled. Please select again."
        fi
    done
}

# Interactive or automatic mode
if [ "${AUTO_INSTALL:-}" = "true" ] && [ -n "${AUTO_TARGET_DISK:-}" ]; then
    # Automatic mode (for unattended installation)
    TARGET_DISK="$AUTO_TARGET_DISK"
    log "Auto-install mode: target=$TARGET_DISK"
    
    if [ ! -b "$TARGET_DISK" ]; then
        log "ERROR: Auto-install target disk $TARGET_DISK does not exist"
        exit 1
    fi
    
    if [ "$TARGET_DISK" = "$BOOT_DEVICE" ]; then
        log "ERROR: Auto-install cannot target boot device"
        exit 1
    fi
else
    # Interactive mode
    select_disk
fi

log "Target disk selected: $TARGET_DISK"

# Unmount any mounted partitions on target disk
echo ""
echo "Unmounting any mounted partitions on $TARGET_DISK..."
umount ${TARGET_DISK}* 2>/dev/null || true
log "Unmounted partitions on $TARGET_DISK"

# Write image
echo ""
echo "╔════════════════════════════════════════════════════════╗"
echo "║  Writing image to $TARGET_DISK...                          "
echo "║  This may take several minutes.                        ║"
echo "╚════════════════════════════════════════════════════════╝"
echo ""

log "Starting image write to $TARGET_DISK"

# Use dd with status=progress
if dd if="$IMAGE_FILE" of="$TARGET_DISK" bs=4M status=progress conv=fsync; then
    sync
    log "Image successfully written to $TARGET_DISK"
    
    echo ""
    echo "╔════════════════════════════════════════════════════════╗"
    echo "║                                                        ║"
    echo "║              ✓ INSTALLATION COMPLETE!                 ║"
    echo "║                                                        ║"
    echo "╚════════════════════════════════════════════════════════╝"
    echo ""
    echo "You can now:"
    echo "  1. Remove the USB installer"
    echo "  2. Shutdown or reboot"
    echo "  3. Boot from $TARGET_DISK"
    echo ""
    
    log "Installation completed successfully"
    
    # Optional: Auto-shutdown
    if [ "${AUTO_SHUTDOWN:-}" = "true" ]; then
        echo "Auto-shutdown in 10 seconds... (Ctrl+C to cancel)"
        sleep 10
        shutdown -h now
    else
        read -p "Press Enter to shutdown, or Ctrl+C to stay in the installer..."
        shutdown -h now
    fi
else
    log "ERROR: Image write failed"
    echo ""
    echo "ERROR: Installation failed!"
    echo "Check $LOG_FILE for details."
    exit 1
fi
