#!/bin/bash

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

echo -e "${BLUE}╔══════════════════════════════════╗${NC}"
echo -e "${BLUE}║   JARVIS - Papka Tartiblovchi    ║${NC}"
echo -e "${BLUE}╚══════════════════════════════════╝${NC}"

# ── Papkalar yaratish ──────────────────────────
echo -e "\n${CYAN}[1/5] Papkalar yaratilmoqda...${NC}"
mkdir -p src
mkdir -p certs
mkdir -p assets
mkdir -p installer
mkdir -p release
mkdir -p _backup
echo -e "${GREEN}✓ Papkalar tayyor${NC}"

# ── Source fayllar ────────────────────────────
echo -e "\n${CYAN}[2/5] Source fayllar ko'chirilmoqda...${NC}"
mv main.go        src/       2>/dev/null && echo "  → main.go"
mv go.mod         src/       2>/dev/null && echo "  → go.mod"
mv go.sum         src/       2>/dev/null && echo "  → go.sum"
mv app.manifest   src/       2>/dev/null && echo "  → app.manifest"
mv version.rc     src/       2>/dev/null && echo "  → version.rc"
mv version        src/       2>/dev/null && echo "  → version"
mv *.syso         src/       2>/dev/null && echo "  → *.syso fayllar"
echo -e "${GREEN}✓ Source fayllar tartilandi${NC}"

# ── Sertifikatlar ─────────────────────────────
echo -e "\n${CYAN}[3/5] Sertifikatlar ko'chirilmoqda...${NC}"
# Dublikatlarni backup ga yubor
mv jarvis_cert.pem  _backup/  2>/dev/null && echo "  → jarvis_cert.pem (dublikat → backup)"
mv jarvis_cert.pfx  _backup/  2>/dev/null && echo "  → jarvis_cert.pfx (dublikat → backup)"
mv jarvis_key.pem   _backup/  2>/dev/null && echo "  → jarvis_key.pem  (dublikat → backup)"
# Asosiylarini certs/ ga
mv cert.pem         certs/    2>/dev/null && echo "  → cert.pem"
mv cert.pfx         certs/    2>/dev/null && echo "  → cert.pfx"
mv key.pem          certs/    2>/dev/null && echo "  → key.pem"
echo -e "${GREEN}✓ Sertifikatlar tartilandi${NC}"

# ── Assets ────────────────────────────────────
echo -e "\n${CYAN}[4/5] Assets ko'chirilmoqda...${NC}"
mv app.ico   assets/  2>/dev/null && echo "  → app.ico"
mv logo.ico  assets/  2>/dev/null && echo "  → logo.ico"
mv logo.svg  assets/  2>/dev/null && echo "  → logo.svg"
echo -e "${GREEN}✓ Assets tartilandi${NC}"

# ── Installer fayllar ─────────────────────────
echo -e "\n${CYAN}[5/5] Installer fayllar ko'chirilmoqda...${NC}"
mv installer.iss       installer/  2>/dev/null && echo "  → installer.iss"
mv INSTALL.bat         installer/  2>/dev/null && echo "  → INSTALL.bat"
mv defender_ex