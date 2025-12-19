#!/usr/bin/env python3
"""
Garage AI Node - Proof of Concept Implementation
This script demonstrates the core functionality of a Garage AI node.
"""

import os
import sys
import json
import time
import psutil
import requests
import platform
import subprocess
from typing import Dict, List, Optional
import logging

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)

class GarageAINode:
    """Garage AI Node implementation"""

    def __init__(self, api_url: str = "https://api.garage.ai"):
        self.api_url = api_url
        self.node_id = self._get_or_create_node_id()
        self.hardware_info = {}
        self.capabilities = {}
        self.token = None

        logger.info(f"Initializing Garage AI Node: {self.node_id}")

    def _get_or_create_node_id(self) -> str:
        """Generate or retrieve persistent node ID"""
        node_id_file = "/opt/garage-ai/node_id"

        if os.path.exists(node_id_file):
            with open(node_id_file, 'r') as f:
                return f.read().strip()

        # Generate new node ID
        import uuid
        node_id = f"garage_{uuid.uuid4().hex[:12]}"

        # Ensure directory exists
        os.makedirs(os.path.dirname(node_id_file), exist_ok=True)

        with open(node_id_file, 'w') as f:
            f.write(node_id)

        return node_id

    def detect_hardware(self) -> Dict:
        """Detect and catalog hardware capabilities"""
        logger.info("Detecting hardware...")

        hardware = {
            'cpu': self._get_cpu_info(),
            'memory': self._get_memory_info(),
            'gpu': self._get_gpu_info(),
            'storage': self._get_storage_info(),
            'network': self._get_network_info(),
            'os': self._get_os_info()
        }

        self.hardware_info = hardware
        return hardware

    def _get_cpu_info(self) -> Dict:
        """Get CPU information"""
        try:
            cpu_info = {
                'cores': psutil.cpu_count(logical=False),
                'threads': psutil.cpu_count(logical=True),
                'frequency': psutil.cpu_freq().max if psutil.cpu_freq() else 0,
                'usage': psutil.cpu_percent(interval=1)
            }
            return cpu_info
        except Exception as e:
            logger.warning(f"Failed to get CPU info: {e}")
            return {'error': str(e)}

    def _get_memory_info(self) -> Dict:
        """Get memory information"""
        try:
            mem = psutil.virtual_memory()
            return {
                'total': mem.total,
                'available': mem.available,
                'used': mem.used,
                'percentage': mem.percent
            }
        except Exception as e:
            logger.warning(f"Failed to get memory info: {e}")
            return {'error': str(e)}

    def _get_gpu_info(self) -> List[Dict]:
        """Get GPU information"""
        gpus = []

        # Try NVIDIA GPUs first
        try:
            result = subprocess.run(['nvidia-smi', '--query-gpu=name,memory.total,memory.free,temperature.gpu',
                                   '--format=csv,noheader,nounits'],
                                  capture_output=True, text=True, timeout=10)

            if result.returncode == 0:
                lines = result.stdout.strip().split('\n')
                for i, line in enumerate(lines):
                    if line.strip():
                        parts = [x.strip() for x in line.split(',')]
                        if len(parts) >= 4:
                            gpu = {
                                'id': i,
                                'name': parts[0],
                                'memory_total': int(parts[1]),
                                'memory_free': int(parts[2]),
                                'temperature': float(parts[3]),
                                'vendor': 'nvidia'
                            }
                            gpus.append(gpu)
        except (subprocess.TimeoutExpired, FileNotFoundError):
            logger.info("NVIDIA GPU detection not available")

        # Try AMD GPUs
        try:
            result = subprocess.run(['rocm-smi', '--showid', '--showproductname', '--showmeminfo', 'vram'],
                                  capture_output=True, text=True, timeout=10)

            if result.returncode == 0 and "GPU" in result.stdout:
                # Parse ROCm output (simplified)
                gpu = {
                    'id': len(gpus),
                    'name': 'AMD GPU',
                    'vendor': 'amd',
                    'memory_total': 0,  # Would need more parsing
                    'temperature': 0
                }
                gpus.append(gpu)
        except (subprocess.TimeoutExpired, FileNotFoundError):
            pass

        return gpus

    def _get_storage_info(self) -> Dict:
        """Get storage information"""
        try:
            disk = psutil.disk_usage('/')
            return {
                'total': disk.total,
                'free': disk.free,
                'used': disk.used,
                'percentage': disk.percent
            }
        except Exception as e:
            logger.warning(f"Failed to get storage info: {e}")
            return {'error': str(e)}

    def _get_network_info(self) -> Dict:
        """Get network information"""
        try:
            net_io = psutil.net_io_counters()
            return {
                'bytes_sent': net_io.bytes_sent,
                'bytes_recv': net_io.bytes_recv,
                'packets_sent': net_io.packets_sent,
                'packets_recv': net_io.packets_recv
            }
        except Exception as e:
            logger.warning(f"Failed to get network info: {e}")
            return {'error': str(e)}

    def _get_os_info(self) -> Dict:
        """Get OS information"""
        return {
            'system': platform.system(),
            'release': platform.release(),
            'version': platform.version(),
            'machine': platform.machine(),
            'processor': platform.processor()
        }

    def test_capabilities(self) -> Dict:
        """Test AI inference capabilities"""
        logger.info("Testing AI capabilities...")

        capabilities = {
            'python_available': self._test_python(),
            'torch_available': self._test_pytorch(),
            'cuda_available': self._test_cuda(),
            'vllm_available': self._test_vllm(),
            'network_connectivity': self._test_network()
        }

        self.capabilities = capabilities
        return capabilities

    def _test_python(self) -> bool:
        """Test Python availability"""
        try:
            import sys
            logger.info(f"Python version: {sys.version}")
            return True
        except:
            return False

    def _test_pytorch(self) -> bool:
        """Test PyTorch availability"""
        try:
            import torch
            logger.info(f"PyTorch version: {torch.__version__}")
            logger.info(f"CUDA available: {torch.cuda.is_available()}")
            if torch.cuda.is_available():
                logger.info(f"CUDA devices: {torch.cuda.device_count()}")
            return True
        except ImportError:
            logger.warning("PyTorch not available")
            return False

    def _test_cuda(self) -> bool:
        """Test CUDA availability"""
        try:
            import torch
            if torch.cuda.is_available():
                device = torch.device("cuda:0")
                tensor = torch.randn(3, 3).to(device)
                logger.info("CUDA test passed")
                return True
            return False
        except:
            return False

    def _test_vllm(self) -> bool:
        """Test vLLM availability"""
        try:
            import vllm
            logger.info(f"vLLM version: {vllm.__version__}")
            return True
        except ImportError:
            logger.warning("vLLM not available")
            return False

    def _test_network(self) -> bool:
        """Test network connectivity"""
        try:
            response = requests.get(f"{self.api_url}/health", timeout=10)
            return response.status_code == 200
        except:
            logger.warning("Network connectivity test failed")
            return False

    def register_node(self) -> bool:
        """Register node with Garage AI network"""
        logger.info("Registering node with Garage AI network...")

        payload = {
            'node_id': self.node_id,
            'hardware': self.hardware_info,
            'capabilities': self.capabilities,
            'location': {'country': 'Sweden', 'region': 'Unknown'},  # Would use geolocation
            'version': '0.1.0'
        }

        try:
            response = requests.post(f"{self.api_url}/nodes/register",
                                   json=payload, timeout=30)

            if response.status_code == 200:
                data = response.json()
                self.token = data.get('token')
                logger.info("Node registered successfully")

                # Save token
                with open('/opt/garage-ai/node_token', 'w') as f:
                    f.write(self.token)

                return True
            else:
                logger.error(f"Registration failed: {response.status_code}")
                return False

        except Exception as e:
            logger.error(f"Registration error: {e}")
            return False

    def start_monitoring(self):
        """Start system monitoring"""
        logger.info("Starting system monitoring...")

        while True:
            try:
                self._report_status()
                time.sleep(60)  # Report every minute
            except KeyboardInterrupt:
                logger.info("Monitoring stopped")
                break
            except Exception as e:
                logger.error(f"Monitoring error: {e}")
                time.sleep(30)

    def _report_status(self):
        """Report current status to central server"""
        if not self.token:
            logger.warning("No token available, skipping status report")
            return

        status = {
            'timestamp': time.time(),
            'cpu_usage': psutil.cpu_percent(),
            'memory_usage': psutil.virtual_memory().percent,
            'gpu_usage': self._get_gpu_usage(),
            'active_sessions': 0,  # Would track active inference sessions
            'earnings_today': 0.0  # Would track token earnings
        }

        headers = {'Authorization': f'Bearer {self.token}'}

        try:
            response = requests.post(f"{self.api_url}/nodes/status",
                                   json=status, headers=headers, timeout=10)

            if response.status_code == 200:
                logger.debug("Status reported successfully")
            else:
                logger.warning(f"Status report failed: {response.status_code}")

        except Exception as e:
            logger.error(f"Status report error: {e}")

    def _get_gpu_usage(self) -> List[Dict]:
        """Get current GPU usage"""
        gpus = []
        try:
            result = subprocess.run(['nvidia-smi', '--query-gpu=utilization.gpu,memory.used,memory.total',
                                   '--format=csv,noheader,nounits'],
                                  capture_output=True, text=True, timeout=5)

            if result.returncode == 0:
                lines = result.stdout.strip().split('\n')
                for i, line in enumerate(lines):
                    if line.strip():
                        parts = [x.strip() for x in line.split(',')]
                        if len(parts) >= 3:
                            gpu = {
                                'id': i,
                                'utilization': int(parts[0]),
                                'memory_used': int(parts[1]),
                                'memory_total': int(parts[2])
                            }
                            gpus.append(gpu)
        except:
            pass

        return gpus

    def run_diagnostics(self):
        """Run full system diagnostics"""
        logger.info("Running full system diagnostics...")

        print("\n" + "="*60)
        print("GARAGE AI NODE DIAGNOSTICS")
        print("="*60)

        # Hardware detection
        hardware = self.detect_hardware()
        print(f"\n🖥️  CPU: {hardware['cpu'].get('cores', 'Unknown')} cores")
        print(f"🧠 Memory: {hardware['memory'].get('total', 0) // (1024**3)} GB total")
        print(f"💾 Storage: {hardware['storage'].get('total', 0) // (1024**3)} GB total")

        # GPU detection
        gpus = hardware.get('gpu', [])
        if gpus:
            print(f"\n🎮 GPUs detected: {len(gpus)}")
            for gpu in gpus:
                print(f"  • {gpu.get('name', 'Unknown')} ({gpu.get('memory_total', 0)} MB VRAM)")
        else:
            print("\n🎮 No GPUs detected")

        # Capability tests
        capabilities = self.test_capabilities()
        print(f"\n🔧 Capabilities:")
        print(f"  • Python: {'✅' if capabilities.get('python_available') else '❌'}")
        print(f"  • PyTorch: {'✅' if capabilities.get('torch_available') else '❌'}")
        print(f"  • CUDA: {'✅' if capabilities.get('cuda_available') else '❌'}")
        print(f"  • vLLM: {'✅' if capabilities.get('vllm_available') else '❌'}")
        print(f"  • Network: {'✅' if capabilities.get('network_connectivity') else '❌'}")

        print(f"\n🆔 Node ID: {self.node_id}")
        print(f"🔗 Network Status: {'Connected' if capabilities.get('network_connectivity') else 'Disconnected'}")

        print("\n" + "="*60)


def main():
    """Main entry point"""
    import argparse

    parser = argparse.ArgumentParser(description='Garage AI Node')
    parser.add_argument('--diagnostics', action='store_true', help='Run diagnostics')
    parser.add_argument('--register', action='store_true', help='Register node')
    parser.add_argument('--monitor', action='store_true', help='Start monitoring')
    parser.add_argument('--api-url', default='https://api.garage.ai', help='API URL')

    args = parser.parse_args()

    node = GarageAINode(args.api_url)

    if args.diagnostics:
        node.run_diagnostics()
    elif args.register:
        if node.register_node():
            print("✅ Node registered successfully!")
        else:
            print("❌ Node registration failed!")
            sys.exit(1)
    elif args.monitor:
        node.start_monitoring()
    else:
        print("Garage AI Node - Proof of Concept")
        print("Usage:")
        print("  python garage_node.py --diagnostics    # Run diagnostics")
        print("  python garage_node.py --register       # Register node")
        print("  python garage_node.py --monitor        # Start monitoring")
        sys.exit(1)


if __name__ == "__main__":
    main()
