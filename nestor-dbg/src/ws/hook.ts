import { useEffect, useState } from 'react';
import WS from '.';

let globalWSInstance: WS | null;

function getWsURL(): string {
  const { REACT_APP_NESTOR_ADDR } = process.env;
  const host = REACT_APP_NESTOR_ADDR !== "" ? REACT_APP_NESTOR_ADDR : window.location.host
  return "ws://" + host + window.location.pathname + "ws"
}

export default function useWS(): [WS | null, boolean] {
  const [ws, setWS] = useState<WS | null>(null);
  const [ready, setReady] = useState(false);
  const wsUrl = getWsURL()

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
