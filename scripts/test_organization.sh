#!/bin/bash

# Organization Module Testing Script
# Usage: ./test_organization.sh [rest|grpc|all]

set -e

COLOR_GREEN='\033[0;32m'
COLOR_RED='\033[0;31m'
COLOR_YELLOW='\033[1;33m'
COLOR_BLUE='\033[0;34m'
NC='\033[0m' # No Color

REST_URL="http://localhost:8080"
GRPC_URL="localhost:9000"

# Mock JWT token for testing (replace with real token if needed)
JWT_TOKEN="Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiI3NzBlODQwMC1lMjliLTQxZDQtYTcxNi00NDY2NTU0NDAwMDAifQ.fake"

# UUIDs for testing
OWNER_USER_ID="770e8400-e29b-41d4-a716-446655440000"

echo -e "${COLOR_BLUE}═══════════════════════════════════════════════════════════${NC}"
echo -e "${COLOR_BLUE}Organization Module Testing Script${NC}"
echo -e "${COLOR_BLUE}═══════════════════════════════════════════════════════════${NC}\n"

# Test REST API
test_rest() {
    echo -e "${COLOR_YELLOW}→ Testing REST API${NC}\n"

    # Health check
    echo -e "${COLOR_BLUE}1. Health Check${NC}"
    HEALTH=$(curl -s -w "\n%{http_code}" http://localhost:8080/health)
    HTTP_CODE=$(echo "$HEALTH" | tail -1)
    BODY=$(echo "$HEALTH" | head -1)
    
    if [ "$HTTP_CODE" = "200" ]; then
        echo -e "${COLOR_GREEN}✓ Health check passed${NC}"
        echo "  Response: $BODY"
    else
        echo -e "${COLOR_RED}✗ Health check failed (HTTP $HTTP_CODE)${NC}"
        return 1
    fi
    echo ""

    # Create organization
    echo -e "${COLOR_BLUE}2. Create Organization${NC}"
    SUBDOMAIN="test-org-$(date +%s)"
    
    CREATE_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$REST_URL/api/organizations" \
        -H "Content-Type: application/json" \
        -H "Authorization: $JWT_TOKEN" \
        -d "{
            \"name\": \"Test Organization\",
            \"subdomain\": \"$SUBDOMAIN\"
        }")
    
    HTTP_CODE=$(echo "$CREATE_RESPONSE" | tail -1)
    BODY=$(echo "$CREATE_RESPONSE" | head -1)
    
    if [ "$HTTP_CODE" = "201" ]; then
        echo -e "${COLOR_GREEN}✓ Organization created${NC}"
        echo "  Response: $BODY"
        
        # Extract organization ID
        ORG_ID=$(echo "$BODY" | grep -o '"id":"[^"]*' | head -1 | sed 's/"id":"//')
        echo "  Organization ID: $ORG_ID"
    else
        echo -e "${COLOR_RED}✗ Failed to create organization (HTTP $HTTP_CODE)${NC}"
        echo "  Response: $BODY"
        return 1
    fi
    echo ""

    # Get organization
    if [ -n "$ORG_ID" ]; then
        echo -e "${COLOR_BLUE}3. Get Organization${NC}"
        GET_RESPONSE=$(curl -s -w "\n%{http_code}" -X GET "$REST_URL/api/organizations/$ORG_ID" \
            -H "Authorization: $JWT_TOKEN")
        
        HTTP_CODE=$(echo "$GET_RESPONSE" | tail -1)
        BODY=$(echo "$GET_RESPONSE" | head -1)
        
        if [ "$HTTP_CODE" = "200" ]; then
            echo -e "${COLOR_GREEN}✓ Organization retrieved${NC}"
            echo "  Response: $BODY"
        else
            echo -e "${COLOR_RED}✗ Failed to get organization (HTTP $HTTP_CODE)${NC}"
            echo "  Response: $BODY"
            return 1
        fi
        echo ""
    fi

    # Test invalid organization ID
    echo -e "${COLOR_BLUE}4. Get Non-existent Organization (expect 404)${NC}"
    INVALID_ID="550e8400-e29b-41d4-a716-446655440099"
    GET_RESPONSE=$(curl -s -w "\n%{http_code}" -X GET "$REST_URL/api/organizations/$INVALID_ID" \
        -H "Authorization: $JWT_TOKEN")
    
    HTTP_CODE=$(echo "$GET_RESPONSE" | tail -1)
    
    if [ "$HTTP_CODE" = "404" ]; then
        echo -e "${COLOR_GREEN}✓ Correctly returned 404 for non-existent organization${NC}"
    else
        echo -e "${COLOR_YELLOW}⚠ Expected 404, got $HTTP_CODE${NC}"
    fi
    echo ""

    # Test duplicate subdomain
    echo -e "${COLOR_BLUE}5. Create with Duplicate Subdomain (expect 409)${NC}"
    DUP_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$REST_URL/api/organizations" \
        -H "Content-Type: application/json" \
        -H "Authorization: $JWT_TOKEN" \
        -d "{
            \"name\": \"Duplicate Org\",
            \"subdomain\": \"$SUBDOMAIN\"
        }")
    
    HTTP_CODE=$(echo "$DUP_RESPONSE" | tail -1)
    
    if [ "$HTTP_CODE" = "409" ]; then
        echo -e "${COLOR_GREEN}✓ Correctly rejected duplicate subdomain with 409${NC}"
    else
        echo -e "${COLOR_YELLOW}⚠ Expected 409, got $HTTP_CODE${NC}"
    fi
    echo ""
}

# Test gRPC API
test_grpc() {
    echo -e "${COLOR_YELLOW}→ Testing gRPC API${NC}\n"

    # Check if grpcurl is installed
    if ! command -v grpcurl &> /dev/null; then
        echo -e "${COLOR_RED}✗ grpcurl not found. Install with: brew install grpcurl${NC}"
        return 1
    fi

    # List services
    echo -e "${COLOR_BLUE}1. List gRPC Services${NC}"
    SERVICES=$(grpcurl -plaintext $GRPC_URL list 2>&1)
    
    if echo "$SERVICES" | grep -q "OrganizationService"; then
        echo -e "${COLOR_GREEN}✓ OrganizationService found${NC}"
        echo "  Services:"
        echo "$SERVICES" | sed 's/^/    /'
    else
        echo -e "${COLOR_RED}✗ OrganizationService not found${NC}"
        echo "  Output: $SERVICES"
        return 1
    fi
    echo ""

    # Describe service
    echo -e "${COLOR_BLUE}2. Describe OrganizationService${NC}"
    DESCRIBE=$(grpcurl -plaintext $GRPC_URL describe organization.OrganizationService 2>&1)
    echo "  Service methods:"
    echo "$DESCRIBE" | sed 's/^/    /'
    echo ""

    # Create organization via gRPC
    echo -e "${COLOR_BLUE}3. Create Organization via gRPC${NC}"
    GRPC_SUBDOMAIN="grpc-test-$(date +%s)"
    
    CREATE_RESPONSE=$(grpcurl -plaintext \
        -d "{
            \"name\": \"gRPC Test Org\",
            \"subdomain\": \"$GRPC_SUBDOMAIN\",
            \"owner_user_id\": \"$OWNER_USER_ID\"
        }" \
        $GRPC_URL organization.OrganizationService/CreateOrganization 2>&1)
    
    if echo "$CREATE_RESPONSE" | grep -q "id"; then
        echo -e "${COLOR_GREEN}✓ Organization created via gRPC${NC}"
        echo "  Response: $CREATE_RESPONSE"
        
        # Extract organization ID
        GRPC_ORG_ID=$(echo "$CREATE_RESPONSE" | grep -o '"id": "[^"]*' | head -1 | sed 's/"id": "//')
        echo "  Organization ID: $GRPC_ORG_ID"
    else
        echo -e "${COLOR_RED}✗ Failed to create organization via gRPC${NC}"
        echo "  Response: $CREATE_RESPONSE"
        return 1
    fi
    echo ""

    # Get organization via gRPC
    if [ -n "$GRPC_ORG_ID" ]; then
        echo -e "${COLOR_BLUE}4. Get Organization via gRPC${NC}"
        GET_RESPONSE=$(grpcurl -plaintext \
            -d "{\"id\": \"$GRPC_ORG_ID\"}" \
            $GRPC_URL organization.OrganizationService/GetOrganization 2>&1)
        
        if echo "$GET_RESPONSE" | grep -q "name"; then
            echo -e "${COLOR_GREEN}✓ Organization retrieved via gRPC${NC}"
            echo "  Response: $GET_RESPONSE"
        else
            echo -e "${COLOR_RED}✗ Failed to get organization via gRPC${NC}"
            echo "  Response: $GET_RESPONSE"
            return 1
        fi
        echo ""
    fi
}

# Main execution
case "${1:-all}" in
    rest)
        test_rest
        ;;
    grpc)
        test_grpc
        ;;
    all)
        test_rest
        test_grpc
        ;;
    *)
        echo "Usage: $0 [rest|grpc|all]"
        exit 1
        ;;
esac

echo -e "${COLOR_BLUE}═══════════════════════════════════════════════════════════${NC}"
echo -e "${COLOR_GREEN}Testing completed!${NC}"
echo -e "${COLOR_BLUE}═══════════════════════════════════════════════════════════${NC}"
