#!/bin/bash
set -e

echo "USB Installer Builder"
echo "====================="
echo ""

# Check if main image exists
MAIN_IMAGE="../mkosi-ubuntu-project/mkosi.output/custom-ubuntu_1.0.raw"

if [ ! -f "$MAIN_IMAGE" ]; then
    echo "ERROR: Main image not found at: $MAIN_IMAGE"
    echo ""
    echo "Please build your main image first:"
    echo "  cd ../mkosi-ubuntu-project"
    echo "  sudo mkosi build"
    exit 1
fi

echo "✓ Found main image: $MAIN_IMAGE"

# Copy image to installer
IMAGE_SIZE=$(du -h "$MAIN_IMAGE" | cut -f1)
echo "  Size: $IMAGE_SIZE"
echo ""
echo "Copying image to installer..."

cp "$MAIN_IMAGE" mkosi.extra/root/custom-ubuntu.img

echo "✓ Image copied"
echo ""
echo "Building installer USB image..."
echo ""

# Build installer
sudo mkosi build

echo ""
echo "╔════════════════════════════════════════════════════════╗"
echo "║              ✓ INSTALLER BUILD COMPLETE                ║"
echo "╚════════════════════════════════════════════════════════╝"
echo ""
echo "Installer image: mkosi.output/ubuntu-installer_1.0.raw"
echo ""
echo "To write to USB:"
echo "  sudo dd if=mkosi.output/ubuntu-installer_1.0.raw of=/dev/sdX bs=4M status=progress"
echo ""
echo "Replace /dev/sdX with your USB device (check with 'lsblk')"
echo ""
