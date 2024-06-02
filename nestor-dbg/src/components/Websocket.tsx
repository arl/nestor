import React, { useEffect } from 'react';
import useWebSocket, { ReadyState } from 'react-use-websocket';

export interface NestorData {
    event: string;
    data: {};
}

interface WebSocketManagerProps {
    onMessage: (message: NestorData) => void;
}

export const WebSocketManager: React.FC<WebSocketManagerProps> = ({ onMessage }) => {
    const WS_URL = 'ws://127.0.0.1:7777';
    const { sendJsonMessage, lastJsonMessage, readyState } = useWebSocket(WS_URL, {
        share: false,
        shouldReconnect: () => true,
    });

    useEffect(() => {
        console.log('Connection state changed');
        if (readyState === ReadyState.OPEN) {
            sendJsonMessage({
                event: 'connected',
                data: {},
            });
        }
    }, [readyState, sendJsonMessage]);

    useEffect(() => {
        if (lastJsonMessage) {
            console.log(`received from nestor: ${JSON.stringify(lastJsonMessage)}`);
            onMessage(lastJsonMessage as NestorData);
        }
    }, [lastJsonMessage, onMessage]);

    return null;
};


