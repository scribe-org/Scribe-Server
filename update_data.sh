#!/bin/bash
# update_data.sh - Automated Scribe-Data update script for Scribe-Server
# Pipeline: clone our repo -> create venv -> install dependencies -> generate json file -> convert to sqlite -> copy to ./packs/sqlite -> run make migrate

set -e  # Exit immediately on error

# Configuration
SCRIBE_DATA_DIR="Scribe-Data"
TEMP_DIR="/tmp/scribe-data-update"
PACKS_DIR="./packs/sqlite"
VENV_DIR="./.venv"
LOG_FILE="/tmp/scribe-data-update.log"

# Save project root
PROJECT_ROOT=$(pwd)

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
log() {
    echo -e "${BLUE}[$(date '+%Y-%m-%d %H:%M:%S')]${NC} $1" | tee -a "$LOG_FILE"
}
error() {
    echo -e "${RED}[ERROR]${NC} $1" | tee -a "$LOG_FILE"
}
success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1" | tee -a "$LOG_FILE"
}
warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1" | tee -a "$LOG_FILE"
}

# Cleanup for temp directory
cleanup() {
    if [ -d "$TEMP_DIR" ]; then
        log "Cleaning up temporary directory: $TEMP_DIR"
        rm -rf "$TEMP_DIR"
    fi
}
trap cleanup EXIT

# Step 0: Create and enter temp working dir
mkdir -p "$TEMP_DIR"
cd "$TEMP_DIR"

log "🚀 Starting Scribe-Data update process..."
log "Working directory: $(pwd)"
log "Log file: $LOG_FILE"

# Step 1: Clone or update Scribe-Data
log "📦 Setting up Scribe-Data repository..."
if [ ! -d "$SCRIBE_DATA_DIR" ]; then
    log "Cloning Scribe-Data repository..."
    git clone --depth=1 https://github.com/scribe-org/Scribe-Data.git "$SCRIBE_DATA_DIR" || {
        error "Failed to clone Scribe-Data repo"
        exit 1
    }
    success "Repository cloned successfully"
else
    log "Repository exists, updating..."
    cd "$SCRIBE_DATA_DIR"
    git pull origin main || warning "Failed to update repository, continuing with existing version"
    cd ..
fi

cd "$SCRIBE_DATA_DIR"

# Step 2: Ensure Python and pip are available
log "🐍 Checking Python environment..."
if ! command -v python3 &> /dev/null; then
    error "Python3 is not installed. Please install Python3 first."
    exit 1
fi

if ! command -v pip &> /dev/null && ! command -v pip3 &> /dev/null; then
    warning "pip not found. Attempting to download and install pip..."
    if [ ! -f "get-pip.py" ]; then
        log "Downloading get-pip.py from PyPA..."
        curl -sS https://bootstrap.pypa.io/get-pip.py -o get-pip.py || {
            error "Failed to download get-pip.py"
            exit 1
        }
    fi
    python3 get-pip.py || {
        error "Failed to install pip"
        exit 1
    }
    success "pip installed successfully"
fi

# Step 3: Create or reuse virtual environment
log "🧪 Setting up virtual environment..."
if [ ! -d "$VENV_DIR" ]; then
    python3 -m venv "$VENV_DIR" || {
        error "Failed to create virtual environment"
        exit 1
    }
    success "Virtual environment created at $VENV_DIR"
else
    log "Using existing virtual environment at $VENV_DIR"
fi

log "🔬 Activating virtual environment..."
source "$VENV_DIR/bin/activate" || {
    error "Failed to activate virtual environment"
    exit 1
}
success "Virtual environment activated"

# Step 4: Install dependencies
log "📚 Installing Scribe-Data dependencies..."
pip install --upgrade pip
pip install -e . || {
    error "Failed to install Scribe-Data dependencies"
    exit 1
}
success "Dependencies installed successfully"

# Step 5: Generate language data (auto-confirming prompt)
log "⚡ Generating language data (auto-confirm)..."
yes y | scribe-data get -a -wdp || {
    error "Failed to generate language data"
    exit 1
}
success "Language data generated successfully"

# Step 6: Convert to SQLite
log "🗄️  Converting to SQLite format..."
scribe-data convert -a -ot sqlite || {
    error "Failed to convert to SQLite format"
    exit 1
}
success "Data converted to SQLite successfully"

# Step 7: Check SQLite output
SQLITE_EXPORT_DIR="./scribe_data_sqlite_export"
if [ ! -d "$SQLITE_EXPORT_DIR" ]; then
    error "SQLite export directory not found: $SQLITE_EXPORT_DIR"
    exit 1
fi

SQLITE_FILES=$(find "$SQLITE_EXPORT_DIR" -name "*.sqlite" | wc -l)
if [ "$SQLITE_FILES" -eq 0 ]; then
    error "No SQLite files found in $SQLITE_EXPORT_DIR"
    exit 1
fi
log "Found $SQLITE_FILES SQLite files to copy"

# Step 8: Copy SQLite files to packs
cd "$PROJECT_ROOT"
mkdir -p "$PACKS_DIR"
log "📁 Copying SQLite files to server..."
cp -f "$TEMP_DIR/$SCRIBE_DATA_DIR/scribe_data_sqlite_export"/*.sqlite "$PACKS_DIR/" || {
    error "Failed to copy SQLite files to $PACKS_DIR"
    exit 1
}

log "Copied files:"
ls -la "$PACKS_DIR"/*.sqlite | while read -r line; do
    log "  ✅ $line"
done
success "SQLite files copied successfully"

# Step 9: Run migration from server root
log "🔄 Running database migration..."
make migrate || {
    error "Migration failed"
    exit 1
}
success "Database migration completed successfully"

# Step 10: Deactivate virtual environment
log "🧹 Deactivating virtual environment..."
deactivate
success "Virtual environment deactivated"

# Step 11: Done
END_TIME=$(date '+%Y-%m-%d %H:%M:%S')
success "✨ Scribe-Data update process completed successfully at $END_TIME"

log "📊 Update Summary:"
log "  • Repository: Updated/Cloned"
log "  • Virtual Environment: Reused or created at $VENV_DIR"
log "  • Dependencies: Installed"
log "  • Data Generation: Completed"
log "  • SQLite Conversion: Completed"
log "  • Files Copied: $SQLITE_FILES files"
log "  • Migration: Completed"
log "  • Log file: $LOG_FILE"

echo
success "🎉 Scribe-Data has been updated and migrated to MariaDB!"
echo
log "Next steps:"
log "  • Restart your server if needed"
log "  • Test the /data-version/:language_iso endpoints"
log "  • Check the log file for detailed information: $LOG_FILE"
