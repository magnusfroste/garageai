# 🤝 Contributing to Garage AI

Welcome! We're thrilled you're interested in contributing to Garage AI. This document provides guidelines and information for contributors.

## 📋 Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [How to Contribute](#how-to-contribute)
- [Development Workflow](#development-workflow)
- [Project Structure](#project-structure)
- [Testing](#testing)
- [Documentation](#documentation)
- [Community](#community)

## 📜 Code of Conduct

This project adheres to a code of conduct to ensure a welcoming environment for all contributors. By participating, you agree to:

- **Be respectful** and inclusive in all interactions
- **Focus on constructive feedback** and collaborative problem-solving
- **Accept responsibility** for mistakes and learn from them
- **Show empathy** towards other contributors and users
- **Help create a positive community** environment

## 🚀 Getting Started

### Prerequisites

- **Python 3.9+** with pip and venv
- **Node.js 18+** and npm/yarn
- **Docker** and Docker Compose
- **Git** for version control
- **GPU** (NVIDIA RTX 30/40 series recommended) for AI development

### Quick Setup

```bash
# Clone the repository
git clone https://github.com/garage-ai/platform.git
cd platform

# Set up Python environment
python -m venv venv
source venv/bin/activate  # On Windows: venv\Scripts\activate
pip install -r requirements.txt

# Set up Node.js environment
npm install

# Start development environment
docker-compose up -d
npm run dev
```

## 🤝 How to Contribute

### Types of Contributions

- **🐛 Bug Reports**: Report bugs via [GitHub Issues](https://github.com/garage-ai/platform/issues)
- **💡 Feature Requests**: Suggest new features or improvements
- **🔧 Code Contributions**: Submit pull requests with fixes or new features
- **📚 Documentation**: Improve documentation, tutorials, or examples
- **🧪 Testing**: Write or improve tests
- **🎨 Design**: UI/UX improvements or design contributions
- **🌐 Translation**: Help translate the project to other languages

### Finding Issues to Work On

- Check [GitHub Issues](https://github.com/garage-ai/platform/issues) for open tasks
- Look for issues labeled `good first issue` or `help wanted`
- Join our [Discord](https://discord.gg/garage-ai) to discuss potential contributions
- Check the [Project Board](https://github.com/garage-ai/platform/projects) for planned work

## 🔄 Development Workflow

### 1. Fork and Clone

```bash
# Fork the repository on GitHub, then:
git clone https://github.com/YOUR_USERNAME/platform.git
cd platform
git remote add upstream https://github.com/garage-ai/platform.git
```

### 2. Create a Branch

```bash
# Create a feature branch
git checkout -b feature/your-feature-name

# Or create a bug fix branch
git checkout -b fix/issue-number-description
```

### 3. Make Changes

- Write clear, concise commit messages
- Follow the existing code style
- Add tests for new functionality
- Update documentation as needed

### 4. Test Your Changes

```bash
# Run tests
python -m pytest

# Run linting
flake8 src/
black --check src/

# Build and test the application
npm run build
npm run test
```

### 5. Submit a Pull Request

```bash
# Push your branch
git push origin feature/your-feature-name

# Create a Pull Request on GitHub with:
# - Clear title describing the change
# - Detailed description of what was changed and why
# - Reference to any related issues
# - Screenshots/videos for UI changes
```

## 🏗️ Project Structure

```
garage-ai-platform/
├── docs/                    # Documentation
│   ├── README.md           # Main documentation
│   ├── CONTRIBUTING.md     # This file
│   └── ...
├── src/                     # Source code
│   ├── garage_ai/          # Core Python package
│   ├── web/               # Web application
│   └── cli/               # Command-line tools
├── tests/                   # Test suites
├── docker/                  # Docker configurations
├── scripts/                 # Utility scripts
├── requirements.txt         # Python dependencies
├── package.json            # Node.js dependencies
├── docker-compose.yml      # Development environment
└── LICENSE                 # MIT License
```

### Key Components

- **`garage_ai/`**: Core distributed AI functionality
- **`web/`**: React-based web interface
- **`cli/`**: Command-line tools for node management
- **`docs/`**: Comprehensive documentation

## 🧪 Testing

### Running Tests

```bash
# Run all tests
python -m pytest

# Run specific test file
python -m pytest tests/test_node.py

# Run with coverage
python -m pytest --cov=src --cov-report=html

# Run integration tests
python -m pytest tests/integration/
```

### Writing Tests

```python
import pytest
from garage_ai.node import GarageAINode

class TestGarageAINode:
    def test_initialization(self):
        node = GarageAINode()
        assert node.node_id is not None
        assert len(node.node_id) > 0

    def test_hardware_detection(self):
        node = GarageAINode()
        hardware = node.detect_hardware()

        assert 'cpu' in hardware
        assert 'memory' in hardware
        assert 'gpu' in hardware
```

### Test Coverage Requirements

- **Unit Tests**: Minimum 80% coverage for new code
- **Integration Tests**: Required for API changes
- **End-to-End Tests**: Required for user-facing features

## 📚 Documentation

### Documentation Standards

- Use **Markdown** for all documentation
- Include code examples where relevant
- Keep documentation up-to-date with code changes
- Use clear, concise language accessible to beginners

### Updating Documentation

```bash
# Preview documentation locally
npm run docs:serve

# Build documentation
npm run docs:build

# Deploy documentation (CI/CD)
npm run docs:deploy
```

### Documentation Structure

- **README.md**: Project overview and quick start
- **docs/**: Detailed technical documentation
- **API Reference**: Auto-generated from code comments
- **Tutorials**: Step-by-step guides for common tasks

## 🎯 Development Areas

### High Priority

- **🔌 Operator onboarding**: Improve the connect script (`scripts/garageai-connect.sh`) and runtime support (Ollama, LM Studio, llama.cpp, vLLM)
- **🔧 Hardware Detection**: Better GPU and hardware compatibility
- **🌐 Network Optimization**: Reduce latency in distributed inference
- **🔒 Security**: Enhance node authentication and data protection

### Medium Priority

- **📊 Monitoring Dashboard**: Real-time node performance visualization
- **🤖 Model Marketplace**: Easy model deployment and management
- **⚡ Performance Optimization**: Faster inference through caching and optimization
- **🔄 Auto-Updates**: Seamless software updates for nodes

### Future Enhancements

- **🌍 Multi-Language Support**: Localization and internationalization
- **📱 Mobile Apps**: iOS/Android companion apps
- **🔗 API Integrations**: Third-party service integrations
- **🎮 Gaming Integrations**: Direct integration with gaming platforms

## 💬 Community

### Communication Channels

- **💬 Discord**: [Garage AI Community](https://discord.gg/garage-ai) - Real-time discussions
- **🐛 GitHub Issues**: Bug reports and feature requests
- **📧 Email**: team@garage.ai for private communications
- **🐦 Twitter**: [@garage_ai](https://twitter.com/garage_ai) for updates

### Community Guidelines

- **Be welcoming** to newcomers and diverse perspectives
- **Keep discussions on topic** and constructive
- **Respect different skill levels** and backgrounds
- **Help others learn** and share knowledge
- **Give credit** where credit is due

### Recognition

Contributors are recognized through:
- **GitHub Contributors**: Listed in repository contributors
- **Changelog**: Mentioned in release notes
- **Community Spotlight**: Featured in community updates
- **Token Rewards**: GAI tokens for significant contributions

## 🙏 Acknowledgments

Thank you for contributing to Garage AI! Your efforts help democratize AI and build a more distributed, privacy-preserving AI ecosystem.

**Questions?** Reach out on [Discord](https://discord.gg/garage-ai) or create a [GitHub Discussion](https://github.com/garage-ai/platform/discussions).

---

*Built with ❤️ by the open source community*
