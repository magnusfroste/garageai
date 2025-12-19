# 🧠 Garage AI Knowledge Base

**Version**: 2.0 | **Last Updated**: December 2025 | **Status**: Active

Denna kunskapsbas innehåller teknisk dokumentation, implementation details och arkitektonisk översikt för Garage AI - ett distribuerat AI-nätverk byggt på idle gaming-datorer i svenska hem.

---

## 📋 Innehållsförteckning

- [Översikt](#översikt)
- [Kärnkoncept](#kärnkoncept)
- [Resursutnyttjande](#resursutnyttjande)
- [Svensk Infrastruktur](#svensk-infrastruktur)
- [Teknisk Implementation](#teknisk-implementation)
- [AI-inferens](#ai-inferens)
- [Community & Utveckling](#community--utveckling)
- [Framtid](#framtid)

---

## 🎯 Översikt

### Vision
**Garage AI**: Demokratisera AI genom decentraliserad infrastruktur. Var med och bygg Europas mest hållbara AI-nätverk - drivet av gaming-datorer och förnybar energi från svenska hem.

### Nyckeltal
- **500,000+ Gaming-PCs** tillgängliga i Sverige idag
- **1.7 Miljoner Garages** potentiella AI-noder
- **98% Fiber-täckning** möjliggör distribuerad databehandling
- **300,000+ Solpaneler** i svenska hem (2024)
- **16-20 timmar/dag** står datorer ofta idle

### Unik Värdeproposition
- **100% Lokal Kontroll** - Ingen data lämnar användarens enhet
- **Resurseffektivitet** - Återanvändning av befintlig hårdvara
- **Miljövänligt** - CO₂-neutral genom förnybar energi
- **Open Source** - MIT-licensierad, community-driven
- **Skalbar** - Från enskild dator till nationellt nätverk

## 🔑 Kärnkoncept

### Resurseffektivitet framför allt
**Garage AI** fokuserar på smart återanvändning av hårdvara som annars står idle. Istället för att gaming-datorer samlar damm när du sover eller arbetar, kan de bidra till AI-inferens och skapa värde.

### Varför detta är unikt
- **Ingen hårdvara-slöseri**: 500k+ gaming-PCs i Sverige står ofta 16-20 timmar/dag
- **Snabb teknikutveckling**: Använd hårdvaran medan den är relevant istället för att den blir föråldrad
- **Bidra med färdigt kluster**: En dator blir en nod i systemet - enkelt och tillgängligt

---

## 🔄 Resursutnyttjande

### Hårdvara som Återanvänds
```
Svenska Gaming-PC Marknaden:
├── 500k+ gaming-datorer finns idag
├── 16-20 timmar/dag står de ofta idle
├── Snabb teknikutveckling kräver utnyttjande
└── Bidra med färdigt kluster - enkelt att komma igång
```

### Teknisk Verklighet
- **RTX 5090**: 120 tokens/sekund kapacitet
- **Idle-förbrukning**: 50-100W när datorn inte används
- **Återanvändning**: Istället för slöseri - skapa AI-värde
- **Skalbarhet**: Från enskild dator till nätverk

---

## 🇸🇪 Svensk Infrastruktur

### Fiber-nätverk
```
PTS-statistik 2024:
├── 98% fiber-täckning i Sverige
├── 123,000+ hushåll med 1 Gbit/s+
├── 10 Gbit/s tillgängligt på fibernät
└── Operatörer: Telia, Telenor, Tele2, lokala nät
```

### Solpaneler & Energi
```
Svenska Solpaneler:
├── 300,000+ hushåll har solpaneler (2024)
├── 25% årlig tillväxt i solcells-installationer
├── 10-12 timmar soltimmar/dag (maj-september)
└── 85% av solenergin produceras av villaägare
```

### Smart Energi-synergi
- **Dagsproduktion**: Solpaneler matar AI-noder dagtid
- **Överskottsenergi**: AI körs på ren solenergi
- **Lokalt producerad**: Ingen nätbelastning eller transmission
- **CO₂-neutral**: 100% förnybar energi för AI

---

## 💻 Teknisk Implementation

### Enkelt att Komma Igång
1. **Ladda ner boot-kit**: `curl -O https://releases.garage.ai/garage-boot-latest.iso`
2. **Skapa USB**: Använd Rufus eller Etcher
3. **Boot från USB**: Välj "Garage AI Inference Mode"
4. **Automatisk setup**: Systemet konfigurerar sig själv

### Tekniska Krav
```yaml
Minimum:
  GPU: RTX 3060 (12GB VRAM)
  RAM: 16GB
  Lagring: 64GB fritt utrymme
  Nätverk: 100 Mbps stabilt

Rekommenderat:
  GPU: RTX 4070+ (24GB+ VRAM)
  RAM: 32GB
  Lagring: 256GB SSD
  Nätverk: 1 Gbps fiber
```

### Distributed AI
- **Ray + vLLM**: Moderna AI-ramverk för distribuerad inferens
- **Docker**: Containerisering för enkel deployment
- **Kubernetes**: Orchestration för skalbarhet
- **Monitoring**: Prometheus + Grafana för övervakning

---

## 🌍 AI-inferens

### Vad är AI-inferens?
AI-inferens är processen där en tränad AI-modell genererar svar på frågor eller skapar innehåll. Istället för att skicka data till molnet (OpenAI, Google) körs allt lokalt på din hårdvara.

### Säkerhetsjämförelse
```
Moln-AI (OpenAI, Google):
❌ Data skickas till externa servrar
❌ Klartext-bearbetning i TEE
❌ Loggning för träning/säkerhet
❌ Kostnad per förfrågan

Garage AI:
✅ 100% lokal bearbetning
✅ Ingen data lämnar din enhet
✅ Ingen loggning eller analytics
✅ Endast el-kostnad
```

---

## 👥 Community & Utveckling

### Open Source & Community-driven
- **MIT-licensierad**: Fri användning och modifiering
- **GitHub-community**: Diskussioner, issues, bidrag
- **Svenskt fokus**: Utvecklat för svenska förhållanden
- **Samarbeten**: Autoversio & Liteit som partners

### Komma Igång för Bidragare
1. **Klona repo**: `git clone https://github.com/magnusfroste/garageai`
2. **Installera beroenden**: `npm install`
3. **Starta utveckling**: `npm run dev`
4. **Bidra**: Dokumentation, kod, tester

### Framtida Utveckling
- **Fler GPU-modeller**: Stöd för AMD, Intel
- **Edge computing**: Offline AI-capabilities
- **Federated learning**: Samarbete mellan noder
- **API-utveckling**: Lokala AI-tjänster

---

## 🎯 Framtid

### Nästa Steg
- **Prototyping**: Testa distributed inferens med Ray + vLLM
- **Community-building**: Skapa onboarding och dokumentation
- **Partner-integrationer**: Samarbeta med Autoversio & Liteit
- **Skalning**: Från proof-of-concept till nationellt nätverk

### Vision 2026
- **10,000+ aktiva noder** i svenska hem
- **1.2M tokens/sekund** total kapacitet
- **50,000+ samtidiga användare** kan betjänas
- **CO₂-neutral AI** genom solenergi-integration

---

## 📚 Referenser

### Officiella Källor
- **SCB**: Statistik om svenska hushåll och boende
- **PTS**: Fiber-infrastruktur och täckning
- **Energimyndigheten**: Solenergi och förnybar energi

### Teknisk Dokumentation
- **NVIDIA**: GPU-prestanda och drivrutiner
- **Ray**: Distributed computing framework
- **vLLM**: AI-inference optimization
- **Docker**: Containerisering och deployment

---

## 🚀 Sammanfattning

Garage AI är ett praktiskt svar på AI:s centraliseringsproblem. Genom att återanvända idle gaming-hårdvara skapar vi ett distribuerat AI-nätverk som:

- **Skyddar integritet**: 100% lokal databehandling
- **Optimerar resurser**: Ingen hårdvara-slöseri
- **Främjar suveränitet**: Svensk infrastruktur, svenskt värde
- **Skyddar miljö**: CO₂-neutral genom solenergi
- **Bygger community**: Open source, tillgängligt för alla

**Redo att börja?** → [GET_STARTED.md](GET_STARTED.md)
