import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { roomAPI, Room } from '../services/api';
import { useAuthStore } from '../store/authStore';
import './Dashboard.css';

export const Dashboard: React.FC = () => {
    const [rooms, setRooms] = useState<Room[]>([]);
    const [roomName, setRoomName] = useState('');
    const [joinRoomId, setJoinRoomId] = useState('');
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState('');

    const navigate = useNavigate();
    const { user, logout } = useAuthStore();

    useEffect(() => {
        loadRooms();
    }, []);

    const loadRooms = async () => {
        try {
            const data = await roomAPI.getRooms();
            setRooms(data);
        } catch (err) {
            console.error('Failed to load rooms:', err);
        }
    };

    const handleCreateRoom = async (e: React.FormEvent) => {
        e.preventDefault();
        setError('');
        setLoading(true);

        try {
            const { room } = await roomAPI.createRoom(roomName || 'Untitled Room');
            navigate(`/room/${room.roomId}`);
        } catch (err: any) {
            setError(err.response?.data?.error || 'Failed to create room');
        } finally {
            setLoading(false);
        }
    };

    const handleJoinRoom = async (e: React.FormEvent) => {
        e.preventDefault();
        if (!joinRoomId.trim()) {
            setError('Please enter a Room ID');
            return;
        }

        setError('');
        setLoading(true);

        try {
            console.log('Attempting to join room:', joinRoomId);
            const response = await roomAPI.joinRoom(joinRoomId.trim());
            console.log('Join room response:', response);
            navigate(`/room/${joinRoomId.trim()}`);
        } catch (err: any) {
            console.error('Join room error:', err);
            console.error('Error response:', err.response);
            const errorMsg = err.response?.data?.error || err.message || 'Failed to join room. Please check the ID.';
            setError(errorMsg);
        } finally {
            setLoading(false);
        }
    };

    return (
        <div className="dashboard-container">
            <nav className="dashboard-nav glass-panel">
                <div className="nav-content">
                    <h2 className="gradient-text">Pair Programming</h2>
                    <div className="nav-actions">
                        <span className="user-info">{user?.username}</span>
                        <button onClick={logout} className="btn btn-secondary">
                            Logout
                        </button>
                    </div>
                </div>
            </nav>

            <div className="dashboard-content">
                <div className="dashboard-header fade-in">
                    <h1>Welcome, {user?.username} ! </h1>
                    <p>Create a new room or join an existing one to start collaborating</p>
                </div>

                {error && (
                    <div className="error-message">
                        {error}
                    </div>
                )}

                <div className="dashboard-grid">
                    {/* Create Room Card */}
                    <div className="card fade-in">
                        <h3>Create New Room</h3>
                        <p className="card-description">
                            Start a new collaborative coding session
                        </p>
                        <form onSubmit={handleCreateRoom} className="room-form">
                            <input
                                type="text"
                                className="input"
                                placeholder="Room name (optional)"
                                value={roomName}
                                onChange={(e) => setRoomName(e.target.value)}
                            />
                            <button type="submit" className="btn btn-primary w-full" disabled={loading}>
                                Create Room
                            </button>
                        </form>
                    </div>

                    {/* Join Room Card */}
                    <div className="card fade-in">
                        <h3>Join Existing Room</h3>
                        <p className="card-description">
                            Enter a room ID to join a session
                        </p>
                        <form onSubmit={handleJoinRoom} className="room-form">
                            <input
                                type="text"
                                className="input"
                                placeholder="Room ID"
                                value={joinRoomId}
                                onChange={(e) => setJoinRoomId(e.target.value)}
                            />
                            <button type="submit" className="btn btn-primary w-full" disabled={loading}>
                                Join Room
                            </button>
                        </form>
                    </div>
                </div>

                {/* Recent Rooms */}
                <div className="recent-rooms fade-in">
                    <h2>Recent Rooms</h2>
                    {rooms.length === 0 ? (
                        <div className="empty-state">
                            <p>No rooms yet. Create one to get started!</p>
                        </div>
                    ) : (
                        <div className="rooms-grid">
                            {rooms.map((room) => (
                                <div
                                    key={room.roomId}
                                    className="room-card glass-panel"
                                    onClick={() => navigate(`/room/${room.roomId}`)}
                                >
                                    <div className="room-header">
                                        <h4>{room.name}</h4>
                                        <span className="badge badge-info">
                                            {room.users.length} {room.users.length === 1 ? 'user' : 'users'}
                                        </span>
                                    </div>
                                    <p className="room-id">ID: {room.roomId}</p>
                                    <p className="room-date">
                                        Created {new Date(room.createdAt).toLocaleDateString()}
                                    </p>
                                </div>
                            ))}
                        </div>
                    )}
                </div>
            </div>
        </div>
    );
};
