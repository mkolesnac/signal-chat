# Signal Chat Server

A secure chat server implementation using WebSockets and Badger DB.

## Features

- Real-time messaging using WebSockets
- Persistent storage using Badger DB
- Conversation management
- Message broadcasting to multiple participants

## Requirements

- Go 1.16 or higher
- Badger DB

## Installation

1. Clone the repository:
   ```
   git clone https://github.com/yourusername/signal-chat.git
   cd signal-chat
   ```

2. Install dependencies:
   ```
   go mod download
   ```

## Running the Server

To run the server with default settings:

```
go run cmd/server/main.go
```

### Command Line Options

- `-port`: Port to listen on (default: 8080)
- `-data-dir`: Directory to store Badger DB data (default: ./data)

Example:
```
go run cmd/server/main.go -port 9000 -data-dir /path/to/data
```

## API Endpoints

### WebSocket Connection

```
GET /ws?userID=<user_id>
```

Establishes a WebSocket connection for a user. The `userID` query parameter is required.

### Create Conversation

```
POST /api/conversations
```

Creates a new conversation with the specified participants.

Request body:
```json
{
  "conversationID": "conv-123",
  "otherParticipants": [
    {
      "id": "user-2",
      "keyDistributionMessage": "..."
    },
    {
      "id": "user-3",
      "keyDistributionMessage": "..."
    }
  ]
}
```

Headers:
- `X-User-ID`: ID of the user creating the conversation

Response:
```json
{
  "conversationID": "conv-123"
}
```

### Send Message

```
POST /api/messages
```

Sends a message to a conversation.

Request body:
```json
{
  "conversationID": "conv-123",
  "encryptedMessage": "..."
}
```

Headers:
- `X-User-ID`: ID of the user sending the message

Response:
```json
{
  "messageID": "msg-1234567890",
  "timestamp": 1234567890
}
```

## WebSocket Message Types

- `MessageTypeSync`: Synchronization message
- `MessageTypeNewMessage`: New message notification
- `MessageTypeNewConversation`: New conversation notification
- `MessageTypeParticipantAdded`: Participant added notification
- `MessageTypeAck`: Acknowledgment message

## License

MIT 