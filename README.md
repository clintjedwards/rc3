# RC3

## Development

### Spin up development Proxmox

In order to develop against the proxmox API we'll need a test instance of Proxmox. You can create one using your
favorite virtual machine manager. Below are instructions of QEMU on a Linux system.

1) Install QEMU

```bash
sudo pacman -Syu --needed qemu-base libvirt edk2-ovmf virt-manager dnsmasq ebtables
```

2) Start libvirtd

```bash
sudo systemctl enable --now libvirtd
```

3) Download the Proxmox VE ISO

```bash
wget https://enterprise.proxmox.com/iso/proxmox-ve_8.3-1.iso -O proxmox.iso
```

4) Create a new QEMU Disk for Proxmox

```bash
qemu-img create -f qcow2 proxmox.qcow2 20G
```

5) Start Proxmox from the iso installer

```bash
qemu-system-x86_64 \
  -m 4096 \
  -smp 2 \
  -drive file=/home/clintjedwards/Downloads/proxmox.qcow2,format=qcow2 \
  -cdrom /home/clintjedwards/Downloads/proxmox.iso \
  -boot d \
  -net nic -net user,hostfwd=tcp::8006-:8006 \
  -cpu host -enable-kvm
```

6) To access the console you can use a vncviewer

```bash
# VNC server running on ::1:5900
gvncviewer :0
```

7) From here on out you can launch Proxmox without the installer image and access it via the web console.

```bash
qemu-system-x86_64 \
  -m 4096 \
  -smp 2 \
  -drive file=/home/clintjedwards/Downloads/proxmox.qcow2,format=qcow2 \
  -boot d \
  -net nic -net user,hostfwd=tcp::8006-:8006 \
  -cpu host -enable-kvm
```

```bash
http://localhost:8006
```

