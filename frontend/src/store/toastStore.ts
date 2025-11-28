import { create } from 'zustand';

export interface ToastMessage {
    id: string;
    message: string;
    type: 'info' | 'success' | 'error';
}

interface ToastStore {
    toasts: ToastMessage[];
    addToast: (message: string, type: 'info' | 'success' | 'error') => void;
    removeToast: (id: string) => void;
}

export const useToastStore = create<ToastStore>((set) => ({
    toasts: [],
    addToast: (message, type) => {
        const id = `${Date.now()}-${Math.random()}`;
        set((state) => {
            const exists = state.toasts.some(t => t.message === message && t.type === type);
            if (exists) return state;

            return {
                toasts: [...state.toasts, { id, message, type }]
            };
        });
    },
    removeToast: (id) => {
        set((state) => ({
            toasts: state.toasts.filter((toast) => toast.id !== id)
        }));
    },
}));
