// Dialogates with the debugger server via WebSocket

import { WSRequest, WSResponse } from './types';

const WS_URL = 'ws://localhost:7777/ws';
const WS_RECONNECTION_TIMEOUT = 5000;
const WS_RECONNECTION_MAX_RETRIES = 5;

type WSEvents = {
  connectionChange: ((open: boolean) => void)[]; // Connection status change
  message: ((msg: WSResponse) => void)[]; // Received response from server
};

export interface WSSettings {
  url: string; // WebSocket URL
  debug: boolean; // Enable debug logs
  autoConnect: boolean; // Automatically connect on instantiation
  shouldReconnect: boolean | (() => boolean); // Reconnect on close
}

const defaultSettings: WSSettings = {
  url: WS_URL,
  debug: false,
  autoConnect: true,
  shouldReconnect: true
};

class WS {
  private retries = 0;
  public settings: WSSettings;
  private socket!: WebSocket | null;

  private events: WSEvents = {
    connectionChange: [],
    message: []
  };

  constructor(settings: Partial<WSSettings> = {}) {
    this.settings = { ...defaultSettings, ...settings };

    if (this.settings.autoConnect) this.connect();
    else this.socket = null;
  }

  private onSocketClose() {
    this.log('Connection closed');

    this.events.connectionChange.forEach((cb) => cb(false));

    const shouldReconnect =
      typeof this.settings.shouldReconnect === 'function'
        ? this.settings.shouldReconnect()
        : this.settings.shouldReconnect;

    if (shouldReconnect) {
      if (this.retries >= WS_RECONNECTION_MAX_RETRIES) {
        this.log('Max retries reached, not reconnecting');
        return;
      }

      this.log(`Reconnecting in ${WS_RECONNECTION_TIMEOUT / 1000}s...`);
      setTimeout(() => this.connect(), WS_RECONNECTION_TIMEOUT);
    }
  }

  // Logging
  // TODO: replace with a proper logger or global logger with global debug flag
  private log(...args: any[]) {
    if (!this.settings.debug) return;

    console.log('[WS]', ...args);
  }

  public connect() {
    if (this.socket) {
      this.log('Warning: already connected, closing existing connection');
      this.socket.close();
    }

    this.log('Connecting to', this.settings.url);

    this.socket = new WebSocket(this.settings.url);
    this.socket.onopen = () => {
      this.log('Connection established');
      this.retries = 0;
      this.events.connectionChange.forEach((cb) => cb(true));
    };
    this.socket.onmessage = (e: MessageEvent) => {
      const resp = JSON.parse(e.data) as WSResponse;
      this.log(`Received from nestor: (${resp.event})`, resp.data);
      this.events.message.forEach((cb) => cb(resp));
    };
    this.socket.onclose = this.onSocketClose.bind(this);
  }

  public close() {
    this.log('Gracefully closing');
    this.socket?.close();
  }

  // Event listeners

  private addEventListener(
    event: keyof WSEvents,
    callback: (...args: any[]) => void,
    callImmediately = true
  ) {
    this.events[event].push(callback);

    callImmediately && callback();

    return () => {
      // @ts-ignore
      this.events[event] = this.events[event].filter((cb) => cb !== callback);
    };
  }

  public onConnectionChange(
    callback: (open: boolean) => void,
    callImmediately = true
  ) {
    return this.addEventListener('connectionChange', callback, callImmediately);
  }

  public onMessage(
    callback: (event: WSRequest['event'], data: WSRequest['data']) => void,
    callImmediately = true
  ) {
    return this.addEventListener('message', callback, callImmediately);
  }

  /**
   * Shortcut for listening to a specific event
   *
   * @example
   * nestorWS.on('state', (data) => console.log('State:', data));
   * is equivalent to: nestorWS.onMessage((e, data) => { if (e === 'state') console.log('State:', data); });
   */
  public on(
    event: WSResponse['event'],
    callback: (data: WSResponse['data']) => void
  ) {
    return this.addEventListener('message', (resp) => {
      if (resp === undefined) {
        return;
      }
      if (resp.event === event) callback(resp.data);
    });
  }

  // Send messages

  // public send(event: WSMessage['event'], data: WSMessage['data']) {
  public send(req: WSRequest) {
    this.log(`Sending to nestor: (${req.event})`, req.data);
    this.socket?.send(JSON.stringify(req));
  }
}

export default WS;
