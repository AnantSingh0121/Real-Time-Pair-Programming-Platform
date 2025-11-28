# Real-Time Pair Programming Platform
A collaborative coding environment that allows multiple developers to edit code together, run code, chat and debug in real-time.

## Features

- **Real-time Collaborative Editing** - Multiple users can edit code simultaneously with cursor synchronization
- **Integrated Chat** - Built-in messaging system for team communication
- **Code Execution** - Run code directly in the browser with shared terminal output
- **AI Autocomplete** - Intelligent code suggestions powered by AI
- **Room Management** - Create and join coding sessions with unique room IDs
- **User Presence** - See who's online and actively coding
- **Multi-language Support** - Execute Python, JavaScript and more

## Prerequisites

- Node.js 18+ and npm
- Go 1.21+
- Python 3.10+
- AWS Account with DynamoDB access

## Setup Instructions

### 1. Clone and Configure

```bash
cd "Real-Time Pair Programming Platform"
```

The `.env` file should already be configured with your AWS credentials.

### 2. Setup DynamoDB Tables

```bash
cd scripts
npm install
node setup-dynamodb.js
```

This creates the required DynamoDB tables: Users, Rooms, Messages and CodeSync.

### 3. Start Backend Services

**Terminal 1 - Golang Backend:**
```bash
cd backend-go
go mod download
go run main.go
```

**Terminal 2 - Python Service:**
```bash
cd backend-python
pip install -r requirements.txt
uvicorn main:app --port 8001 --reload
```

### 4. Start Frontend

**Terminal 3 - React Frontend:**
```bash
cd frontend
npm install
npm run dev
```

Access the application at: **http://localhost:5173**

## Usage

1. **Sign Up** - Create a new account
2. **Create Room** - Start a new coding session
3. **Share Room ID** - Invite others to join
4. **Code Together** - Edit, chat and run code in real-time

## Tech Stack

- **Frontend**: React, TypeScript, Monaco Editor, Vite
- **Backend**: Golang (Gorilla WebSocket, Chi Router, AWS SDK)
- **Execution Service**: Python (FastAPI)
- **Database**: AWS DynamoDB
- **Authentication**: JWT

## API Endpoints

### Authentication
- `POST /api/auth/signup` - Register new user
- `POST /api/auth/login` - Login and get JWT token

### Rooms
- `GET /api/rooms` - List all rooms
- `POST /api/rooms` - Create new room
- `POST /api/rooms/:roomId/join` - Join a room

### WebSocket
- `WS /ws/:roomId` - Real-time communication

### Code Execution
- `POST http://localhost:8001/execute` - Run code
- `POST http://localhost:8001/autocomplete` - Get suggestions

## Security Notes

- Change `JWT_SECRET` in `.env` for production
- Never commit `.env` file to version control
- Code execution runs in a sandboxed environment
- Consider Docker containers for production code execution

## License

MIT License - Feel free to use this project for learning and development.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
