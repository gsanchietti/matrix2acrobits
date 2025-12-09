# Matrix ↔ Acrobits Push Notification Integration


## Overview

This document describes the complete bidirectional push notification integration between Acrobits mobile clients and Matrix/Synapse homeserver, as implemented by the matrix2acrobits proxy. It covers both pusher registration (client → proxy → Synapse) and push gateway handling (Synapse → proxy → Acrobits PNM → device).

---

## Architecture

```
┌───────────────┐     ┌────────────────────┐     ┌───────────────┐
│ Acrobits App  │────>│  matrix2acrobits   │────>│   Synapse     │
│ (Mobile)      │     │      Proxy         │     │ Homeserver    │
└───────────────┘     └────────────────────┘     └───────────────┘
      │                     │      ▲
      │                     │      │
      ▼                     ▼      │
  ┌─────────────┐        ┌─────────────┐
  │ Acrobits    │        │ Mobile      │
  │   PNM       │        │ Device      │
  └─────────────┘        └─────────────┘
```

### Data Flow

1. **Push Token Registration:**
  - Acrobits app sends push token to proxy
  - Proxy saves token and registers a pusher with Synapse
2. **Push Notification Delivery:**
  - Synapse sends push notification to proxy (Push Gateway API)
  - Proxy translates and forwards notification to Acrobits PNM
  - Acrobits PNM delivers to mobile device

---

## How It Works

### 1. Pusher Registration (Client → Proxy → Synapse)

When the Acrobits mobile app starts, it reports its push token to the proxy. The proxy:
1. Stores the token and app info in its local database
2. Resolves the selector to a Matrix user ID
3. Registers a pusher with Synapse using the Matrix Client-Server API (`/_matrix/client/v3/pushers/set`)
  - The pusher tells Synapse to send push notifications for this user to the proxy's push gateway endpoint

### 2. Push Notification Delivery (Synapse → Proxy → Acrobits PNM → Device)

When a message arrives for a user, Synapse:
1. Finds registered pushers for the user
2. Sends a push notification to the proxy (`/_matrix/push/v1/notify`)
3. The proxy:
  - Looks up the pushkey in its database
  - Translates the Matrix notification format to Acrobits PNM format:
    - Maps `event_id` → `Id` (deduplication)
    - Maps `sender`/`sender_display_name` → `UserName`/`UserDisplayName`
    - Maps message `body` → `Message`
    - Maps `unread` count → `Badge`
    - Maps `room_id` → `ThreadId`
    - Extracts `sound` from `tweaks`
  - Forwards the notification to Acrobits PNM (`https://pnm.cloudsoftphone.com/pnm2/send`)
  - Handles response: returns rejected pushkeys to Synapse if tokens are invalid (404 from Acrobits)

---

## 1. Push Token Registration (Client → Proxy → Synapse)

### Flow

1. **Client reports push token:**
   - `POST /api/client/push_token_report`
   - Example:
     ```json
     {
       "selector": "12869E0E6E553673C54F29105A0647204C416A2A:7C3A0D14",
       "token_msgs": "APA91bG9aqWvmnxnYBZWG9hxvtkgzTXSopfiufzmc6tP3Kb...",
       "app_id_msgs": "com.acrobits.softphone",
       "token_calls": "...",
       "app_id_calls": "..."
     }
     ```
2. **Proxy saves token** to local SQLite DB
3. **Proxy resolves selector** to Matrix user ID
4. **Proxy registers pusher** with Synapse:
   - `POST /_matrix/client/v3/pushers/set`
   - Example:
     ```json
     {
       "app_display_name": "com.acrobits.softphone",
       "app_id": "com.acrobits.softphone",
       "append": false,
       "device_display_name": "Acrobits Softphone",
       "kind": "http",
       "lang": "en",
       "pushkey": "APA91bG9aqWvmnxnYBZWG9hxvtkgzTXSopfiufzmc6tP3Kb...",
       "data": {
         "format": "event_id_only",
         "url": "https://matrix-proxy.example.com/_matrix/push/v1/notify"
       }
     }
     ```
   - Tells Synapse to send push notifications to the proxy's push gateway endpoint.

---

## 2. Push Notification Delivery (Synapse → Proxy → Acrobits PNM)

### Flow

1. **Synapse sends notification** to proxy:
   - `POST /_matrix/push/v1/notify`
   - Example:
     ```json
     {
       "notification": {
         "event_id": "$event_id",
         "room_id": "!room:example.com",
         "sender": "@alice:example.com",
         "sender_display_name": "Alice",
         "content": {
           "msgtype": "m.text",
           "body": "Hello!"
         },
         "devices": [
           {
             "app_id": "com.acrobits.softphone",
             "pushkey": "APA91bG9aqWvmnxnYBZWG9hxvtkgzTXSopfiufzmc6tP3Kb...",
             "tweaks": {}
           }
         ]
       }
     }
     ```
2. **Proxy looks up push token** in DB using pushkey
3. **Proxy translates notification** to Acrobits format:
   - Maps `event_id` → `Id`
   - Maps `sender`/`sender_display_name` → `UserName`/`UserDisplayName`
   - Maps message `body` → `Message`
   - Maps `unread` count → `Badge`
   - Maps `room_id` → `ThreadId`
   - Extracts `sound` from `tweaks`
4. **Proxy forwards to Acrobits PNM** at `https://pnm.cloudsoftphone.com/pnm2/send`:
   - Example:
     ```json
     {
       "verb": "NotifyTextMessage",
       "AppId": "com.acrobits.softphone",
       "DeviceToken": "APA91bG9aqWvmnxnYBZWG9hxvtkgzTXSopfiufzmc6tP3Kb...",
       "Badge": 1,
       "Sound": "default",
       "UserName": "@alice:example.com",
       "Message": "Hello!",
       "ContentType": "text/plain",
       "ThreadId": "!room:example.com"
     }
     ```
5. **Proxy handles response:**
   - Returns rejected pushkeys to Synapse if tokens are invalid (404 from Acrobits)
   - Example response to Synapse:
     ```json
     {
       "rejected": ["pushkey1", "pushkey2"]
     }
     ```

---

## Configuration

### Matrix Homeserver Setup
- Set push gateway URL to:
  ```
  https://your-proxy-domain.com/_matrix/push/v1/notify
  ```

### Proxy Environment Variable
- Required:
  ```bash
  export PROXY_URL="https://matrix-proxy.example.com"
  ```
- Must be publicly accessible from Synapse and use HTTPS.
- If unset, proxy will skip pusher registration but still handle push notifications.

### Push Token Registration
- Clients must report tokens via `/api/client/push_token_report`.
- Stores selector, token/app IDs for messages/calls.

---

## Implementation Details

### Error Handling
- **Push token not found:** Pushkey added to `rejected` list
- **Acrobits PNM 404:** Token is invalid, added to `rejected` list
- **Other Acrobits errors:** Logged, not marked as rejected
- **Network errors:** Logged, not rejected (homeserver will retry)
- **Pusher registration errors:** Logged, token still saved

### Deduplication
- Matrix `event_id` passed as Acrobits `Id` for deduplication

### Design Decisions
- **Append=false:** Only one active pusher per app/device
- **Format=event_id_only:** Minimal data sent, privacy preserved
- **User resolution:** Selector resolved to Matrix user ID

---

## References
- [Matrix Push Gateway API Spec](https://spec.matrix.org/v1.16/push-gateway-api/)
- [Matrix Client-Server API - Pushers](https://spec.matrix.org/v1.16/client-server-api/#post_matrixclientv3pushersset)
- [Acrobits Push Notifications API](https://doc.acrobits.net/api/server/http_push.html)
- [Acrobits Push Notification Manager](https://doc.acrobits.net/api/server/http_push.html)
