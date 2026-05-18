# [Toolforge Deployment](https://toolsadmin.wikimedia.org/)

[Toolforge](https://wikitech.wikimedia.org/wiki/Help:Toolforge) is Wikimedia's hosting platform for community tools. This guide walks you through deploying Scribe-Server on Toolforge from scratch — no prior Toolforge experience needed. To deploy Scribe Server here, you first apply at [toolsadmin.wikimedia.org](https://toolsadmin.wikimedia.org/) with the relevant project details. After your application is approved, you gain SSH access and can set up the environment, database, and web service described below.

## Contents

- [Practical Workflow](#first-steps-as-a-contributor)
  - [SSH into Toolforge](#ssh-into-toolforge)
  - [Set Up Go](#set-up-go)
  - [Configure the Database](#configure-the-database)
  - [Build and Run the Server](#build-and-run-the-server)
  - [Install pip](#install-pip)
  - [Install PyICU](#install-pyicu)

## Practical Workflow

### SSH into Toolforge

Connect to the login node using your Wikimedia developer account:

```bash
ssh {user_id}@login.toolforge.org
```

Once inside, switch to the Scribe tool account so all subsequent commands run in the correct project context:

```bash
become testserver-scribe
```

Then clone the repository into your project directory:

```bash
git clone https://github.com/scribe-org/Scribe-Server.git
cd Scribe-Server
```

### Set Up Go

Toolforge's [pre-built web images](https://wikitech.wikimedia.org/wiki/Help:Toolforge/Web#Other_/_generic_web_servers) do not include Go, so you install it manually into your home directory.

> **Note:** `go-sqlite3` requires CGo and `make`, neither of which is available on Toolforge. Use a pure-Go SQLite driver instead.

```bash
# Download Go
wget https://go.dev/dl/go1.23.6.linux-amd64.tar.gz

# Extract the tarball
tar -xzf go1.23.6.linux-amd64.tar.gz -C ~/

# Rename the directory for organization
mv ~/go ~/go1.23

# Persist environment variables across sessions
echo 'export GOROOT=$HOME/go1.23' >> ~/.bashrc
echo 'export PATH=$PATH:$GOROOT/bin' >> ~/.bashrc

# Apply changes to the current shell
source ~/.bashrc

# Verify the installation
go version

# Clean up the downloaded archive
rm go1.23.6.linux-amd64.tar.gz
```

Why this layout:

- Go is extracted to `~/go1.23` rather than `~/go` to make the version explicit and avoid conflicts if you later upgrade.
- `GOROOT` must point at your custom location because the system has no Go in `PATH`.
- Running `go run .` inside Toolforge binds to `0.0.0.0:8000` — this is expected behavior within the Toolforge network.

### Configure the Database

Toolforge provides a shared MariaDB cluster. Your credentials are pre-written to `~/replica.my.cnf` during tool creation.

First, read your credentials:

```bash
cat ~/replica.my.cnf
```

Then copy the example config and fill in the values:

```bash
mv config-example.yaml config.yaml
nano config.yaml
```

```yaml
# Server configuration
hostPort: 8000
fileSystem: "./packs"

# Database configuration
database:
  user: {user}
  password: {password}
  host: tools-db.tools.eqiad1.wikimedia.cloud
  port: "3306"
  name: {user}__scribe_server_p
```

Replace `{user}` and `{password}` with the values from `replica.my.cnf`. The database name follows the Toolforge convention: `{user}__<db_name>`.

To inspect the database directly at any time:

```bash
mysql --defaults-file=~/replica.my.cnf \
  -h tools-db.tools.eqiad1.wikimedia.cloud \
  {user}__scribe_server_p
```

Example:

```bash
mysql --defaults-file=~/replica.my.cnf \
  -h tools-db.tools.eqiad1.wikimedia.cloud \
  s123456__scribe_server_p
```

### Build and Run the Server

Each time you deploy an update, stop the running service, pull the latest code, rebuild the binary, and restart:

```bash
chmod +x update_data.sh
toolforge webservice stop
git pull origin main
go build -o Scribe-Server .
toolforge webservice --mem 4Gi --cpu 2 jdk17 start \
  /data/project/scribe-server/Scribe-Server/update_data.sh
```

Why this sequence:

- `toolforge webservice stop` ensures the old binary is not locked when you overwrite it.
- `go build -o Scribe-Server .` produces a statically-linked binary that Toolforge can execute directly.
- The `--mem 4Gi --cpu 2` flags allocate enough headroom for data loading on startup.

### Install pip

If you need Python-based tooling in the project, open a Python 3.13 shell and bootstrap pip:

```bash
toolforge webservice python3.13 shell

source venv/bin/activate

curl -sS https://bootstrap.pypa.io/get-pip.py | python
```

### Install PyICU

> [!NOTE]
> The following should be done if ICU Detection Fails.

The standard PyICU build uses `pkg-config` or `icu-config` to locate ICU headers and libraries. Neither tool is installed on Toolforge, so you must set the paths manually before running `pip install`:

```bash
export ICU_VERSION=76
export PYICU_LFLAGS="-L/usr/lib/x86_64-linux-gnu -licui18n -licuuc -licudata"
export PYICU_CFLAGS="-I/usr/include"
pip install PyICU
```

Why these variables:

- `ICU_VERSION` tells the build script which header subdirectory to target.
- `PYICU_LFLAGS` points the linker at the system ICU shared objects already present on Toolforge nodes.
- `PYICU_CFLAGS` points the compiler at the system ICU headers.
