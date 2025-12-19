# 🚀 Komma Igång med Garage AI

Välkommen till Garage AI! Den här guiden hjälper dig att snabbt komma igång, oavsett om du är gaming-entusiast eller teknisk bidragare.

## 📋 Innehåll

- [Snabbstart för Gaming-entusiaster](#snabbstart-för-gaming-entusiaster)
- [Teknisk Installation](#teknisk-installation)
- [Community & Support](#community--support)
- [Bidra till Projektet](#bidra-till-projektet)

---

## 🎮 Snabbstart för Gaming-entusiaster

### Vad är Garage AI?
Garage AI låter dig använda din gaming-dator för att bidra till ett globalt AI-nätverk när du inte spelar. Du tjänar tokens genom att "låna ut" din GPU-kraft till AI-beräkningar.

### Steg-för-steg Guide

#### 1. **Kolla Kompatibilitet**
   - **GPU**: RTX 3060 eller bättre (NVIDIA)
   - **RAM**: Minst 16GB
   - **Lagring**: 64GB ledig plats för USB
   - **Internet**: Stabil 100Mbps+ anslutning

#### 2. **Ladda Ner Boot Kit**
   ```bash
   # Ladda ner från vår hemsida
   curl -O https://releases.garage.ai/garage-boot-latest.iso
   ```

#### 3. **Skapa Bootbar USB**
   - Använd Rufus (Windows) eller Etcher (macOS/Linux)
   - Välj din USB-enhet (minst 64GB)
   - Flasha ISO-filen till USB:en

#### 4. **Boot från USB**
   - Starta om datorn
   - Tryck F12/F10/Delete för boot-menu (beroende på motherboard)
   - Välj USB-enheten
   - Välj "Garage AI Inference Mode"

#### 5. **Första Körningen**
   - Systemet konfigurerar sig automatiskt
   - Registrerar din nod i nätverket
   - Börjar tjäna GAI-tokens!

#### 6. **Återvänd till Windows**
   - Stäng av datorn
   - Ta bort USB eller välj Windows i boot-menu
   - Fortsätt spela som vanligt!

### Vad Händer Nu?

- **Automatisk körning**: Datorn bidrar när du inte använder den
- **Token-belöningar**: Tjäna GAI-tokens för varje AI-uppgift
- **Community-rang**: Klättra på leaderboards
- **Inga kostnader**: Endast din elförbrukning

### Felsökning

#### "GPU inte hittad"?
```bash
# Kontrollera NVIDIA-drivers
nvidia-smi
```

#### "Ingen nätverksanslutning"?
```bash
# Testa internet
ping 8.8.8.8
```

#### Behöver hjälp?
- **[Discord Community](https://discord.gg/garage-ai)** - Fråga andra användare
- **E-post**: support@garage.ai

---

## 💻 Teknisk Installation

För utvecklare och avancerade användare som vill modifiera eller bidra.

### Förutsättningar

- **Ubuntu 22.04+** eller kompatibel Linux-distribution
- **Python 3.9+**
- **NVIDIA GPU** med CUDA 11.8+
- **Git**

### Klona Repository

```bash
git clone https://github.com/garage-ai/platform.git
cd platform
```

### Snabbinstallation

```bash
# Installera alla beroenden
./scripts/setup.sh

# Konfigurera GPU
./scripts/setup_gpu.sh

# Starta utvecklingsserver
npm run dev
```

### Manuell Installation

#### 1. Python Environment
```bash
# Skapa virtuell miljö
python3 -m venv garage-env
source garage-env/bin/activate

# Installera beroenden
pip install -r requirements.txt
```

#### 2. AI Libraries
```bash
# Installera PyTorch med CUDA
pip install torch torchvision torchaudio --index-url https://download.pytorch.org/whl/cu118

# Installera övriga AI-bibliotek
pip install transformers accelerate vllm ray[all] fastapi uvicorn
```

#### 3. GPU Setup
```bash
# Verifiera CUDA installation
nvidia-smi
python3 -c "import torch; print(f'CUDA: {torch.cuda.is_available()}')"

# Installera vLLM
pip install vllm
```

### Testa Installationen

```bash
# Kör diagnostik
python3 garage_node.py --diagnostics

# Registrera test-nod
python3 garage_node.py --register --api-url http://localhost:8000

# Starta övervakning
python3 garage_node.py --monitor
```

### Docker Installation (Alternativ)

```bash
# Bygg container
docker build -t garage-ai .

# Kör med GPU-stöd
docker run --gpus all -p 8000:8000 garage-ai
```

---

## 🌍 Community & Support

### Kommunikationskanaler

- **💬 [Discord](https://discord.gg/garage-ai)** - Realtidsdiskussioner, hjälp, events
- **🐛 [GitHub Issues](https://github.com/garage-ai/platform/issues)** - Bug-rapporter, feature-förfrågningar
- **📧 [E-post](mailto:team@garage.ai)** - Privat kommunikation
- **🐦 [Twitter](https://twitter.com/garage_ai)** - Uppdateringar och nyheter

### Community Events

- **Månatliga Workshops**: Tekniska tutorials och Q&A
- **Gaming-turneringar**: Tävla och tjäna extra tokens
- **Globala Meetups**: Träffa andra contributors IRL

### Support Nivåer

| Typ | Svarstid | Kanaler |
|-----|----------|---------|
| **Allmän Frågor** | 24h | Discord, GitHub |
| **Teknisk Support** | 12h | Discord, Issues |
| **Bug-rapporter** | 6h | GitHub Issues |
| **Säkerhet** | 2h | security@garage.ai |

---

## 🤝 Bidra till Projektet

### Varför Bidra?

- **Påverka framtiden**: Forma hur distribuerad AI fungerar
- **GAI Token-belöningar**: Tjäna för kod-bidrag
- **Community-erkännande**: Bli featured i releases
- **Lärande**: Arbeta med cutting-edge AI-teknik

### Enkla Första Bidrag

1. **🐛 Rapportera Bugs**: Beskriv tydligt vad som gick fel
2. **💡 Förslag**: Dela idéer för förbättringar
3. **📚 Dokumentation**: Förbättra guider och tutorials
4. **🧪 Testning**: Hjälp testa nya features

### Tekniska Bidrag

#### Kod-standarder
```python
# Följ PEP 8
# Använd type hints
def process_inference(input_data: Dict[str, Any]) -> Dict[str, Any]:
    """Process AI inference request.

    Args:
        input_data: Input data dictionary

    Returns:
        Processed result dictionary
    """
    pass
```

#### Pull Request Process
1. **Fork** repository
2. **Skapa branch**: `feature/din-feature` eller `fix/issue-nummer`
3. **Commit**-meddelanden: "Add: New feature" eller "Fix: Issue description"
4. **Test**: Alla tester måste passera
5. **PR**: Beskrivning, screenshots, test-instruktioner

#### Review Process
- **Auto-checks**: Linting, testing, security scans
- **Peer review**: Minst en maintainer review
- **CI/CD**: Automatisk testning och deployment

### Bidrag Områden

#### 🔥 Hög Prioritet
- **Boot Image Optimization**: Snabbare boot-tider
- **GPU Compatibility**: Stöd för fler GPU-modeller
- **Network Optimization**: Bättre latenshantering

#### 📈 Medel Prioritet
- **Mobile App**: iOS/Android companion
- **Web Dashboard**: Förbättrad nod-hantering
- **Model Marketplace**: Dynamisk model-loading

#### 🚀 Framtid
- **Federated Learning**: Collaborative training
- **Multi-Modal AI**: Bild/text/video stöd
- **Edge Computing**: Offline capabilities

### Erkännande

Contributors nämns i:
- **Changelog**: Release notes
- **Contributors-fil**: GitHub contributors lista
- **Community Spotlight**: Discord announcements
- **Token Rewards**: GAI tokens för kvalificerade bidrag

---

## 📚 Nästa Steg

### För Användare
1. **Prova det själv**: Ladda ner och testa boot-USB
2. **Gå med i Discord**: Träffa andra användare
3. **Dela feedback**: Hjälp förbättra upplevelsen

### För Bidragare
1. **Läs CONTRIBUTING.md**: Detaljerade riktlinjer
2. **Utforska Issues**: Hitta något att arbeta med
3. **Starta enkelt**: Dokumentation eller små bug-fixes
4. **Fråga i Discord**: Få hjälp från communityn

### För Teknikintresserade
1. **Studera arkitekturen**: Läs tekniska specifikationer
2. **Experimentera**: Testa proof-of-concept koden
3. **Bygg vidare**: Skapa nya features eller förbättringar

---

*Frågor? Besök vår [Discord](https://discord.gg/garage-ai) eller skapa ett [GitHub Discussion](https://github.com/garage-ai/platform/discussions)!*

*Byggd med ❤️ av open source communityn*
