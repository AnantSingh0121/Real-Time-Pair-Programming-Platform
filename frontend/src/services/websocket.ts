export interface WSMessage {
    type: string;
    payload: any;
}

export class WebSocketClient {
    private ws: WebSocket | null = null;
    private roomId: string;
    private userId: string;
    private username: string;
    private reconnectAttempts = 0;
    private maxReconnectAttempts = 5;
    private messageHandlers: Map<string, (payload: any) => void> = new Map();

    constructor(roomId: string, userId: string, username: string) {
        this.roomId = roomId;
        this.userId = userId;
        this.username = username;
    }

    connect(): Promise<void> {
        return new Promise((resolve, reject) => {
            const wsUrl = `ws://localhost:8080/ws/${this.roomId}?userId=${this.userId}&username=${encodeURIComponent(this.username)}`;

            this.ws = new WebSocket(wsUrl);

            this.ws.onopen = () => {
                console.log('WebSocket connected');
                this.reconnectAttempts = 0;
                resolve();
            };

            this.ws.onmessage = (event) => {
                try {
                    const message: WSMessage = JSON.parse(event.data);
                    const handler = this.messageHandlers.get(message.type);
                    if (handler) {
                        handler(message.payload);
                    }
                } catch (error) {
                    console.error('Error parsing WebSocket message:', error);
                }
            };

            this.ws.onerror = (error) => {
                console.error('WebSocket error:', error);
                reject(error);
            };

            this.ws.onclose = () => {
                console.log('WebSocket disconnected');
                this.attemptReconnect();
            };
        });
    }

    private attemptReconnect() {
        if (this.reconnectAttempts < this.maxReconnectAttempts) {
            this.reconnectAttempts++;
            console.log(`Attempting to reconnect (${this.reconnectAttempts}/${this.maxReconnectAttempts})...`);
            setTimeout(() => {
                this.connect().catch(console.error);
            }, 2000 * this.reconnectAttempts);
        }
    }

    on(messageType: string, handler: (payload: any) => void) {
        this.messageHandlers.set(messageType, handler);
    }

    send(type: string, payload: any) {
        if (this.ws && this.ws.readyState === WebSocket.OPEN) {
            const message: WSMessage = { type, payload };
            this.ws.send(JSON.stringify(message));
        } else {
            console.error('WebSocket is not connected');
        }
    }

    sendCodeChange(code: string, language: string, changes?: string) {
        this.send('code_change', {
            roomId: this.roomId,
            userId: this.userId,
            username: this.username,
            code,
            language,
            changes,
        });
    }

    sendChatMessage(text: string) {
        this.send('chat', {
            roomId: this.roomId,
            userId: this.userId,
            username: this.username,
            text,
            timestamp: new Date().toISOString(),
        });
    }

    sendCursorPosition(lineNumber: number, column: number) {
        this.send('cursor', {
            roomId: this.roomId,
            userId: this.userId,
            username: this.username,
            position: { lineNumber, column },
        });
    }

    disconnect() {
        if (this.ws) {
            this.ws.close();
            this.ws = null;
        }
    }
}
