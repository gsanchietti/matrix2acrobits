# Matrix to Acrobits Proxy

This service acts as a proxy between an Acrobits softphone client and a Matrix homeserver, allowing users to send and receive Matrix messages through an SMS-like interface.

## Technical Overview

The proxy is written in Go and uses the following key technologies:
- **Web Framework**: `github.com/labstack/echo/v4`
- **Matrix Client Library**: `maunium.net/go/mautrix`

The service authenticates to the Matrix homeserver as an **Application Service**, which grants it the ability to perform actions (like sending messages) on behalf of other Matrix users (impersonation).

## Synapse Configuration

To function correctly, the proxy must be registered as an Application Service with your Synapse homeserver.

### 1. Registration File

First, create a registration YAML file (e.g., `acrobits-proxy.yaml`) and place it on your homeserver. This file tells Synapse how to communicate with the proxy.

**`acrobits-proxy.yaml`:**
```yaml
# A unique identifier for the application service.
id: acrobits-proxy
# The URL where Synapse can reach your proxy.
# This may not be used for sending messages but is a required field.
url: http://localhost:8080 
# A secure, randomly generated token your proxy will use to authenticate with Synapse.
as_token: "YOUR_SECURE_APPLICATION_SERVICE_TOKEN"
# A secure, randomly generated token Synapse will use to authenticate with your proxy.
hs_token: "YOUR_SECURE_HOMESERVER_TOKEN"
# The localpart of the 'bot' user for this application service.
sender_localpart: _acrobits_proxy
# This section grants the proxy the power to impersonate users.
namespaces:
  users:
    - exclusive: true
      # This regex must match the user IDs the proxy should control.
      # Setting exclusive: true means the AppService can auto-provision users
      # matching this regex if they don't already exist.
      # Replace with your actual homeserver name.
      regex: '@.*:your-homeserver-name.com'
  aliases: []
  rooms: []
```
*You must generate your own secure random strings for `as_token` and `hs_token`.*

### 2. homeserver.yaml

Next, add the path to your registration file to your Synapse `homeserver.yaml`:

```yaml
app_service_config_files:
  - "/data/config/acrobits-proxy.yaml"
```

Finally, **restart your Synapse server** to load the new configuration.

### 3. dex_config.yaml

Add this do the end of the config file:
```yaml
oauth2:
  skipApprovalScreen: true
  passwordConnector: ldap
```

## Proxy Configuration & Running

The proxy is configured via environment variables.

### Environment Variables

- `PROXY_PORT`: The port for the proxy to listen on (default: `8080`).
- `MATRIX_HOMESERVER_URL`: The full URL of your Matrix homeserver (e.g., `https://matrix.your-homeserver-name.com`).
- `SUPER_ADMIN_TOKEN`: The Application Service token (`as_token`) you defined in the registration file.

### Building and Running

1.  **Build the binary:**
    ```shell
    go build -o matrix2acrobits ./cmd/server
    ```
2.  **Run the server:**
    ```shell
    export PROXY_PORT=8080
    export MATRIX_HOMESERVER_URL="https://matrix.your-homeserver-name.com"
    export SUPER_ADMIN_TOKEN="YOUR_SECURE_APPLICATION_SERVICE_TOKEN"
    
    ./matrix2acrobits
    ```

## API Endpoints

### Client API

These endpoints are used by the Acrobits client. The `password` fields in the requests are ignored, as authentication is handled by the Application Service.

- `POST /api/client/send_message`: Sends a message to a Matrix room on behalf of a user. The `from` field in the JSON body specifies the Matrix user to impersonate.
- `POST /api/client/fetch_messages`: Fetches new messages for a user by performing a Matrix sync. The `username` field specifies the Matrix user to impersonate.

### Internal API

These endpoints are for managing the service and are protected. Access requires passing the `SUPER_ADMIN_TOKEN` in the `X-Super-Admin-Token` HTTP header.

- `POST /api/internal/map_sms_to_matrix`: Creates a mapping between a phone number and a Matrix room ID.
- `GET /api/internal/map_sms_to_matrix`: Looks up a mapping.

## Running Integration Tests

Integration tests verify the end-to-end functionality of the proxy by interacting with a live Matrix homeserver.

### Option 1: Local Testing with Podman Compose (Recommended)

This option brings up an entire local Matrix stack (OpenLDAP → Dex → Synapse) that mirrors the NethServer deployment described in the `homeserver.yaml` and `dex_config.yaml` files.

**Prerequisites:**

- Podman and Podman Compose installed (rootless mode works out of the box)
- Go 1.23+ installed
- `curl` command-line tool

**Quick Start:**

1.  **Run the setup script:**
    ```shell
    ./setup-local-testing.sh
    ```
    This will:
    - Generate a Synapse configuration if missing and add the application-service wiring
    - Provision OpenLDAP by running `ghcr.io/nethserver/openldap-server:latest new-domain` with the same environment variables used in `podman-compose.yaml` (the script creates `lenv` and launches the one-shot `podman run -ti --rm --env-file ./lenv --volume=openldap_data:/var/lib/openldap:z ... new-domain` command)
    - Start OpenLDAP (port 10389), Dex (port 20053), and Synapse (ports 8008/20054)
    - Create the `test.env` file with the synthetic user credentials used by the integration tests

2.  **Run the integration tests:**
    ```shell
    ./run-tests.sh
    ```

3.  **View the stack logs (optional):**
    ```shell
    podman-compose -f podman-compose.yaml logs -f synapse
    podman-compose -f podman-compose.yaml logs -f dex
    podman-compose -f podman-compose.yaml logs -f openldap
    ```

4.  **Clean up when done:**
    ```shell
    ./cleanup-local-testing.sh
    ```

**What the Setup Includes:**

- OpenLDAP instance seeded with `dc=ldap1,dc=local`, managed via `ghcr.io/nethserver/openldap-server`
- Dex OIDC provider configured from `dex_config.yaml`, pointing to the embedded LDAP backend
- Synapse homeserver using `homeserver.yaml` to glue Dex and Application Service registration together
- SQLite-backed Synapse data (`synapse_data` volume)
- Rootless Podman support and pre-configured test users (giacomo & mario)

### Option 2: Remote Testing with External Synapse

If you prefer to test against an existing Synapse installation:

**Prerequisites:**

- A running Synapse homeserver configured as described in the "Synapse Configuration" section
- The Application Service correctly loaded and permissions granted
- The `test.env` file (located in the project root) populated with valid Matrix user credentials and homeserver details

**How to Run:**

1.  **Set Environment Variables:** Create a `test.env` file with the following content:
    ```
    MATRIX_HOMESERVER_URL=https://matrix.your-homeserver-name.com
    SUPER_ADMIN_TOKEN=your_application_service_token
    AS_USER_ID=@_acrobits_proxy:your-homeserver-name.com
    USER1=username1
    USER1_PASSWORD=password1
    USER1_NUMBER=+1001001000
    USER2=username2
    USER2_PASSWORD=password2
    USER2_NUMBER=+1002002000
    ```

2.  **Execute Tests:**
    ```shell
    export RUN_INTEGRATION_TESTS=1
    go test -v ./internal/integration
    ```

**Note:** Remote tests are dependent on a live external service and may be flaky if the network is unstable or the server is misconfigured. Due to the external nature of these tests, potential failures might indicate issues with the Synapse server setup rather than the proxy code.

### Running Specific Tests

To run only specific integration tests:

```shell
# Run only the send and fetch messages test
go test -v ./internal/integration -run TestIntegration_SendAndFetchMessages

# Run only the room messaging test
go test -v ./internal/integration -run TestIntegration_RoomMessaging

# Run only the mapping API test
go test -v ./internal/integration -run TestIntegration_MappingAPI
```


## Acrobits documentation

- https://doc.acrobits.net/api/client/fetch_messages_modern.html
- https://doc.acrobits.net/api/client/send_message.html
