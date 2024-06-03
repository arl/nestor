import React, { useState, useEffect } from 'react';
import useWebSocket, { ReadyState } from 'react-use-websocket';

enum CPUState {
    Running = 'running',
    Paused = 'paused',
    Stepping = 'stepping'
}

export interface NestorData {
    event: string;
    data: {};
}

export interface DebuggerState {
    cpu: CPUState;
}

interface WebSocketManagerProps {
    onMessage: (message: NestorData) => void;
    onStateReceived: (state: DebuggerState) => void;
}

export const WebSocketManager: React.FC<WebSocketManagerProps> = ({ onMessage, onStateReceived }) => {
    const WS_URL = 'ws://localhost:7777/ws';
    const { sendJsonMessage, lastJsonMessage, readyState } = useWebSocket(WS_URL, {
        share: false,
        shouldReconnect: () => true,
    });

    useEffect(() => {
        console.log('Connection state changed');
        if (readyState === ReadyState.OPEN) {
            console.log('WebSocket connection opened');
        }
    }, [readyState, sendJsonMessage]);

    useEffect(() => {
        if (lastJsonMessage) {
            console.log(`received from nestor: "`, JSON.stringify(lastJsonMessage));
            const parsedMessage = lastJsonMessage as NestorData;
            if (parsedMessage.event === 'state') {
                const state = parsedMessage.data as DebuggerState;
                onStateReceived(state);
            }
            onMessage(parsedMessage);
        }
    }, [lastJsonMessage, onMessage, onStateReceived]);

    return null;
};




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
    const [debuggerState, setDebuggerState] = useState<DebuggerState | null>(null);

    const handleNewMessage = (newMessage: NestorData) => {
        setMessage(newMessage);
    };

    const handleStateReceived = (state: DebuggerState) => {
        setDebuggerState(state);
    };

    const handleStart = () => { /*setDebuggerState(CPUState.Running);*/ };
    const handlePause = () => { /*setDebuggerState(CPUState.Paused);*/ };
    const handleStep = () => { /*setDebuggerState(CPUState.Stepping);*/ };

    return (
        <div>
            <WebSocketManager onMessage={handleNewMessage} onStateReceived={handleStateReceived} />
            <Button onClick={handleStart} disabled={debuggerState?.cpu === CPUState.Running}>
                Start
            </Button>
            <Button onClick={handlePause} disabled={debuggerState?.cpu !== CPUState.Running}>
                Pause
            </Button>
            <Button onClick={handleStep} disabled={debuggerState?.cpu === CPUState.Running}>
                Step
            </Button>
            <div>
                <p>Debugger state: {debuggerState?.cpu}</p>
            </div>
        </div>
    );
};

export default Debugger;

