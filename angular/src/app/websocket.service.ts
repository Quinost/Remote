import { DestroyRef, inject, Injectable } from "@angular/core";
import { BehaviorSubject, Subject } from "rxjs";

export interface WebSocketMessage {
  type: 'open_url' | 'send_button' | 'click_at' | 'scroll' | 'screenshot' | 'type_enter' ;
  payload: any;
}

@Injectable({
  providedIn: 'root'
})
export class WebsocketService {
  private destroyRef = inject(DestroyRef);

  private socket: WebSocket | null = null;
  private connectionStatus = new BehaviorSubject<boolean>(false);
  private messageSubject = new Subject<WebSocketMessage>();

  public isConnected$ = this.connectionStatus.asObservable();
  public messages$ = this.messageSubject.asObservable();

  constructor() {
    this.destroyRef.onDestroy(() => {
      this.disconnect();
      this.connectionStatus.complete();
      this.messageSubject.complete();
    });
  }

  public connect(url: string): void {
    if (this.socket && this.socket.readyState === WebSocket.OPEN) {
      return;
    }

    this.socket = new WebSocket(url);
    this.socket.onopen = () => {
      this.connectionStatus.next(true);
    };
    this.socket.onclose = () => {
      this.connectionStatus.next(false);
      this.socket = null;
    };
    this.socket.onerror = (error) => {
      console.error("WebSocket error:", error);
      this.connectionStatus.next(false);
    };
    this.socket.onmessage = (event) => {
      try {
        const message: WebSocketMessage = JSON.parse(event.data);
        this.messageSubject.next(message);
      } catch (error) {
        console.error("Error parsing WebSocket message:", error);
      }
    };
  }

  public disconnect(): void {
    if (this.socket) {
      this.socket.close();
      this.socket = null;
    }
  }

  public sendMessage(message: WebSocketMessage): void {
    if (!this.socket || this.socket.readyState !== WebSocket.OPEN) {
      console.error("WebSocket is not connected. Cannot send message.");
      return;
    }
    try {
      this.socket.send(JSON.stringify(message));
    } catch (error) {
      console.error("Error sending WebSocket message:", error);
    }
  }
}
