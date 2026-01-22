#!/bin/bash

# Script para configurar dominios locales para desarrollo multi-tenant
# Este script agrega entradas al archivo /etc/hosts para que los dominios
# funcionen en tu máquina local.

DOMAIN="pulzifi.local"

echo "Configurando dominios locales para desarrollo ($DOMAIN)..."
echo ""
echo "Este script agregará las siguientes entradas a /etc/hosts:"
echo "  127.0.0.1  $DOMAIN"
echo "  127.0.0.1  jcsoftdev-inc.$DOMAIN"
echo "  127.0.0.1  volkswagen.$DOMAIN"
echo ""
echo "Necesitas permisos de administrador (se te pedirá tu contraseña)"
echo ""

# Función para agregar dominio si no existe
add_domain() {
    local domain=$1
    if grep -q "$domain" /etc/hosts; then
        echo "✓ $domain ya está configurado"
    else
        echo "127.0.0.1  $domain" | sudo tee -a /etc/hosts > /dev/null
        echo "✓ Agregado $domain"
    fi
}

add_domain "$DOMAIN"
add_domain "jcsoftdev-inc.$DOMAIN"
add_domain "volkswagen.$DOMAIN"

echo ""
echo "✅ Configuración completada"
echo ""
echo "Pasos siguientes:"
echo "1. Reinicia tus servidores (backend y frontend)"
echo "2. Accede a: http://$DOMAIN:3000"
echo ""
