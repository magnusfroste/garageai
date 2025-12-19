# Garage AI Boot Stick - Technical Specification

## Overview

The Garage AI Boot Stick enables gaming PCs to participate in distributed AI inference without disrupting normal gaming activities. Users boot from USB to run AI workloads, then reboot to Windows for gaming.

## Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Gaming PC     │    │   USB Boot      │    │   Garage AI     │
│   (Windows)     │───▶│   Environment   │───▶│   Network       │
│                 │    │                 │    │                 │
│ • RTX 4090      │    │ • Ubuntu 22.04  │    │ • Kubernetes    │
│ • 32GB RAM      │    │ • NVIDIA drivers│    │ • vLLM          │
│ • SSD Storage   │    │ • Python AI env │    │ • Token rewards │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## Hardware Requirements

### Minimum Specifications
```yaml
cpu: "AMD Ryzen 5 3600 / Intel i5-10600K or equivalent (6+ cores)"
ram: "16GB DDR4 (32GB recommended for larger models)"
gpu: "NVIDIA RTX 3060 (12GB VRAM) or equivalent"
storage: "32GB USB 3.0+ drive (64GB recommended)"
network: "100Mbps stable internet connection"
motherboard: "UEFI-compatible with boot menu access"
```

### Supported GPUs
- **NVIDIA RTX 30/40 Series**: Full support with CUDA 11.8+
- **AMD RX 6000/7000 Series**: ROCm support (experimental)
- **Intel Arc A-Series**: OneAPI support (experimental)

## Boot Image Creation

### Base OS: Ubuntu Server 22.04.3 LTS

#### Installation Script
```bash
#!/bin/bash
# create_boot_image.sh

# Create bootable USB image
sudo apt update && sudo apt install -y debootstrap squashfs-tools

# Create base Ubuntu filesystem
sudo debootstrap --arch=amd64 focal /mnt/ubuntu http://archive.ubuntu.com/ubuntu/

# Mount and configure
sudo mount --bind /dev /mnt/ubuntu/dev
sudo mount --bind /proc /mnt/ubuntu/proc
sudo mount --bind /sys /mnt/ubuntu/sys

# Chroot into filesystem
sudo chroot /mnt/ubuntu /bin/bash << 'EOF'

# Update and install base packages
apt update && apt upgrade -y
apt install -y ubuntu-server linux-generic initramfs-tools

# Install NVIDIA drivers
apt install -y nvidia-driver-470 nvidia-cuda-toolkit

# Install AI dependencies
apt install -y python3 python3-pip python3-venv git curl wget

# Create AI user
useradd -m -s /bin/bash garage
usermod -aG sudo garage
echo "garage:garage" | chpasswd

EOF

# Create squashfs image
sudo mksquashfs /mnt/ubuntu filesystem.squashfs -comp xz

# Create ISO with GRUB bootloader
# (Additional GRUB configuration needed)
```

### AI Environment Setup

#### Conda Environment Creation
```bash
#!/bin/bash
# setup_ai_environment.sh

# Install Miniconda
wget https://repo.anaconda.com/miniconda/Miniconda3-latest-Linux-x86_64.sh
bash Miniconda3-latest-Linux-x86_64.sh -b -p /opt/miniconda

# Create Garage AI environment
/opt/miniconda/bin/conda create -n garage-ai python=3.9 -y
/opt/miniconda/bin/conda activate garage-ai

# Install PyTorch with CUDA support
pip install torch torchvision torchaudio --index-url https://download.pytorch.org/whl/cu118

# Install core AI libraries
pip install transformers accelerate vllm ray[all] fastapi uvicorn

# Install Garage AI client
pip install git+https://github.com/garage-ai/client.git

# Install system monitoring tools
pip install psutil GPUtil
```

### Boot Configuration

#### GRUB Bootloader Setup
```bash
# /boot/grub/grub.cfg
set timeout=10
set default=0

menuentry "Garage AI Inference Mode" {
    linux /boot/vmlinuz root=/dev/sdb1 ro quiet splash
    initrd /boot/initrd.img
}

menuentry "Ubuntu Recovery Mode" {
    linux /boot/vmlinuz root=/dev/sdb1 ro recovery nomodeset
    initrd /boot/initrd.img
}
```

#### Auto-Login Configuration
```bash
# /etc/lightdm/lightdm.conf
[Seat:*]
autologin-user=garage
autologin-user-timeout=0
```

## Node Registration & Management

### Auto-Discovery Service
```python
# /opt/garage-ai/discovery.py
import requests
import json
import psutil
import GPUtil
import uuid
from geopy.geocators import Nominatim

class GarageNodeDiscovery:
    def __init__(self):
        self.api_url = "https://api.garage.ai"
        self.node_id = self.get_or_create_node_id()

    def get_hardware_info(self):
        """Detect and catalog hardware capabilities"""
        gpus = GPUtil.getGPUs()
        gpu_info = []
        for gpu in gpus:
            gpu_info.append({
                'id': gpu.id,
                'name': gpu.name,
                'memory_total': gpu.memoryTotal,
                'memory_free': gpu.memoryFree,
                'uuid': gpu.uuid
            })

        cpu_info = {
            'cores': psutil.cpu_count(),
            'frequency': psutil.cpu_freq().max,
            'usage': psutil.cpu_percent()
        }

        ram_info = {
            'total': psutil.virtual_memory().total,
            'available': psutil.virtual_memory().available
        }

        return {
            'cpu': cpu_info,
            'ram': ram_info,
            'gpus': gpu_info,
            'network': self.get_network_info()
        }

    def get_location(self):
        """Get approximate location for regional routing"""
        try:
            geolocator = Nominatim(user_agent="garage-ai-node")
            location = geolocator.geocode("Sweden")  # Default to Sweden
            return {
                'country': location.address.split(',')[-1].strip(),
                'latitude': location.latitude,
                'longitude': location.longitude
            }
        except:
            return {'country': 'Sweden', 'latitude': 59.3293, 'longitude': 18.0686}

    def register_node(self):
        """Register node with Garage AI network"""
        hardware = self.get_hardware_info()
        location = self.get_location()

        payload = {
            'node_id': self.node_id,
            'hardware': hardware,
            'location': location,
            'capabilities': self.test_capabilities(),
            'version': '1.0.0'
        }

        response = requests.post(f"{self.api_url}/nodes/register", json=payload)
        if response.status_code == 200:
            data = response.json()
            with open('/opt/garage-ai/node_token', 'w') as f:
                f.write(data['token'])
            return True
        return False

    def test_capabilities(self):
        """Test AI inference capabilities"""
        try:
            import torch
            return {
                'cuda_available': torch.cuda.is_available(),
                'cuda_devices': torch.cuda.device_count(),
                'vllm_supported': self.test_vllm(),
                'max_model_size': self.estimate_max_model_size()
            }
        except:
            return {'error': 'AI libraries not available'}

    def get_or_create_node_id(self):
        """Generate persistent node ID"""
        try:
            with open('/opt/garage-ai/node_id', 'r') as f:
                return f.read().strip()
        except:
            node_id = f"garage_{uuid.uuid4().hex[:12]}"
            with open('/opt/garage-ai/node_id', 'w') as f:
                f.write(node_id)
            return node_id
```

### Service Management
```bash
# /etc/systemd/system/garage-ai.service
[Unit]
Description=Garage AI Inference Service
After=network.target nvidia-persistenced.service

[Service]
Type=simple
User=garage
ExecStart=/opt/garage-ai/start_inference.sh
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

#### Start Script
```bash
#!/bin/bash
# /opt/garage-ai/start_inference.sh

# Activate conda environment
source /opt/miniconda/bin/activate garage-ai

# Start vLLM server
python -m vllm.entrypoints.api_server \
    --model microsoft/DialoGPT-large \
    --tensor-parallel-size 1 \
    --gpu-memory-utilization 0.9 \
    --host 0.0.0.0 \
    --port 8000 \
    --api-key $(cat /opt/garage-ai/node_token)

# Start monitoring service
python /opt/garage-ai/monitor.py &
```

## Performance Optimization

### vLLM Configuration
```python
# /opt/garage-ai/config/vllm_config.py
vllm_config = {
    'model': 'microsoft/DialoGPT-large',
    'tensor_parallel_size': 1,
    'gpu_memory_utilization': 0.85,
    'max_model_len': 2048,
    'quantization': 'awq',  # 4-bit quantization
    'dtype': 'float16',
    'seed': 42,
    'enforce_eager': False,
    'max_num_seqs': 16,
    'max_num_batched_tokens': 2048
}
```

### Memory Management
```python
# /opt/garage-ai/memory_manager.py
import torch
import gc

class MemoryManager:
    def __init__(self):
        self.gpu_memory_threshold = 0.9

    def optimize_memory(self):
        """Aggressive memory cleanup"""
        if torch.cuda.is_available():
            torch.cuda.empty_cache()
        gc.collect()

    def should_accept_request(self, estimated_tokens):
        """Check if we can handle new request"""
        if not torch.cuda.is_available():
            return False

        free_memory = torch.cuda.get_device_properties(0).total_memory * (1 - self.gpu_memory_threshold)
        estimated_usage = estimated_tokens * 2 * 1024 * 1024  # Rough estimate

        return free_memory > estimated_usage
```

## Security & Privacy

### Data Protection
- **Local Processing Only**: No user data leaves the PC
- **Encrypted Communication**: TLS 1.3 for all network traffic
- **Model Encryption**: Downloaded models are encrypted
- **Secure Boot**: UEFI secure boot verification

### Node Authentication
```python
# /opt/garage-ai/auth.py
import jwt
import datetime

class NodeAuthenticator:
    def __init__(self):
        self.secret_key = self.load_secret_key()

    def generate_token(self, node_id):
        """Generate JWT for node authentication"""
        payload = {
            'node_id': node_id,
            'iat': datetime.datetime.utcnow(),
            'exp': datetime.datetime.utcnow() + datetime.timedelta(days=30)
        }
        return jwt.encode(payload, self.secret_key, algorithm='HS256')

    def validate_token(self, token):
        """Validate JWT token"""
        try:
            payload = jwt.decode(token, self.secret_key, algorithms=['HS256'])
            return payload['node_id']
        except jwt.ExpiredSignatureError:
            return None
        except jwt.InvalidTokenError:
            return None
```

## Monitoring & Logging

### System Monitoring
```python
# /opt/garage-ai/monitor.py
import time
import psutil
import GPUtil
import requests

class SystemMonitor:
    def __init__(self):
        self.api_url = "https://api.garage.ai"
        self.node_token = self.load_token()

    def collect_metrics(self):
        """Collect system and performance metrics"""
        cpu_percent = psutil.cpu_percent()
        ram_percent = psutil.virtual_memory().percent

        gpu_metrics = []
        gpus = GPUtil.getGPUs()
        for gpu in gpus:
            gpu_metrics.append({
                'id': gpu.id,
                'utilization': gpu.load * 100,
                'memory_used': gpu.memoryUsed,
                'memory_total': gpu.memoryTotal,
                'temperature': gpu.temperature
            })

        return {
            'timestamp': time.time(),
            'cpu_usage': cpu_percent,
            'ram_usage': ram_percent,
            'gpu_metrics': gpu_metrics,
            'network_stats': self.get_network_stats()
        }

    def report_status(self):
        """Send metrics to central server"""
        metrics = self.collect_metrics()
        headers = {'Authorization': f'Bearer {self.node_token}'}

        try:
            response = requests.post(f"{self.api_url}/nodes/metrics",
                                   json=metrics, headers=headers, timeout=5)
            return response.status_code == 200
        except:
            return False

    def run_monitoring_loop(self):
        """Continuous monitoring loop"""
        while True:
            self.report_status()
            time.sleep(60)  # Report every minute
```

## User Interface

### Boot Menu Integration
- **UEFI Boot Manager**: Detects USB and shows "Garage AI" option
- **Auto-Return**: Option to reboot directly back to Windows
- **Status Display**: Shows earnings and current session info

### Web Dashboard (Local)
```html
<!-- /var/www/html/dashboard/index.html -->
<!DOCTYPE html>
<html>
<head>
    <title>Garage AI Node Dashboard</title>
    <script src="https://cdn.jsdelivr.net/npm/chart.js"></script>
</head>
<body>
    <h1>Garage AI Node Status</h1>

    <div class="metrics">
        <div id="earnings-today">Today's Earnings: <span id="earnings">0.00 GAI</span></div>
        <div id="uptime">Uptime: <span id="uptime-display">0h 0m</span></div>
        <div id="requests-served">Requests Served: <span id="requests">0</span></div>
    </div>

    <canvas id="performance-chart"></canvas>

    <script src="dashboard.js"></script>
</body>
</html>
```

## Distribution & Updates

### USB Stick Creation
```bash
# create_installer.sh
#!/bin/bash

# Download base image
wget https://releases.garage.ai/boot-image-v1.0.iso

# Create bootable USB
sudo dd if=garage-boot-image.iso of=/dev/sdb bs=4M status=progress

# Verify installation
sync
```

### Over-the-Air Updates
```python
# /opt/garage-ai/updater.py
import requests
import subprocess

class OTAUpdater:
    def __init__(self):
        self.update_url = "https://updates.garage.ai"

    def check_for_updates(self):
        """Check for available updates"""
        response = requests.get(f"{self.update_url}/latest")
        if response.status_code == 200:
            latest_version = response.json()['version']
            current_version = self.get_current_version()

            if latest_version > current_version:
                return response.json()
        return None

    def apply_update(self, update_info):
        """Download and apply update"""
        update_file = update_info['download_url']
        subprocess.run(['wget', update_file])

        # Apply update (could be package or full image)
        if update_info['type'] == 'package':
            subprocess.run(['dpkg', '-i', 'update.deb'])
        elif update_info['type'] == 'image':
            # Full image update requires reboot
            self.schedule_image_update()

        subprocess.run(['systemctl', 'restart', 'garage-ai'])
```

## Testing & Validation

### Hardware Compatibility Tests
```bash
#!/bin/bash
# test_hardware.sh

echo "Testing hardware compatibility..."

# Test GPU
nvidia-smi --query-gpu=name,memory.total --format=csv,noheader,nounits

# Test CUDA
python3 -c "import torch; print(f'CUDA available: {torch.cuda.is_available()}')"

# Test vLLM
python3 -c "from vllm import LLM; llm = LLM(model='microsoft/DialoGPT-small'); print('vLLM OK')"

# Test network
curl -s https://api.garage.ai/health

echo "Hardware tests completed."
```

### Performance Benchmarks
```python
# benchmark.py
import time
from vllm import LLM, SamplingParams

def run_benchmark():
    llm = LLM(model="microsoft/DialoGPT-large")
    sampling_params = SamplingParams(temperature=0.8, top_p=0.95, max_tokens=100)

    prompts = [
        "Hello, how are you?",
        "What is artificial intelligence?",
        "Explain quantum computing in simple terms."
    ] * 10  # 30 prompts total

    start_time = time.time()

    outputs = llm.generate(prompts, sampling_params)

    end_time = time.time()

    total_tokens = sum(len(output.outputs[0].token_ids) for output in outputs)
    total_time = end_time - start_time
    tokens_per_second = total_tokens / total_time

    print(f"Benchmark Results:")
    print(f"Total tokens generated: {total_tokens}")
    print(f"Total time: {total_time:.2f} seconds")
    print(f"Tokens/second: {tokens_per_second:.2f}")
    print(f"Average latency: {total_time/len(prompts):.2f} seconds per request")

if __name__ == "__main__":
    run_benchmark()
```

## Troubleshooting Guide

### Common Issues & Solutions

#### Issue: GPU Not Detected
```bash
# Check NVIDIA drivers
nvidia-smi

# Reinstall drivers
sudo apt purge nvidia-*
sudo apt install nvidia-driver-470

# Check kernel modules
lsmod | grep nvidia
```

#### Issue: Out of Memory
```bash
# Monitor memory usage
nvidia-smi --query-gpu=memory.used,memory.total --format=csv

# Reduce model size or batch size
# Edit vLLM config to use smaller models
```

#### Issue: Network Connectivity
```bash
# Test connectivity
curl -v https://api.garage.ai/health

# Check firewall
sudo ufw status

# DNS resolution
nslookup api.garage.ai
```

## Deployment Checklist

- [ ] USB drive formatted correctly (FAT32 for boot, ext4 for OS)
- [ ] Boot order set in BIOS/UEFI
- [ ] Hardware detection working
- [ ] AI environment installed and activated
- [ ] Node registered with network
- [ ] Services starting automatically
- [ ] Monitoring and logging functional
- [ ] Performance benchmarks run
- [ ] Backup/restore procedures tested

## Future Enhancements

### Planned Features
- **Multi-GPU Support**: Tensor parallelism across multiple GPUs
- **Model Marketplace**: Download different AI models dynamically
- **Energy Optimization**: Adjust performance based on electricity pricing
- **Cross-Platform**: Support for non-NVIDIA GPUs (AMD, Intel)
- **Container Orchestration**: Kubernetes integration for complex deployments

### Research Areas
- **Federated Learning**: Collaborative model training
- **Privacy-Preserving Inference**: Zero-knowledge proofs
- **Edge Computing**: Optimize for variable network conditions
- **Hardware Acceleration**: Utilize TPUs, IPUs, and other accelerators

---

*This technical specification provides a complete blueprint for implementing the Garage AI Boot Stick system. The design prioritizes ease of use, security, and performance while maintaining compatibility with existing gaming PC hardware.*
