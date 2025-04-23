import { CommonModule } from '@angular/common';
import { Component, DestroyRef, inject, OnDestroy, OnInit } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { WebSocketMessage, WebsocketService } from './websocket.service';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';

@Component({
  selector: 'app-root',
  standalone: true,
  imports: [
    CommonModule,
    FormsModule
  ],
  templateUrl: './app.component.html',
  styleUrl: './app.component.scss'
})
export class AppComponent implements OnInit, OnDestroy {
  isConnected = false;
  screenshotData: string | null = null;
  urlInput: string = '';

  private destroyRef = inject(DestroyRef);
  private wsUrl = `ws://${window.location.hostname}:${window.location.port}/ws`

  constructor(private websocketService: WebsocketService) { }

  ngOnDestroy(): void {
    this.websocketService.disconnect();
  }

  ngOnInit(): void {
    this.websocketService.isConnected$
      .pipe(takeUntilDestroyed(this.destroyRef)).subscribe(
        status => {
          this.isConnected = status;
          if (!status) {
            this.screenshotData = null;
          }

          console.log('Connection status:', status);
        }
      );

    this.websocketService.messages$
      .pipe(takeUntilDestroyed(this.destroyRef)).subscribe(
        (message: WebSocketMessage) => {
          if (message.type === 'screenshot' && typeof message.payload === 'string') {
            this.screenshotData = message.payload;
          }
        }
      )
  }

  connectWebSocket(): void {
    this.websocketService.connect(this.wsUrl);
  }

  disconnectWebSocket(): void {
    this.websocketService.disconnect();
  }

  openUrl(): void {
    if (!this.urlInput && !this.isConnected)
      return
    const message: WebSocketMessage = {
      type: 'open_url',
      payload: this.urlInput
    }

    this.websocketService.sendMessage(message);
  }

  sendBtn(button: 'volume_up' | 'volume_down' | 'exit_fullscreen'): void {
    if (!this.isConnected)
      return;
    const message: WebSocketMessage = {
      type: 'send_button',
      payload: button
    }

    this.websocketService.sendMessage(message);
  }

  handleScreenshotClick(event: MouseEvent): void {
    if (!this.isConnected && !this.screenshotData)
      return;

    const target = event.target as HTMLImageElement;

    const naturalWidth = target.naturalWidth;
    const naturalHeight = target.naturalHeight;

    const displayWidth = target.clientWidth;
    const displayHeight = target.clientHeight;

    const scaleX = naturalWidth / displayWidth;
    const scaleY = naturalHeight / displayHeight;

    const clickX = Math.round(event.offsetX * scaleX);
    const clickY = Math.round(event.offsetY * scaleY);

    const message: WebSocketMessage = {
      type: 'click_at',
      payload: { x: clickX, y: clickY }
    }

    this.websocketService.sendMessage(message);
  }

  scrollPage(direction: 'up' | 'down'): void {
    if (!this.isConnected)
      return;
    const percentToScroll = 90;
    const message: WebSocketMessage = {
      type: 'scroll',
      payload: {
        direction: direction,
        percent: percentToScroll
      }
    }

    this.websocketService.sendMessage(message);
  }
}
