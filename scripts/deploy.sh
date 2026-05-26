#!/bin/bash
set -e

# Deploy script for img-fwd demo site
# Deploys both proxy and nginx apps to Fly.io

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(dirname "$SCRIPT_DIR")"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}=== img-fwd Deployment Script ===${NC}"
echo ""

# Check if flyctl is installed
if ! command -v flyctl &> /dev/null; then
    echo -e "${RED}Error: flyctl is not installed${NC}"
    echo "Install it with:"
    echo "  macOS: brew install flyctl"
    echo "  Linux: curl -L https://fly.io/install.sh | sh"
    exit 1
fi

# Check if authenticated
if ! flyctl auth whoami &> /dev/null; then
    echo -e "${RED}Error: Not authenticated with Fly.io${NC}"
    echo "Run: flyctl auth login"
    exit 1
fi

echo -e "${GREEN}✓ flyctl is installed and authenticated${NC}"
echo ""

# Parse arguments
DEPLOY_PROXY=false
DEPLOY_NGINX=false
DEPLOY_ALL=false

if [ $# -eq 0 ]; then
    DEPLOY_ALL=true
else
    for arg in "$@"; do
        case $arg in
            --proxy)
                DEPLOY_PROXY=true
                ;;
            --nginx)
                DEPLOY_NGINX=true
                ;;
            --all)
                DEPLOY_ALL=true
                ;;
            --help)
                echo "Usage: $0 [OPTIONS]"
                echo ""
                echo "Options:"
                echo "  --proxy   Deploy only the proxy app (img-fwd)"
                echo "  --nginx   Deploy only the nginx app (img-fwd-demo-static)"
                echo "  --all     Deploy both apps (default)"
                echo "  --help    Show this help message"
                exit 0
                ;;
            *)
                echo -e "${RED}Unknown option: $arg${NC}"
                echo "Use --help for usage information"
                exit 1
                ;;
        esac
    done
fi

if [ "$DEPLOY_ALL" = true ]; then
    DEPLOY_PROXY=true
    DEPLOY_NGINX=true
fi

# Deploy proxy
if [ "$DEPLOY_PROXY" = true ]; then
    echo -e "${YELLOW}Deploying proxy (img-fwd)...${NC}"
    cd "$ROOT_DIR"
    flyctl deploy -a img-fwd
    echo -e "${GREEN}✓ Proxy deployed successfully${NC}"
    echo ""
fi

# Deploy nginx
if [ "$DEPLOY_NGINX" = true ]; then
    echo -e "${YELLOW}Deploying nginx (img-fwd-demo-static)...${NC}"
    cd "$ROOT_DIR/demo"
    flyctl deploy -a img-fwd-demo-static
    echo -e "${GREEN}✓ Nginx deployed successfully${NC}"
    echo ""
fi

# Verify deployment
echo -e "${YELLOW}Verifying deployment...${NC}"
echo ""

echo "Proxy status:"
flyctl status -a img-fwd | grep -E "(Status|Version)" || echo "  Unable to get status"
echo ""

echo "Nginx status:"
flyctl status -a img-fwd-demo-static | grep -E "(Status|Version)" || echo "  Unable to get status"
echo ""

echo -e "${GREEN}=== Deployment Complete ===${NC}"
echo ""
echo "Test the site: https://img-fwd.driedel.dev/"
echo ""
echo "Useful commands:"
echo "  View proxy logs:  flyctl logs -a img-fwd"
echo "  View nginx logs:  flyctl logs -a img-fwd-demo-static"
echo "  SSH to proxy:     flyctl ssh console -a img-fwd"
echo "  SSH to nginx:     flyctl ssh console -a img-fwd-demo-static"
