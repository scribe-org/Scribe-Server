#!/bin/bash
# update_data.sh - Automated Scribe-Data update script for Scribe-Server

set -e  # exit immediately on error

# MARK: Config

SCRIBE_DATA_DIR="Scribe-Data"
TEMP_DIR="/tmp/scribe-data-update"
PACKS_DIR="./packs/sqlite"
VENV_DIR="./.venv"
LOG_FILE="/tmp/scribe-data-update.log"
SKIP_MIGRATION=${1:-false}

# Save project root.
PROJECT_ROOT=$(pwd)

# Define target languages and data types
TARGET_LANGUAGES=("english" "french" "german" "italian" "spanish" "portuguese" "russian" "swedish")

# FOR TESTING PURPOSES
# TARGET_LANGUAGES=("english" "french")


DATA_TYPES=("adjectives" "adverbs" "conjunctions" "emoji-keywords" "nouns" "personal-pronouns" "postpositions" "prepositions" "pronouns" "proper-nouns" "verbs")

# FOR TESTING PURPOSES
# DATA_TYPES=("nouns" "verbs")

# Colors for output.
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # no color

# Logging functions.
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

# Cleanup for temp directory.
cleanup() {
    if [ -d "$TEMP_DIR" ]; then
        log "Cleaning up temporary directory: $TEMP_DIR"
        rm -rf "$TEMP_DIR"
    fi
}
trap cleanup EXIT

# MARK: Enter TMP Dir

mkdir -p "$TEMP_DIR"
cd "$TEMP_DIR"

log "🚀 Starting Scribe-Data update process..."
log "Working directory: $(pwd)"
log "Log file: $LOG_FILE"

# MARK: Get Scribe-Data

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

# MARK: Python / Pip

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

# MARK: Make Venv

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

# MARK: Dependencies

log "📚 Installing Scribe-Data dependencies..."
pip install --upgrade pip
pip install -e . || {
    error "Failed to install Scribe-Data dependencies"
    exit 1
}
success "Dependencies installed successfully"

# MARK: Download Wikidata Dump First

DUMP_DIR="./scribe_data_wikidata_dumps_export"
DUMP_FILE="$DUMP_DIR/latest-lexemes.json.bz2"

if [ ! -f "$DUMP_FILE" ]; then
    log "📥 Downloading Wikidata lexeme dump..."
    # Auto-confirm the download prompt with "y" for the initial confirmation
    echo "y" | scribe-data download -wdv 20250730 || {
        error "Failed to download Wikidata dump"
        exit 1
    }
    success "Wikidata dump downloaded successfully"
else
    log "✅ Wikidata dump already exists: $DUMP_FILE"
fi

# MARK: Generate Language Data

log "⚡ Generating language data for target languages (auto-confirm)..."

# Convert arrays to space-separated strings (no quotes around the expansion)
LANG_STRING="${TARGET_LANGUAGES[*]}"
DATA_TYPES_STRING="${DATA_TYPES[*]}"

log "Languages: $LANG_STRING"
log "Data types: $DATA_TYPES_STRING"

# Calculate total number of combinations for the responses
NUM_LANGUAGES=${#TARGET_LANGUAGES[@]}
NUM_DATA_TYPES=${#DATA_TYPES[@]}
TOTAL_COMBINATIONS=$((NUM_LANGUAGES * NUM_DATA_TYPES))

log "Total combinations to process: $TOTAL_COMBINATIONS"
log "Each combination will prompt to 'Use existing latest dump'"
log "Running: scribe-data get -l $LANG_STRING -dt $DATA_TYPES_STRING -wdp $DUMP_DIR"

# Send Down Arrow twice + Enter for each combination
# This selects the 3rd option "Use existing latest dump"
{
    for ((i=1; i<=TOTAL_COMBINATIONS; i++)); do
        printf "\033[B\033[B\n"  # Down arrow, Down arrow, Enter
    done
} | scribe-data get -l $LANG_STRING -dt $DATA_TYPES_STRING -wdp "$DUMP_DIR"

success "Language data generated successfully"

# Sanity Check: Verify generated files
log "🔍 Checking generated data in scribe_data_json_export..."
if [ -d "scribe_data_json_export" ]; then
    # List all generated JSON files organized by language
    for lang in "${TARGET_LANGUAGES[@]}"; do
        if [ -d "scribe_data_json_export/$lang" ]; then
            log "📁 $lang:"
            find "scribe_data_json_export/$lang" -name "*.json" | sort | while read -r file; do
                filename=$(basename "$file")
                log "  ✅ $filename"
            done
        else
            log "⚠️  Missing directory: scribe_data_json_export/$lang"
        fi
    done
    
    # Count total files
    total_files=$(find scribe_data_json_export -name "*.json" | wc -l)
    expected_files=$TOTAL_COMBINATIONS
    log "📊 Generated $total_files/$expected_files JSON files"
    
    if [ "$total_files" -lt "$expected_files" ]; then
        error "⚠️  Expected $expected_files files but only found $total_files"
        error "Some data may not have been generated successfully"
    fi
else
    error "scribe_data_json_export directory not found"
    exit 1
fi

# MARK: Filter Data

CONTRACTS_DIR="./scribe_data_contracts"
log "🔍 Filtering JSON data using contracts..."

if [ ! -d "$CONTRACTS_DIR" ]; then
    warning "Contracts directory not found: $CONTRACTS_DIR"
    warning "Skipping filtering step - proceeding with unfiltered data"
    FILTERED_EXPORT_DIR="./scribe_data_json_export"  # Use original data
else
    FILTERED_EXPORT_DIR="./scribe_data_json_filtered"
    mkdir -p "$FILTERED_EXPORT_DIR"

    log "Running: scribe-data fd -cd $CONTRACTS_DIR -id scribe_data_json_export -od $FILTERED_EXPORT_DIR"
    scribe-data fd -cd "$CONTRACTS_DIR" -id scribe_data_json_export -od "$FILTERED_EXPORT_DIR" || {
        error "Failed to filter JSON data"
        exit 1
    }
    success "JSON data filtered successfully"

    # Debug: Check filtered files
    if [ -d "$FILTERED_EXPORT_DIR" ]; then
        filtered_files=$(find "$FILTERED_EXPORT_DIR" -name "*.json" | wc -l)
        log "📊 Generated $filtered_files filtered JSON files"
    fi

    # MARK: Convert Filtered Data to SQLite

    log "🗄️  Converting filtered data to SQLite format..."
    scribe-data convert -if "$FILTERED_EXPORT_DIR" -lang $LANG_STRING -dt $DATA_TYPES_STRING  -ot sqlite || {
        error "Failed to convert filtered data to SQLite format"
        exit 1
    }
    success "Filtered data converted to SQLite successfully"
fi

# MARK: Check SQLite

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

# MARK: To Packs

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

# MARK: Migration

if [ "$SKIP_MIGRATION" != "true" ]; then
    log "🔄 Running database migration..."
    make migrate || {
        error "Migration failed"
        exit 1
    }
    success "Database migration completed successfully"
else
    log "⏭️ Skipping migration (running in CI/CD)"
fi

# MARK: Finish

log "🧹 Deactivating virtual environment..."
deactivate
success "Virtual environment deactivated"

END_TIME=$(date '+%Y-%m-%d %H:%M:%S')
success "✨ Scribe-Data update process completed successfully at $END_TIME"

log "📊 Update Summary:"
log "  • Repository: Updated/Cloned"
log "  • Virtual Environment: Reused or created at $VENV_DIR"
log "  • Dependencies: Installed"
log "  • Languages processed: ${#TARGET_LANGUAGES[@]} (${TARGET_LANGUAGES[*]})"
log "  • Data types processed: ${#DATA_TYPES[@]}"
log "  • Total combinations: $TOTAL_COMBINATIONS"
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