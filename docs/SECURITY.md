# 🔒 Security Policy

## Supported Versions

We take security seriously. The following versions of Garage AI are currently supported with security updates:

| Version | Supported          |
| ------- | ------------------ |
| 1.0.x   | :white_check_mark: |
| 0.9.x   | :white_check_mark: |
| < 0.9   | :x:                |

## Reporting a Vulnerability

We appreciate your help in keeping Garage AI secure. If you discover a security vulnerability, please follow these steps:

### 🚨 **Responsible Disclosure Process**

1. **Do not** create public GitHub issues for security vulnerabilities
2. **Email** security@garage.ai with details of the vulnerability
3. **Include** the following information in your report:
   - A clear description of the vulnerability
   - Steps to reproduce the issue
   - Potential impact and severity assessment
   - Your contact information for follow-up
   - Any suggested fixes or mitigations

### 📧 **Contact Information**

- **Email**: security@garage.ai
- **PGP Key**: [Download our PGP public key](https://garage.ai/security/pgp-key.asc)
- **Response Time**: We aim to respond within 48 hours

### 🏆 **Recognition**

We appreciate security researchers who help keep our community safe. With your permission, we'll acknowledge your contribution in our security advisories and Hall of Fame.

## Security Best Practices

### For Node Operators

- **Keep your boot image updated** to the latest version
- **Use strong, unique passwords** for node authentication
- **Monitor your node's activity** regularly
- **Report suspicious behavior** immediately
- **Isolate your Garage AI node** from your personal network if possible

### For Contributors

- **Follow secure coding practices** outlined in our [Contributing Guide](CONTRIBUTING.md)
- **Use dependency scanning** in your development environment
- **Report security issues** in dependencies immediately
- **Implement proper input validation** and sanitization
- **Use secure defaults** for all configurations

## Security Updates

### How We Handle Security Issues

1. **Acknowledgment**: We'll acknowledge receipt of your report within 48 hours
2. **Investigation**: Our security team will investigate and validate the issue
3. **Fix Development**: We'll develop and test a fix
4. **Disclosure**: We'll coordinate disclosure with you
5. **Release**: Security patches will be released as soon as possible

### Security Advisory Format

Our security advisories follow the [Common Vulnerability Scoring System (CVSS)](https://www.first.org/cvss/) and include:

- Vulnerability description
- Affected versions
- Severity score and vector
- Mitigation steps
- Timeline of events
- Credits to reporters

## Known Security Considerations

### Current Limitations

- **Local Processing**: While data stays local, physical access to the machine could compromise security
- **Network Security**: Inter-node communication requires proper firewall configuration
- **Boot Security**: USB boot images should be stored securely to prevent tampering

### Mitigation Strategies

- **Physical Security**: Keep your Garage AI hardware in a secure location
- **Network Segmentation**: Isolate AI workloads from personal/home networks
- **Regular Updates**: Apply security patches promptly
- **Monitoring**: Enable logging and monitoring for suspicious activity

## Bug Bounty Program

We're planning to launch a bug bounty program in Q2 2026. Stay tuned for details on rewards and eligible vulnerabilities.

## Security Hall of Fame

We maintain a [Security Hall of Fame](https://garage.ai/security/hall-of-fame) to recognize security researchers who have helped improve Garage AI's security.

## Contact

For general security questions or concerns, reach out to our community on [Discord](https://discord.gg/garage-ai) or email security@garage.ai.

---

*Security is everyone's responsibility. Thank you for helping keep Garage AI safe!* 🔒
