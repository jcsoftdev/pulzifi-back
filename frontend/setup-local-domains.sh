#!/bin/bash

# Script para configurar subdominios locales para desarrollo multi-tenant
# Este script agrega entradas al archivo /etc/hosts para que los subdominios
# funcionen en localhost

echo "Configurando subdominios locales para desarrollo..."
echo ""
echo "Este script agregará las siguientes entradas a /etc/hosts:"
echo "  127.0.0.1  volkswagen.localhost"
echo "  127.0.0.1  otraempresa.localhost"
echo ""
echo "Necesitas permisos de administrador (se te pedirá tu contraseña)"
echo ""

# Verificar si ya existen las entradas
if grep -q "volkswagen.localhost" /etc/hosts; then
    echo "✓ volkswagen.localhost ya está configurado"
else
    echo "127.0.0.1  volkswagen.localhost" | sudo tee -a /etc/hosts > /dev/null
    echo "✓ Agregado volkswagen.localhost"
fi

if grep -q "otraempresa.localhost" /etc/hosts; then
    echo "✓ otraempresa.localhost ya está configurado"
else
    echo "127.0.0.1  otraempresa.localhost" | sudo tee -a /etc/hosts > /dev/null
    echo "✓ Agregado otraempresa.localhost"
fi

echo ""
echo "✅ Configuración completada"
echo ""
echo "Ahora puedes acceder a:"
echo "  - http://volkswagen.localhost:3000"
echo "  - http://otraempresa.localhost:3000"
echo ""
echo "Nota: Si necesitas agregar más tenants, edita /etc/hosts manualmente"
echo "      o ejecuta: echo '127.0.0.1  [tenant].localhost' | sudo tee -a /etc/hosts"
