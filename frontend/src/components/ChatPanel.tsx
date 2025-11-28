import React, { useState, useEffect, useRef } from 'react';
import { WebSocketClient } from '../services/websocket';
import { useAuthStore } from '../store/authStore';
import './ChatPanel.css';

interface Message {
    messageId: string;
    userId: string;
    username: string;
    text: string;
    timestamp: string;
    isSystem?: boolean;
}

interface ChatPanelProps {
    roomId: string;
    wsClient: WebSocketClient | null;
    messages: Message[];
}

export const ChatPanel: React.FC<ChatPanelProps> = ({ wsClient, messages }) => {
    const [newMessage, setNewMessage] = useState('');
    const messagesEndRef = useRef<HTMLDivElement>(null);
    const { user } = useAuthStore();

    useEffect(() => {
        scrollToBottom();
    }, [messages]);

    const scrollToBottom = () => {
        messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
    };

    const handleSendMessage = (e: React.FormEvent) => {
        e.preventDefault();
        if (!newMessage.trim() || !wsClient) return;

        wsClient.sendChatMessage(newMessage);
        setNewMessage('');
    };

    return (
        <div className="chat-panel glass-panel">
            <div className="chat-header">
                <h3>Chat</h3>
            </div>

            <div className="chat-messages">
                {messages.length === 0 ? (
                    <div className="empty-chat">
                        <p>No messages yet. Say hello! 👋</p>
                    </div>
                ) : (
                    messages.map((msg, index) => {
                        const isMe = msg.userId === user?.userId;
                        const isSystem = msg.isSystem || msg.userId === 'system';

                        if (isSystem) {
                            return (
                                <div key={index} className="message message-system">
                                    <div className="message-content system-message">
                                        {msg.text}
                                    </div>
                                </div>
                            );
                        }

                        return (
                            <div key={index} className={`message ${isMe ? 'message-me' : 'message-other'}`}>
                                <div className="message-header">
                                    <span className="message-user">{isMe ? 'You' : msg.username}</span>
                                    <span className="message-time">
                                        {new Date(msg.timestamp).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
                                    </span>
                                </div>
                                <div className="message-content">
                                    {msg.text}
                                </div>
                            </div>
                        );
                    })
                )}
                <div ref={messagesEndRef} />
            </div>

            <form onSubmit={handleSendMessage} className="chat-input-form">
                <input
                    type="text"
                    className="chat-input"
                    placeholder="Type a message..."
                    value={newMessage}
                    onChange={(e) => setNewMessage(e.target.value)}
                />
                <button type="submit" className="btn btn-primary chat-send-btn">
                    Send
                </button>
            </form>
        </div>
    );
};
