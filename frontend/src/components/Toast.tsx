import React, { useEffect } from 'react';
import './Toast.css';

export interface ToastMessage {
    id: string;
    message: string;
    type: 'info' | 'success' | 'error';
}

interface ToastProps {
    toasts: ToastMessage[];
    onRemove: (id: string) => void;
}

export const Toast: React.FC<ToastProps> = ({ toasts, onRemove }) => {
    console.log('Toast component rendered with:', toasts);
    useEffect(() => {
        const timers: ReturnType<typeof setTimeout>[] = [];

        toasts.forEach((toast) => {
            const timer = setTimeout(() => {
                onRemove(toast.id);
            }, 4000);
            timers.push(timer);
        });

        return () => {
            timers.forEach(timer => clearTimeout(timer));
        };
    }, [toasts, onRemove]);

    return (
        <div className="toast-container">
            {toasts.map((toast) => (
                <div key={toast.id} className={`toast toast-${toast.type}`}>
                    <span className="toast-message">{toast.message}</span>
                    <button
                        className="toast-close"
                        onClick={() => onRemove(toast.id)}
                        aria-label="Close"
                    >
                        ×
                    </button>
                </div>
            ))}
        </div>
    );
};
