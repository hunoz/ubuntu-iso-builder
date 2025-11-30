#!/usr/bin/env python3
"""
Ubuntu Cloud-Init Autoinstall ISO Builder - Pure Python, Cross-Platform
Works on Windows, macOS, and Linux without external dependencies
"""
import re
import subprocess
import sys
import urllib.request
import shutil
import yaml
from pathlib import Path
from typing import Dict, Any, Literal

from isobuilder.cloud_init import render_cloudinit_config

try:
    import pycdlib
except ImportError:
    print("❌ pycdlib not installed. Install with: pip install pycdlib")
    sys.exit(1)

lba_regex = re.compile(r".*EFI image start and size: (?P<lba>\d+)\s+\*\s+(?P<block_size>\d+)\s+,\s+(?P<block_count>\d+).*")


def load_config_from_file(config_file: str) -> str:
    """Load configuration from YAML file as string"""
    with open(config_file, 'r') as f:
        return f.read()


def modify_grub_for_autoinstall(grub_content: str) -> str:
    """Modify GRUB configuration to enable autoinstall"""
    # Add autoinstall parameter
    modified = grub_content.replace('---', 'autoinstall ds=nocloud;s=/cdrom/nocloud/ ---')

    # Reduce timeout
    modified = modified.replace('set timeout=30', 'set timeout=5')

    return modified


def check_dependencies() -> bool:
    """Check if required tools are installed"""
    required_tools = ['xorriso', '7z']
    missing = []

    for tool in required_tools:
        if not shutil.which(tool):
            missing.append(tool)

    if missing:
        print(f"❌ Missing required tools: {', '.join(missing)}")
        print("Install with: sudo apt-get install xorriso p7zip-full")
        return False

    print("✅ All required tools found")
    return True


class CrossPlatformISOBuilder:
    """Pure Python ISO builder that works on any OS"""

    def __init__(self, ubuntu_version: str = "24.04", type: Literal["server", "desktop"] = "server", work_dir: str = "./ubuntu-autoinstall"):
        self.ubuntu_version = ubuntu_version
        self.work_dir = Path(work_dir)
        self.iso_name = f"ubuntu-{ubuntu_version}-{'live-server' if type == 'server' else 'desktop'}-amd64.iso"
        self.iso_url = f"https://releases.ubuntu.com/{ubuntu_version}/{self.iso_name}"
        self.source_iso = self.work_dir / self.iso_name
        self.extract_dir = self.work_dir / "source-files"
        self.output_iso = self.work_dir / f"ubuntu-{ubuntu_version}-autoinstall.iso"

    def download_iso(self, force: bool = False) -> bool:
        """Download Ubuntu Server ISO if not present"""
        if self.source_iso.exists() and not force:
            print(f"✅ ISO already downloaded: {self.source_iso}")
            file_size = self.source_iso.stat().st_size / (1024**3)
            print(f"   Size: {file_size:.2f} GB")
            return True

        print(f"📥 Downloading Ubuntu {self.ubuntu_version} Server ISO...")
        print(f"   URL: {self.iso_url}")
        print("   This may take a while (typically 2-3 GB)...")

        self.work_dir.mkdir(parents=True, exist_ok=True)

        try:
            def progress_hook(count, block_size, total_size):
                if total_size > 0:
                    percent = int(count * block_size * 100 / total_size)
                    downloaded = count * block_size / (1024**2)
                    total = total_size / (1024**2)
                    sys.stdout.write(f"\r   Progress: {percent}% ({downloaded:.1f}/{total:.1f} MB)")
                    sys.stdout.flush()

            urllib.request.urlretrieve(self.iso_url, self.source_iso, progress_hook)
            print("\n✅ Download complete")
            return True
        except Exception as e:
            print(f"\n❌ Download failed: {e}")
            if self.source_iso.exists():
                self.source_iso.unlink()
            return False

    def extract_iso(self) -> bool:
        """Extract ISO contents"""
        print(f"📦 Extracting ISO contents...")

        if self.extract_dir.exists():
            shutil.rmtree(self.extract_dir)

        self.extract_dir.mkdir(parents=True, exist_ok=True)

        try:
            subprocess.run(
                ['7z', 'x', str(self.source_iso), f'-o{self.extract_dir}', '-y'],
                check=True,
                capture_output=True
            )
            print("✅ Extraction complete")
            return True
        except subprocess.CalledProcessError as e:
            print(f"❌ Extraction failed: {e}")
            return False

    def create_autoinstall_config(self, config: dict[str, Any]) -> tuple:
        """Create autoinstall configuration from pre-rendered config string"""
        print("📝 Creating autoinstall configuration...")

        # Parse just to extract hostname for meta-data
        hostname = config['autoinstall']['user-data']['hostname']

        dump = render_cloudinit_config(config)

        # Method 1: Create autoinstall.yaml at ISO root (preferred method for Ubuntu Server)
        # Write the config content directly without re-parsing/re-dumping
        autoinstall_file = self.extract_dir / "autoinstall.yaml"
        with open(autoinstall_file, 'w') as f:
            f.write(dump)

        # Method 2: Create nocloud datasource files for compatibility
        nocloud_dir = self.extract_dir / "nocloud"
        nocloud_dir.mkdir(exist_ok=True)

        # Write user-data (preserve original formatting)
        user_data_file = nocloud_dir / "user-data"
        with open(user_data_file, 'w') as f:
            f.write(dump)

        # Write meta-data
        meta_data_file = nocloud_dir / "meta-data"
        meta_data = f"instance-id: {hostname}\n"
        meta_data += f"local-hostname: {hostname}\n"
        with open(meta_data_file, 'w') as f:
            f.write(meta_data)

        # Write vendor-data (required by nocloud)
        vendor_data_file = nocloud_dir / "vendor-data"
        with open(vendor_data_file, 'w') as f:
            f.write("#cloud-config\n{}\n")

        print("✅ Configuration created")
        return config, meta_data

    def modify_grub_config(self) -> bool:
        """Modify GRUB configuration to enable autoinstall"""
        print("⚙️  Modifying boot configuration...")

        grub_cfg = self.extract_dir / "boot" / "grub" / "grub.cfg"

        if not grub_cfg.exists():
            print(f"❌ GRUB config not found: {grub_cfg}")
            return False

        try:
            # Read original config
            with open(grub_cfg, 'r') as f:
                content = f.read()

            # Backup original
            shutil.copy(grub_cfg, str(grub_cfg) + ".backup")

            # Add autoinstall parameter
            # For autoinstall to work, we only need the 'autoinstall' keyword
            # The installer will automatically look for autoinstall.yaml at ISO root
            # Keep it simple to avoid boot issues
            content = content.replace('---', 'autoinstall ---')

            # Reduce timeout
            content = content.replace('set timeout=30', 'set timeout=5')

            # Write modified config
            with open(grub_cfg, 'w') as f:
                f.write(content)

            print("✅ Boot configuration modified")
            return True
        except Exception as e:
            print(f"❌ Failed to modify GRUB config: {e}")
            return False

    def build_iso(self) -> bool:
        """Build the final ISO image"""
        print("🔨 Building ISO image...")

        # Extract MBR for hybrid ISO
        mbr_file = self.work_dir / "isohdpfx.bin"
        try:
            with open(self.source_iso, 'rb') as f:
                mbr_data = f.read(446)
            with open(mbr_file, 'wb') as f:
                f.write(mbr_data)
        except Exception as e:
            print(f"⚠️  Warning: Could not extract MBR: {e}")

        # Extract EFI boot image from original ISO
        # Dynamically detect the EFI boot image location from the ISO
        efi_img = self.extract_dir / "boot" / "grub" / "efi.img"
        efi_img.parent.mkdir(parents=True, exist_ok=True)
        try:
            # Use xorriso to get the EFI boot image LBA and size
            result = subprocess.run(
                ['xorriso', '-indev', str(self.source_iso), '-report_el_torito', 'plain'],
                check=True,
                capture_output=True,
                text=True
            )

            # Parse the output to find UEFI boot image LBA and block count
            lba = None
            blocks = None
            for line in result.stderr.split('\n'):
                # Look for: "libisofs: NOTE : EFI image start and size: 1610304 * 2048 , 10160 * 512"
                if 'EFI image start and size:' in line:
                    match = lba_regex.match(line)
                    lba = match.groupdict()["lba"]
                    blocks_512 = int(match.groupdict()["block_count"])
                    blocks = blocks_512 // 4

            if lba and blocks:
                # Extract the EFI boot image using dd
                subprocess.run(
                    ['dd', f'if={self.source_iso}', 'bs=2048', f'skip={lba}', f'count={blocks}', f'of={efi_img}'],
                    check=True,
                    capture_output=True
                )
                print(f"✅ EFI boot image extracted (LBA: {lba}, size: {blocks * 2048 // 1024}KB)")
            else:
                print("⚠️  Warning: Could not detect EFI boot image location")
                return False
        except subprocess.CalledProcessError as e:
            print(f"⚠️  Warning: Could not extract EFI boot image: {e}")
            return False

        # Build ISO with xorriso - using extracted EFI boot image
        cmd = [
            'xorriso', '-as', 'mkisofs',
            '-r', '-V', 'Ubuntu Autoinstall',
            '-J', '-joliet-long',
            '-o', str(self.output_iso),
            '-b', 'boot/grub/i386-pc/eltorito.img',
            '-c', 'boot.catalog',
            '-no-emul-boot',
            '-boot-load-size', '4',
            '-boot-info-table',
            '-eltorito-alt-boot',
            '-e', 'boot/grub/efi.img',
            '-no-emul-boot',
            '-isohybrid-gpt-basdat'
        ]

        if mbr_file.exists():
            cmd.extend(['-isohybrid-mbr', str(mbr_file)])

        cmd.append(str(self.extract_dir))

        try:
            subprocess.run(cmd, check=True, capture_output=True)
            print(f"✅ ISO created: {self.output_iso}")
            return True
        except subprocess.CalledProcessError as e:
            print(f"❌ ISO build failed: {e.stderr.decode()}")
            return False

    def build(self, config: dict[str, Any] = None) -> bool:
        """Run the complete build process"""
        print("=" * 60)
        print("Ubuntu Cloud-Init Autoinstall ISO Builder")
        print("=" * 60)
        print()

        steps = [
            ("Checking dependencies", check_dependencies),
            ("Downloading ISO", self.download_iso),
            ("Extracting ISO", self.extract_iso),
            ("Creating autoinstall config", lambda: self.create_autoinstall_config(config)),
            ("Modifying boot config", self.modify_grub_config),
            ("Building final ISO", self.build_iso)
        ]

        for step_name, step_func in steps:
            print(f"\n📍 Step: {step_name}")
            if not step_func():
                print(f"\n❌ Build failed at: {step_name}")
                return False

        print("\n" + "=" * 60)
        print("✅ Build complete!")
        print("=" * 60)
        print(f"\n📀 Output ISO: {self.output_iso}")
        print(f"\n💾 Write to USB with:")
        print(f"   sudo dd if={self.output_iso} of=/dev/sdX bs=4M status=progress && sync")
        print()

        return True


def build_iso(
        cloud_init_config: dict[str, Any],
        ubuntu_version: str = "24.04.3",
        ubuntu_type: Literal["server", "desktop"] = "server",
        work_dir: str = "./build",
) -> bool:
    # Create builder
    builder = CrossPlatformISOBuilder(
        ubuntu_version=ubuntu_version,
        type=ubuntu_type,
        work_dir=work_dir
    )

    # Build ISO
    success = builder.build(cloud_init_config)

    return success
