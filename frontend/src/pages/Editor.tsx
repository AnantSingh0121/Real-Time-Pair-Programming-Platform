import React, { useState, useEffect, useRef } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import Editor, { OnMount } from '@monaco-editor/react';
import { WebSocketClient } from '../services/websocket';
import { roomAPI, executionAPI } from '../services/api';
import { useAuthStore } from '../store/authStore';
import { useToastStore } from '../store/toastStore';
import * as monaco from "monaco-editor";
import { ChatPanel } from '../components/ChatPanel';
import { UserList } from '../components/UserList';
import { Terminal } from '../components/Terminal';
import { Toast } from '../components/Toast';
import './Editor.css';

interface RoomData {
    room: {
        roomId: string;
        name: string;
        users: string[];
    };
    codeSync: {
        code: string;
        language: string;
    };
}

export const EditorPage: React.FC = () => {
    const { roomId } = useParams<{ roomId: string }>();
    const navigate = useNavigate();
    const { user } = useAuthStore();
    const { toasts, addToast, removeToast } = useToastStore();

    const [roomData, setRoomData] = useState<RoomData | null>(null);
    const [code, setCode] = useState('// Loading...');
    const [language, setLanguage] = useState('python');
    const [stdin, setStdin] = useState('');
    const [output, setOutput] = useState<string[]>([]);
    const [isExecuting, setIsExecuting] = useState(false);
    const [onlineUsers, setOnlineUsers] = useState<any[]>([]);
    const [chatMessages, setChatMessages] = useState<any[]>([]);
    const wsClientRef = useRef<WebSocketClient | null>(null);

    useEffect(() => {
        if (!roomId || !user) return;

        const initRoom = async () => {
            try {
                const data = await roomAPI.getRoom(roomId);
                setRoomData(data);
                if (data.codeSync) {
                    setCode(data.codeSync.code);
                    setLanguage(data.codeSync.language || 'python');
                }

                if (wsClientRef.current) {
                    wsClientRef.current.disconnect();
                }

                const client = new WebSocketClient(roomId, user.userId, user.username);
                await client.connect();
                wsClientRef.current = client;

                client.on('code_change', (payload: any) => {
                    console.log('Code change received:', payload);
                    if (payload.userId !== user.userId) {
                        setCode(payload.code);
                    }
                });

                client.on('user_list', (payload: any) => {
                    console.log('User list received:', payload);
                    setOnlineUsers(payload);
                });

                client.on('user_joined', (payload: any) => {
                    console.log('User joined event received:', payload);
                    if (payload.userId !== user.userId) {
                        console.log('Adding toast for user joined:', payload.username);
                        addToast(`${payload.username} joined the room`, 'success');
                    }
                });

                client.on('user_left', (payload: any) => {
                    console.log('User left:', payload);
                    addToast(`${payload.username} left the room`, 'info');
                });

                client.on('chat', (payload: any) => {
                    console.log('Chat message received:', payload);
                    if (payload.text && (
                        payload.text.includes('joined the room') ||
                        payload.text.includes('left the room')
                    )) {
                        return;
                    }

                    setChatMessages((prev) => {
                        if (prev.some(msg => msg.messageId === payload.messageId)) {
                            return prev;
                        }
                        return [...prev, payload];
                    });
                });

                setOnlineUsers([{ userId: user.userId, username: user.username, status: 'online' }]);

            } catch (err) {
                console.error('Failed to join room:', err);
                navigate('/dashboard');
            }
        };

        initRoom();

        return () => {
            if (wsClientRef.current) {
                wsClientRef.current.disconnect();
                wsClientRef.current = null;
            }
        };
    }, [roomId, user]);

    const handleEditorDidMount: OnMount = () => {
        monaco.languages.registerCompletionItemProvider(language, {
            provideCompletionItems: async (
                model: monaco.editor.ITextModel,
                position: monaco.Position
            ) => {
                const textUntilPosition = model.getValueInRange({
                    startLineNumber: 1,
                    startColumn: 1,
                    endLineNumber: position.lineNumber,
                    endColumn: position.column,
                });

                try {
                    const response = await executionAPI.getAutocomplete(
                        model.getValue(),
                        language,
                        textUntilPosition.length
                    );

                    return {
                        suggestions: response.suggestions || []
                    };
                } catch {
                    return { suggestions: [] };
                }
            }
        });
    };

    const handleCodeChange = (value: string | undefined) => {
        if (value !== undefined) {
            setCode(value);
            wsClientRef.current?.sendCodeChange(value, language);
        }
    };

    const handleRunCode = async () => {
        setIsExecuting(true);
        setOutput([]);

        try {
            const result = await executionAPI.executeCode(code, language, stdin);

            if (result.error) {
                setOutput([result.error]);
            } else {
                setOutput(result.output.split('\n'));
            }
        } catch (err: any) {
            setOutput(['Error executing code: ' + err.message]);
        } finally {
            setIsExecuting(false);
        }
    };

    const handleLanguageChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
        const newLang = e.target.value;
        setLanguage(newLang);
    };

    const copyRoomId = () => {
        if (roomId) {
            navigator.clipboard.writeText(roomId);
            addToast('Room ID copied to clipboard!', 'success');
        }
    };

    return (
        <div className="editor-page">
            <header className="editor-header glass-panel">
                <div className="header-left">
                    <button onClick={() => navigate('/dashboard')} className="back-btn">
                        ←
                    </button>
                    <div className="room-info">
                        <h2>{roomData?.room.name || 'Loading...'}</h2>
                        <div className="room-id-badge" onClick={copyRoomId} title="Click to copy">
                            ID: {roomId}
                        </div>
                    </div>
                </div>

                <div className="header-controls">
                    <select
                        value={language}
                        onChange={handleLanguageChange}
                        className="language-select"
                    >
                        <option value="python">Python</option>
                        <option value="javascript">JavaScript</option>
                        <option value="go">Go</option>
                        <option value="cpp">C++</option>
                    </select>

                    <button
                        className="btn btn-primary run-btn"
                        onClick={handleRunCode}
                        disabled={isExecuting}
                    >
                        {isExecuting ? 'Running...' : '▶ Run Code'}
                    </button>
                </div>

                <div className="header-right">
                    <div className="user-avatar">
                        {user?.username[0].toUpperCase()}
                    </div>
                </div>
            </header>

            <div className="editor-layout">
                <div className="sidebar-left">
                    <UserList users={onlineUsers} currentUser={user} />
                </div>

                <div className="main-content">
                    <div className="editor-container glass-panel">
                        <Editor
                            height="100%"
                            language={language}
                            value={code}
                            theme="vs-dark"
                            onChange={handleCodeChange}
                            onMount={handleEditorDidMount}
                            options={{
                                minimap: { enabled: false },
                                fontSize: 14,
                                fontFamily: "'Fira Code', monospace",
                                fontLigatures: true,
                                automaticLayout: true,
                                scrollBeyondLastLine: false,
                                padding: { top: 16, bottom: 16 },
                            }}
                        />
                    </div>

                    <div className="terminal-container">
                        <Terminal
                            output={output}
                            isExecuting={isExecuting}
                            stdin={stdin}
                            onStdinChange={setStdin}
                            onClear={() => setOutput([])}
                        />
                    </div>
                </div>

                <div className="sidebar-right">
                    <ChatPanel
                        roomId={roomId || ''}
                        wsClient={wsClientRef.current}
                        messages={chatMessages}
                    />
                </div>
            </div>

            <Toast toasts={toasts} onRemove={removeToast} />
        </div>
    );
};
