import axios from 'axios';

const API_BASE_URL = 'http://localhost:8080/api';

const api = axios.create({
    baseURL: API_BASE_URL,
    headers: {
        'Content-Type': 'application/json',
    },
});

api.interceptors.request.use((config) => {
    const token = localStorage.getItem('token');
    if (token) {
        config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
});

export interface User {
    userId: string;
    username: string;
    email: string;
    createdAt: string;
    lastSeen: string;
}

export interface Room {
    roomId: string;
    name: string;
    createdBy: string;
    users: string[];
    createdAt: string;
}

export interface AuthResponse {
    token: string;
    userId: string;
    user: User;
}

export const authAPI = {
    signup: async (username: string, email: string, password: string): Promise<AuthResponse> => {
        const { data } = await api.post('/auth/signup', { username, email, password });
        return data;
    },

    login: async (email: string, password: string): Promise<AuthResponse> => {
        const { data } = await api.post('/auth/login', { email, password });
        return data;
    },
};

export const roomAPI = {
    getRooms: async (): Promise<Room[]> => {
        const { data } = await api.get('/rooms');
        return data;
    },

    createRoom: async (name: string): Promise<{ room: Room; message: string }> => {
        const { data } = await api.post('/rooms', { name });
        return data;
    },

    getRoom: async (roomId: string): Promise<{ room: Room; codeSync: any }> => {
        const { data } = await api.get(`/rooms/${roomId}`);
        return data;
    },

    joinRoom: async (roomId: string): Promise<{ message: string }> => {
        try {
            const { data } = await api.post(`/rooms/${roomId}/join`);
            return data;
        } catch (error) {
            console.error('API Join Room Error:', error);
            throw error;
        }
    },
};

export const executionAPI = {
    executeCode: async (code: string, language: string, stdin: string = ''): Promise<any> => {
        const { data } = await axios.post('http://localhost:8001/execute', {
            code,
            language,
            stdin,
        });
        return data;
    },

    getAutocomplete: async (code: string, language: string, cursorPosition: number): Promise<any> => {
        const { data } = await axios.post('http://localhost:8001/autocomplete', {
            code,
            language,
            cursorPosition,
        });
        return data;
    },
};

export default api;
