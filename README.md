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

### Connecting to Proxmox for Local Development

Before running the backend, you’ll need to create a Proxmox API token and configure the environment.

#### Create API Token in Proxmox

1. Log in to your Proxmox instance (typically at `http://localhost:8006` if you're using the above commands).
2. Navigate to **Datacenter → Permissions → API Tokens**.
3. Select your user (e.g. `root@pam`) and click **Add**.
4. For testing purposes, **disable *Privilege Separation*** (this allows the token full access for now — do not do this in production).
5. Save the token. You will be provided a **Token ID** and **Token Secret**.

#### Export Required Environment Variables

Once you have your token, export the following environment variables:

```bash
export RC3_PROXMOX__TOKEN_ID='root@pam!your-token-id'
export RC3_PROXMOX__TOKEN_SECRET='your-token-secret'
```

You'll then be able to run `make run-backend` to get RC3 to connect to Proxmox.
