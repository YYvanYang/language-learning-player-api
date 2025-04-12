**Deployment Guide: Go API + PostgreSQL + MinIO on Raspberry Pi 5**

This guide outlines the steps to deploy the language learning player backend API, its database (PostgreSQL), and object storage (MinIO) as Docker containers onto a Raspberry Pi 5. It assumes you are building the API image on a separate machine (like your Surface - likely x86_64) and deploying to the Raspberry Pi (ARM64).

**I. Prerequisites**

1.  **Development Machine (e.g., Surface):**
    *   Docker Desktop installed (includes `docker buildx`).
    *   Git installed.
    *   Access to your project code (`language-learning-player-api`).
    *   A Docker Hub account (or other container registry).
    *   `make` utility (optional, but helpful for using the Makefile).

2.  **Raspberry Pi 5:**
    *   **Hardware:** Raspberry Pi 5 (8GB RAM model strongly recommended) with an **external SSD connected via USB 3.0** (highly recommended for data storage).
    *   **OS:** Raspberry Pi OS (or another compatible 64-bit Linux distribution).
    *   **Network:** Connected to your network (Ethernet preferred). Note its IP address.
    *   **SSH Access:** Enabled for remote management.
    *   **Docker:** Installed. If not, run:
        ```bash
        curl -fsSL https://get.docker.com -o get-docker.sh
        sudo sh get-docker.sh
        sudo usermod -aG docker $USER
        # Log out and log back in for the group change to take effect!
        ```
    *   **Docker Compose (v2):** Should be included with Docker Desktop installs, or installed as a plugin on Linux. Verify with `docker compose version`. If needed on Pi OS:
        ```bash
        sudo apt-get update
        sudo apt-get install docker-compose-plugin
        ```

**II. Step 1: Build and Push API Docker Image (On Surface)**

This step builds the Go application into a Docker image specifically for the Raspberry Pi's ARM64 architecture and pushes it to Docker Hub.

1.  **Clone Your Code:** If you haven't already, clone your project repository onto your Surface and navigate into the project directory:
    ```bash
    git clone <your-repo-url>
    cd language-learning-player-api
    ```
2.  **Customize Makefile:** Open the `Makefile` and find the `DOCKER_IMAGE_NAME` variable. **Change `your-dockerhub-username` to your actual Docker Hub username.**
    ```makefile
    # Docker image settings (IMPORTANT: Customize for your Docker Hub/Registry)
    DOCKER_IMAGE_NAME ?= your-dockerhub-username/language-player-api # <<< CHANGE THIS
    DOCKER_IMAGE_TAG ?= latest
    ```
3.  **Login to Docker Hub:** Open a terminal on your Surface and log in:
    ```bash
    docker login
    ```
    Enter your Docker Hub username and password when prompted.
4.  **Build and Push:** Run the specific Make target to build the ARM64 image and push it:
    ```bash
    make docker-build-push-arm64
    ```
    *(This uses `docker buildx build --platform linux/arm64 --push ...`)*

    Wait for the build and push process to complete.

**III. Step 2: Prepare Deployment Files (On Raspberry Pi)**

Connect to your Raspberry Pi via SSH and set up the deployment environment.

1.  **Create Deployment Directory:**
    ```bash
    mkdir ~/language-player-deployment
    cd ~/language-player-deployment
    ```
2.  **Create `docker-compose.yml`:** Create the compose file using a text editor (like `nano`):
    ```bash
    nano docker-compose.yml
    ```
    Paste the following content **exactly**, **making sure to replace** `your-dockerhub-username/language-player-api:latest` with the image name you pushed in Step 1. Also, **carefully review the `volumes` section** for `postgres` and `minio` and choose either the SSD bind mount or the named volume option.

    ```yaml
    # docker-compose.yml (Compose V2 / Compose Specification Syntax)

    services:
      # --- API Backend Service ---
      api:
        # !! IMPORTANT: Replace with your actual image on Docker Hub !!
        image: your-dockerhub-username/language-player-api:latest # <<< MAKE SURE THIS MATCHES STEP 1
        container_name: language_player_api
        restart: unless-stopped
        ports:
          - "8080:8080" # Map host port 8080 to container port 8080
        environment:
          # Configuration via environment variables (overrides config files)
          # Sourced from the .env file in the same directory
          APP_ENV: production
          SERVER_PORT: "8080"
          DATABASE_DSN: "postgresql://${POSTGRES_USER}:${POSTGRES_PASSWORD}@postgres:5432/${POSTGRES_DB}?sslmode=disable"
          JWT_SECRETKEY: ${JWT_SECRETKEY}        # MUST be set in .env
          MINIO_ENDPOINT: "minio:9000"          # Service name acts as hostname
          MINIO_ACCESSKEYID: ${MINIO_ACCESS_KEY}
          MINIO_SECRETACCESSKEY: ${MINIO_SECRET_KEY}
          MINIO_BUCKETNAME: ${MINIO_BUCKET_NAME}
          MINIO_USESSL: "false"
          GOOGLE_CLIENTID: ${GOOGLE_CLIENTID}      # Optional: if using Google Login
          GOOGLE_CLIENTSECRET: ${GOOGLE_CLIENTSECRET} # Optional: if using Google Login
          LOG_LEVEL: "info"
          LOG_JSON: "true"
          # CORS_ALLOWEDORIGINS: "http://your-frontend-domain.com" # Adjust as needed
        depends_on:
          postgres:
            condition: service_healthy # Wait for PostgreSQL health check
          minio:
            condition: service_healthy # Wait for Minio health check
        networks:
          - language_player_net # Connect to the custom network

      # --- PostgreSQL Database Service ---
      postgres:
        image: postgres:16-alpine # ARM64 compatible Alpine variant
        container_name: language_player_postgres
        restart: unless-stopped
        environment:
          POSTGRES_USER: ${POSTGRES_USER}         # Sourced from .env
          POSTGRES_PASSWORD: ${POSTGRES_PASSWORD} # Sourced from .env
          POSTGRES_DB: ${POSTGRES_DB}             # Sourced from .env
        volumes:
          # !! RECOMMENDED for Pi: Map to SSD (Choose ONE option) !!

          # Option A: Bind Mount (Map to a directory on your SSD)
          # 1. Ensure the directory exists on the Pi: sudo mkdir -p /mnt/ssd/pgdata && sudo chown $USER:$USER /mnt/ssd/pgdata (Replace path!)
          # 2. Uncomment the following lines and CHANGE the source path:
          # - type: bind
          #   source: /mnt/ssd/pgdata # <--- CHANGE THIS to your actual SSD directory
          #   target: /var/lib/postgresql/data

          # Option B: Named Volume (Easier, stored in /var/lib/docker/volumes on your main drive (SSD))
          # 1. Uncomment the following lines:
          - type: volume
            source: pgdata # Use the named volume defined below
            target: /var/lib/postgresql/data

        # ports: # Optional: Expose port 5432 externally if needed for direct access/migrations
        #  - "5432:5432"
        networks:
          - language_player_net
        healthcheck: # Ensure PostgreSQL is ready before API starts
          test: ["CMD-SHELL", "pg_isready -U ${POSTGRES_USER} -d ${POSTGRES_DB}"]
          interval: 10s
          timeout: 5s
          retries: 5
          start_period: 15s # Allow 15 seconds for PostgreSQL to start up

      # --- MinIO Object Storage Service ---
      minio:
        image: minio/minio:latest # Official image is multi-arch
        container_name: language_player_minio
        restart: unless-stopped
        ports:
          - "9000:9000" # API Port
          - "9001:9001" # Console Port
        environment:
          MINIO_ROOT_USER: ${MINIO_ACCESS_KEY}     # Sourced from .env
          MINIO_ROOT_PASSWORD: ${MINIO_SECRET_KEY} # Sourced from .env
          # MINIO_SERVER_URL / MINIO_BROWSER_REDIRECT_URL might be needed if using a reverse proxy
        volumes:
          # !! RECOMMENDED for Pi: Map to SSD (Choose ONE option) !!

          # Option A: Bind Mount
          # 1. Ensure the directory exists on the Pi: sudo mkdir -p /mnt/ssd/miniodata && sudo chown $USER:$USER /mnt/ssd/miniodata (Replace path!)
          # 2. Uncomment the following lines and CHANGE the source path:
          # - type: bind
          #   source: /mnt/ssd/miniodata # <--- CHANGE THIS to your actual SSD directory
          #   target: /data

          # Option B: Named Volume
          # 1. Uncomment the following lines:
          - type: volume
            source: miniodata # Use the named volume defined below
            target: /data

        command: server /data --console-address ":9001" # Start server and enable console
        networks:
          - language_player_net
        healthcheck: # Ensure MinIO is ready before API starts
          # Use the readiness probe endpoint as per official docs
          test: ["CMD", "curl", "-f", "http://localhost:9000/minio/health/ready"]
          interval: 10s      # Check every 10 seconds
          timeout: 5s       # Wait max 5 seconds for response
          retries: 5        # Retry 5 times before marking as unhealthy
          start_period: 10s # Allow 10 seconds grace period after start

    # --- Define Custom Network ---
    networks:
      language_player_net:
        driver: bridge # Default network driver

    # --- Define Named Volumes (Only needed if using Option B for volumes) ---
    volumes:
      pgdata:
      miniodata:
    ```
    Save and close the file (e.g., `Ctrl+X`, then `Y`, then `Enter` in `nano`).

3.  **Create `.env` File:** Create the environment file:
    ```bash
    nano .env
    ```
    Paste the following content, **replacing all placeholder values** with your own **secure** credentials and settings:
    ```env
    # .env file for docker-compose on Raspberry Pi

    # --- REQUIRED ---

    # PostgreSQL Credentials
    POSTGRES_USER=llp_user        # Example: Choose a username
    POSTGRES_PASSWORD=ReplaceMeWithAVeryStrongDBPassword! # Example: Generate a strong password
    POSTGRES_DB=language_learner_db

    # MinIO Credentials
    MINIO_ACCESS_KEY=llp_minio_key    # Example: Choose an access key
    MINIO_SECRET_KEY=ReplaceMeWithAVeryStrongMinioSecret! # Example: Generate a strong secret
    MINIO_BUCKET_NAME=language-audio

    # API JWT Secret (CRITICAL - make this long and random)
    JWT_SECRETKEY=ReplaceMeWithAReallyLong_Random_UnGuEsSaBlE_JWT_Secret_Key_123$%^

    # --- OPTIONAL (Uncomment and set if using Google Login) ---
    # GOOGLE_CLIENTID=YOUR_GOOGLE_CLIENT_ID_HERE.apps.googleusercontent.com
    # GOOGLE_CLIENTSECRET=YOUR_GOOGLE_CLIENT_SECRET_HERE

    # --- OPTIONAL (Adjust if needed) ---
    # SERVER_PORT=8080
    # LOG_LEVEL=info
    # LOG_JSON=true
    # CORS_ALLOWEDORIGINS="http://your-frontend.com" # Example: Set your frontend URL
    ```
    Save and close the file. **Secure this file:** `chmod 600 .env`.

**IV. Step 3: Run Services (On Raspberry Pi)**

1.  **Start Containers:** In the `~/language-player-deployment` directory, run:
    ```bash
    docker compose up -d
    ```
    Docker Compose will:
    *   Pull the `postgres` and `minio` images (if not already present).
    *   Pull your `api` image from Docker Hub.
    *   Create the network and volumes (if using named volumes).
    *   Start the containers in the correct order based on `depends_on`.
    *   Run services in the background (`-d`).

2.  **Check Initial Status:**
    ```bash
    docker compose ps
    ```
    You should see all three services (`api`, `postgres`, `minio`) listed with a `running` or `healthy` status (it might take a minute for health checks to pass).

**V. Step 4: Post-Deployment Setup (Mandatory)**

These steps need to be done **once** after the first successful startup.

1.  **Run Database Migrations:**
    *   **Option A (Remote from Surface - Recommended):**
        *   Ensure PostgreSQL port 5432 is accessible from your Surface (uncomment `ports` in `postgres` service if needed, check Pi firewall).
        *   On your Surface (in the project directory where the `Makefile` and `migrations` folder are):
            ```bash
            export DATABASE_URL="postgresql://<DB_USER>:<DB_PASSWORD>@<PI_IP_ADDRESS>:5432/<DB_NAME>?sslmode=disable"
            # Replace <...> with values from your Pi's .env file and the Pi's IP Address
            make migrate-up
            ```
    *   **Option B (On Raspberry Pi):**
        *   Install `migrate` CLI on the Pi (refer to `Makefile` `install-migrate` target or official docs).
        *   Copy the `migrations` folder from your project to the Pi (e.g., using `scp`).
        *   Run migrate on the Pi:
            ```bash
            export DATABASE_URL="postgresql://<DB_USER>:<DB_PASSWORD>@localhost:5432/<DB_NAME>?sslmode=disable"
            # Replace <...> with values from your .env file (use localhost here)
            migrate -database "$DATABASE_URL" -path /path/to/copied/migrations up
            ```

2.  **Create MinIO Bucket:**
    *   **Option A (Web Console):** Open `http://<PI_IP_ADDRESS>:9001` in your browser. Log in using `MINIO_ACCESS_KEY` and `MINIO_SECRET_KEY` from your `.env` file. Create a new bucket named exactly as defined in `MINIO_BUCKET_NAME` (e.g., `language-audio`).
    *   **Option B (MinIO Client `mc`):**
        *   Install `mc` on your Surface or Pi.
        *   Configure access to your Pi's MinIO:
            ```bash
            mc alias set pi-minio http://<PI_IP_ADDRESS>:9000 <MINIO_ACCESS_KEY> <MINIO_SECRET_KEY>
            # Replace <...> with values from your .env file and Pi IP
            ```
        *   Create the bucket:
            ```bash
            mc mb pi-minio/${MINIO_BUCKET_NAME}
            # Replace with bucket name from .env
            ```

**VI. Step 5: Verification and Access**

1.  **Check Logs:** Tail the logs to ensure services started correctly and see ongoing activity:
    ```bash
    docker compose logs -f          # Follow logs for all services
    docker compose logs -f api      # Follow logs only for the API service
    docker compose logs -f postgres # Follow logs only for PostgreSQL
    docker compose logs -f minio    # Follow logs only for Minio
    ```
    Press `Ctrl+C` to stop following logs.
2.  **Access API:** Your API should be accessible at `http://<PI_IP_ADDRESS>:8080`. Try accessing a public endpoint (like `/api/v1/audio/tracks` or `/healthz`) in your browser or using `curl`.
3.  **Access MinIO Console:** `http://<PI_IP_ADDRESS>:9001`.

**VII. Step 6: Stopping Services (On Raspberry Pi)**

1.  Navigate to the deployment directory: `cd ~/language-player-deployment`
2.  Stop and remove the containers:
    ```bash
    docker compose down
    ```
    *(This will **not** delete your persistent data stored in volumes or bind mounts.)*
3.  To stop **and delete data volumes** (use with caution!):
    ```bash
    docker compose down -v
    ```

**VIII. Troubleshooting & Tips**

*   **Connection Refused:** If the API can't connect to Postgres or MinIO, double-check:
    *   Service names (`postgres`, `minio`) in `DATABASE_DSN` and `MINIO_ENDPOINT` environment variables.
    *   Network configuration (`language_player_net`) in `docker-compose.yml`.
    *   Credentials match between `.env` file and service configurations.
    *   Postgres/MinIO containers are actually running and healthy (`docker compose ps`).
*   **Permission Denied (Volumes):** If using bind mounts to an SSD, ensure the directories on the Pi (`/mnt/ssd/pgdata`, etc.) exist and have the correct permissions for the user ID running inside the container (often needs wider permissions like `sudo chown -R 999:999 /mnt/ssd/pgdata` for the official Postgres image, or check the specific image documentation). Named volumes handle permissions automatically.
*   **Image Not Found / Wrong Architecture:** If `docker compose up` fails pulling the `api` image, verify:
    *   You pushed the image correctly to Docker Hub with the exact name used in `docker-compose.yml`.
    *   You pushed the `linux/arm64` version (`make docker-build-push-arm64`).
*   **Resource Limits:** If the Pi becomes unresponsive, consider adding resource limits (`deploy.resources.limits`) to services in `docker-compose.yml`, especially for `postgres`.
*   **Updates:** To update the API, run `make docker-build-push-arm64` on your Surface again, then on the Pi run `docker compose pull api && docker compose up -d --force-recreate api`.