#!/bin/bash
# Garage AI Boot Image Creation Script
# This script creates a bootable USB image for Garage AI

set -e

echo "🚀 Garage AI Boot Image Creator"
echo "================================"

# Configuration
IMAGE_SIZE="8G"  # 8GB image
IMAGE_NAME="garage-ai-boot.img"
MOUNT_POINT="/mnt/garage-boot"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if running as root
if [[ $EUID -ne 0 ]]; then
   log_error "This script must be run as root (sudo)"
   exit 1
fi

# Install required tools
log_info "Installing required tools..."
apt update
apt install -y debootstrap squashfs-tools grub-pc-bin grub-efi-amd64-bin \
               xorriso isolinux syslinux-common syslinux-utils

# Create empty image file
log_info "Creating ${IMAGE_SIZE} image file..."
dd if=/dev/zero of="$IMAGE_NAME" bs=1 count=0 seek="$IMAGE_SIZE"

# Set up loop device
LOOP_DEVICE=$(losetup -f)
losetup "$LOOP_DEVICE" "$IMAGE_NAME"

# Create partitions
log_info "Creating partitions..."
parted -s "$LOOP_DEVICE" mklabel gpt
parted -s "$LOOP_DEVICE" mkpart EFI fat32 1MiB 512MiB
parted -s "$LOOP_DEVICE" set 1 esp on
parted -s "$LOOP_DEVICE" mkpart primary ext4 512MiB 100%

# Format partitions
log_info "Formatting partitions..."
mkfs.vfat -F 32 "${LOOP_DEVICE}p1"
mkfs.ext4 "${LOOP_DEVICE}p2"

# Mount partitions
log_info "Mounting partitions..."
mkdir -p "$MOUNT_POINT"
mount "${LOOP_DEVICE}p2" "$MOUNT_POINT"
mkdir -p "$MOUNT_POINT/boot/efi"
mount "${LOOP_DEVICE}p1" "$MOUNT_POINT/boot/efi"

# Bootstrap Ubuntu
log_info "Bootstrapping Ubuntu base system..."
debootstrap --arch=amd64 jammy "$MOUNT_POINT" http://archive.ubuntu.com/ubuntu/

# Configure the system
log_info "Configuring the system..."

# Mount system directories
mount -t proc proc "$MOUNT_POINT/proc"
mount -t sysfs sys "$MOUNT_POINT/sys"
mount --bind /dev "$MOUNT_POINT/dev"
mount --bind /dev/pts "$MOUNT_POINT/dev/pts"

# DNS resolution
cp /etc/resolv.conf "$MOUNT_POINT/etc/resolv.conf"

# Chroot and configure
chroot "$MOUNT_POINT" /bin/bash << 'EOF'

# Set hostname
echo "garage-ai-boot" > /etc/hostname

# Configure networking
cat > /etc/netplan/01-netcfg.yaml << 'NETCFG'
network:
  version: 2
  ethernets:
    ens18:
      dhcp4: true
NETCFG

# Update package list
apt update
apt upgrade -y

# Install base packages
apt install -y linux-generic linux-headers-generic ubuntu-server \
               network-manager netplan.io openssh-server \
               python3 python3-pip python3-venv git curl wget \
               htop iotop nload

# Install NVIDIA drivers (generic - will be configured at runtime)
apt install -y nvidia-driver-470 nvidia-cuda-toolkit clinfo

# Create garage user
useradd -m -s /bin/bash garage
usermod -aG sudo garage
echo "garage:garage" | chpasswd

# Configure sudo
echo "garage ALL=(ALL) NOPASSWD:ALL" >> /etc/sudoers.d/garage

# Configure auto-login for garage user
mkdir -p /etc/systemd/system/getty@tty1.service.d
cat > /etc/systemd/system/getty@tty1.service.d/override.conf << 'AUTOLOGIN'
[Service]
ExecStart=
ExecStart=-/sbin/agetty --noissue --autologin garage %I $TERM
Type=idle
AUTOLOGIN

# Create garage-ai directory structure
mkdir -p /opt/garage-ai
mkdir -p /var/log/garage-ai

# Set permissions
chown garage:garage /opt/garage-ai
chown garage:garage /var/log/garage-ai

EOF

# Copy Garage AI files
log_info "Installing Garage AI components..."
cp -r garage-ai-files/* "$MOUNT_POINT/opt/garage-ai/"
chown -R garage:garage "$MOUNT_POINT/opt/garage-ai"

# Install bootloader
log_info "Installing bootloader..."
chroot "$MOUNT_POINT" /bin/bash << 'EOF'

# Install GRUB
apt install -y grub-efi-amd64-signed shim-signed

# Configure GRUB
cat > /etc/default/grub << 'GRUBCFG'
GRUB_DEFAULT=0
GRUB_TIMEOUT=5
GRUB_DISTRIBUTOR="Garage AI"
GRUB_CMDLINE_LINUX_DEFAULT="quiet splash nomodeset"
GRUB_CMDLINE_LINUX=""
GRUBCFG

# Install GRUB to EFI partition
grub-install --target=x86_64-efi --efi-directory=/boot/efi --bootloader-id=garage-ai --recheck
update-grub
EOF

# Cleanup
log_info "Cleaning up..."
umount "$MOUNT_POINT/dev/pts"
umount "$MOUNT_POINT/dev"
umount "$MOUNT_POINT/sys"
umount "$MOUNT_POINT/proc"
umount "$MOUNT_POINT/boot/efi"
umount "$MOUNT_POINT"

# Detach loop device
losetup -d "$LOOP_DEVICE"

log_info "Boot image created successfully: $IMAGE_NAME"
log_info "To create bootable USB:"
log_info "  sudo dd if=$IMAGE_NAME of=/dev/sdX bs=4M status=progress"
log_warn "Replace /dev/sdX with your actual USB device (use lsblk to find it)"
log_warn "WARNING: This will erase all data on the target device!"
