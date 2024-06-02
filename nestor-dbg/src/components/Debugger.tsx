import React, { useState } from 'react';
import { WebSocketManager, NestorData } from './Websocket';

// Define the DebuggerState enum
enum DebuggerState {
    Running = 'running',
    Paused = 'paused',
    Stepping = 'stepping'
}

// Define ButtonProps interface
interface ButtonProps {
    onClick: () => void;
    disabled: boolean;
    children: React.ReactNode;
}

// Button component
const Button: React.FC<ButtonProps> = ({ onClick, disabled, children }) => (
    <button onClick={onClick} disabled={disabled}>
        {children}
    </button>
);

// Debugger component
const Debugger: React.FC = () => {
    const [message, setMessage] = React.useState<NestorData | null>(null);

    const handleNewMessage = (newMessage: NestorData) => {
        setMessage(newMessage);
    };

    const [debuggerState, setDebuggerState] = useState<DebuggerState>(DebuggerState.Paused);

    const handleStart = () => setDebuggerState(DebuggerState.Running);
    const handlePause = () => setDebuggerState(DebuggerState.Paused);
    const handleStep = () => setDebuggerState(DebuggerState.Stepping);

    return (
        <div>
            <WebSocketManager onMessage={handleNewMessage} />
            <Button onClick={handleStart} disabled={debuggerState === DebuggerState.Running}>
                Start
            </Button>
            <Button onClick={handlePause} disabled={debuggerState !== DebuggerState.Running}>
                Pause
            </Button>
            <Button onClick={handleStep} disabled={debuggerState === DebuggerState.Running}>
                Step
            </Button>
            <div>
                <p>Debugger state: {debuggerState}</p>
            </div>
        </div>
    );
};

export default Debugger;
