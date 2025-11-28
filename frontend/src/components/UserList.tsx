import React from 'react';
import './UserList.css';

interface User {
    userId: string;
    username: string;
    status: 'online' | 'offline' | 'typing';
}

interface UserListProps {
    users: User[];
    currentUser: any;
}

export const UserList: React.FC<UserListProps> = ({ users, currentUser }) => {
    return (
        <div className="user-list glass-panel">
            <div className="user-list-header">
                <h3>Online Users ({users.length})</h3>
            </div>

            <div className="users-container">
                {users.map((user) => (
                    <div key={user.userId} className="user-item">
                        <div className={`status-indicator ${user.status}`}></div>
                        <span className="username">
                            {user?.username || 'Unknown'} {user?.userId === currentUser?.userId && '(You)'}
                        </span>
                        {user.status === 'typing' && (
                            <span className="typing-indicator">typing...</span>
                        )}
                    </div>
                ))}
            </div>
        </div>
    );
};
