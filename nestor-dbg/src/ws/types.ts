import { CPUState } from '../types';

// Types for the WebSocket communication
export interface WSStateMessage {
  event: 'state';
  data: {
    cpu: CPUState;
  };
}

// All possible messages
export type WSMessage = WSStateMessage /* | OtherMessage */;
