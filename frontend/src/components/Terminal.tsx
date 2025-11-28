import React, { useEffect, useRef } from 'react';
import './Terminal.css';

interface TerminalProps {
    output: string[];
    isExecuting: boolean;
    stdin?: string;
    onStdinChange?: (value: string) => void;
    onClear: () => void;
}

export const Terminal: React.FC<TerminalProps> = ({ output, isExecuting, stdin = '', onStdinChange, onClear }) => {
    const terminalEndRef = useRef<HTMLDivElement>(null);

    useEffect(() => {
        terminalEndRef.current?.scrollIntoView({ behavior: 'smooth' });
    }, [output]);

    return (
        <div className="terminal glass-panel">
            <div className="terminal-header">
                <div className="terminal-title">
                    <span className="terminal-icon">_</span>
                    <h3>Terminal</h3>
                </div>
                <button onClick={onClear} className="clear-btn" title="Clear Terminal">
                    Clear
                </button>
            </div>

            {onStdinChange && (
                <div className="terminal-input-section">
                    <label htmlFor="stdin-input" className="stdin-label">Program Input (stdin):</label>
                    <textarea
                        id="stdin-input"
                        className="stdin-input"
                        placeholder="Enter input for your program (e.g., for input() or scanf)"
                        value={stdin}
                        onChange={(e) => onStdinChange(e.target.value)}
                        rows={3}
                    />
                </div>
            )}

            <div className="terminal-content">
                {output.length === 0 && !isExecuting && (
                    <div className="terminal-welcome">
                        Output will appear here...
                    </div>
                )}

                {output.map((line, index) => (
                    <div key={index} className="terminal-line">
                        {line}
                    </div>
                ))}

                {isExecuting && (
                    <div className="terminal-line executing">
                        <span className="spinner-small"></span>
                        Running code...
                    </div>
                )}

                <div ref={terminalEndRef} />
            </div>
        </div>
    );
};
