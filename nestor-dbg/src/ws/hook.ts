import { useEffect, useState } from 'react';
import WS from '.';

let globalWSInstance: WS | null;

export default function useWS(
  wsUrl: string = 'ws://localhost:7777/ws'
): [WS | null, boolean] {
  const [ws, setWS] = useState<WS | null>(null);
  const [ready, setReady] = useState(false);

  useEffect(() => {
    if (globalWSInstance && globalWSInstance.settings.url !== wsUrl) {
      globalWSInstance.close();
      globalWSInstance = null;
    }

    if (!globalWSInstance) {
      globalWSInstance = new WS({
        url: wsUrl,
        debug: process.env.NODE_ENV === 'development',
        shouldReconnect: true
      });
    }

    setWS(globalWSInstance);
  }, [wsUrl]);

  useEffect(() => {
    if (!ws) return;

    return ws.onConnectionChange(setReady);
  }, [ws]);

  return [ws, ready];
}
