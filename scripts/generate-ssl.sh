#!/bin/bash

# ============================================================
# SSL Certificate Generation Script
# Generates self-signed certificates for development
# or Let's Encrypt certificates for production
# ============================================================

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SSL_DIR="$SCRIPT_DIR/ssl"
ENVIRONMENT=${1:-development}

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}SSL Certificate Generation${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""

# Create SSL directory
mkdir -p "$SSL_DIR"

if [ "$ENVIRONMENT" == "production" ]; then
    echo -e "${YELLOW}Production Mode: Using Let's Encrypt${NC}"
    echo ""
    
    # Check for certbot
    if ! command -v certbot &> /dev/null; then
        echo -e "${RED}Error: certbot is not installed${NC}"
        echo "Install with: sudo apt-get install certbot python3-certbot-nginx"
        exit 1
    fi
    
    read -p "Enter your domain name: " DOMAIN
    read -p "Enter your email for Let's Encrypt: " EMAIL
    
    echo ""
    echo -e "${YELLOW}Generating Let's Encrypt certificate for $DOMAIN...${NC}"
    
    certbot certonly \
        --standalone \
        --non-interactive \
        --agree-tos \
        --email "$EMAIL" \
        --domain "$DOMAIN" \
        --cert-path "$SSL_DIR/cert.pem" \
        --key-path "$SSL_DIR/key.pem"
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✓ Certificate generated successfully${NC}"
        echo -e "${GREEN}Certificate: $SSL_DIR/cert.pem${NC}"
        echo -e "${GREEN}Key: $SSL_DIR/key.pem${NC}"
    else
        echo -e "${RED}Failed to generate certificate${NC}"
        exit 1
    fi

else
    echo -e "${YELLOW}Development Mode: Using self-signed certificate${NC}"
    echo ""
    
    # Check if certificates already exist
    if [ -f "$SSL_DIR/cert.pem" ] && [ -f "$SSL_DIR/key.pem" ]; then
        read -p "Certificates already exist. Regenerate? (y/n) " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            echo -e "${GREEN}Using existing certificates${NC}"
            exit 0
        fi
    fi
    
    echo -e "${YELLOW}Generating self-signed certificate for development...${NC}"
    echo ""
    
    # Generate self-signed certificate for 365 days
    openssl req \
        -x509 \
        -newkey rsa:4096 \
        -nodes \
        -out "$SSL_DIR/cert.pem" \
        -keyout "$SSL_DIR/key.pem" \
        -days 365 \
        -subj "/C=US/ST=State/L=City/O=Organization/OU=Department/CN=localhost"
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✓ Self-signed certificate generated successfully${NC}"
        echo -e "${GREEN}Certificate: $SSL_DIR/cert.pem${NC}"
        echo -e "${GREEN}Key: $SSL_DIR/key.pem${NC}"
        echo ""
        echo -e "${YELLOW}Note: This certificate is self-signed and will trigger browser warnings.${NC}"
        echo -e "${YELLOW}This is normal for development. Use Let's Encrypt for production.${NC}"
    else
        echo -e "${RED}Failed to generate self-signed certificate${NC}"
        exit 1
    fi
fi

echo ""
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}✓ SSL Setup Complete${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""

# Check certificate details
if [ -f "$SSL_DIR/cert.pem" ]; then
    echo -e "${YELLOW}Certificate Details:${NC}"
    openssl x509 -in "$SSL_DIR/cert.pem" -text -noout | grep -E "Subject:|Issuer:|Not Before|Not After"
    echo ""
fi

echo -e "${GREEN}Ready to start Docker services with: make docker-up${NC}"
