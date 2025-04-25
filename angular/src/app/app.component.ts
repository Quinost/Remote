
import { CommonModule } from '@angular/common';
import { Component, DestroyRef, inject, OnDestroy, OnInit, ElementRef, ViewChild } from '@angular/core';
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
  private wsUrl = `ws://${window.location.hostname}:${window.location.port}/ws`;

  @ViewChild('screenshotImg') screenshotImg!: ElementRef<HTMLImageElement>;

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
      );
  }

  connectWebSocket(): void {
    this.websocketService.connect(this.wsUrl);
  }

  disconnectWebSocket(): void {
    this.websocketService.disconnect();
  }

  openUrl(): void {
    if (!this.urlInput && !this.isConnected)
      return;
    const message: WebSocketMessage = {
      type: 'open_url',
      payload: this.urlInput
    };
    this.websocketService.sendMessage(message);
  }

  sendBtn(button: 'volume_up' | 'volume_down' | 'exit_fullscreen'): void {
    if (!this.isConnected)
      return;
    const message: WebSocketMessage = {
      type: 'send_button',
      payload: button
    };
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
    };
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
    };
    this.websocketService.sendMessage(message);
  }

  private touchStartX: number = 0;
private touchStartY: number = 0;
private touchStartTime: number = 0;
private isTouchMoved: boolean = false;
private readonly CLICK_THRESHOLD: number = 10; // Próg przemieszczenia w pikselach, poniżej którego gest jest interpretowany jako kliknięcie
private readonly SWIPE_THRESHOLD: number = 30;

  handleTouchStart(event: TouchEvent): void {
    if (!this.isConnected || !this.screenshotData)
      return;

    if (event.touches.length === 1) {
      this.touchStartX = event.touches[0].clientX;
      this.touchStartY = event.touches[0].clientY;
      this.touchStartTime = Date.now();
      this.isTouchMoved = false;
    }
  }

  handleTouchMove(event: TouchEvent): void {
    if (!this.isConnected || !this.screenshotData)
      return;

    if (event.touches.length === 1) {
      const deltaX = event.touches[0].clientX - this.touchStartX;
      const deltaY = event.touches[0].clientY - this.touchStartY;

      // Jeśli przemieszczenie jest większe niż próg, oznaczamy że to nie jest kliknięcie
      if (Math.abs(deltaX) > this.CLICK_THRESHOLD || Math.abs(deltaY) > this.CLICK_THRESHOLD) {
        this.isTouchMoved = true;

        // Zapobiegaj domyślnym akcjom przeglądarki tylko jeśli to potencjalny swipe
        event.preventDefault();
      }
    }
  }

  handleTouchEnd(event: TouchEvent): void {
    if (!this.isConnected || !this.screenshotData)
      return;

    if (event.changedTouches.length === 1) {
      const touchEndX = event.changedTouches[0].clientX;
      const touchEndY = event.changedTouches[0].clientY;
      const touchEndTime = Date.now();
      const duration = touchEndTime - this.touchStartTime;

      const deltaX = touchEndX - this.touchStartX;
      const deltaY = touchEndY - this.touchStartY;
      const distance = Math.sqrt(deltaX * deltaX + deltaY * deltaY);

      // Konwertuj współrzędne na rzeczywisty rozmiar ekranu
      const target = this.screenshotImg.nativeElement;
      const naturalWidth = target.naturalWidth;
      const naturalHeight = target.naturalHeight;
      const displayWidth = target.clientWidth;
      const displayHeight = target.clientHeight;
      const scaleX = naturalWidth / displayWidth;
      const scaleY = naturalHeight / displayHeight;

      // Jeśli to kliknięcie (nie było dużego przemieszczenia)
      if (!this.isTouchMoved && distance < this.CLICK_THRESHOLD) {
        const clickX = Math.round(touchEndX * scaleX);
        const clickY = Math.round(touchEndY * scaleY);

        const message: WebSocketMessage = {
          type: 'click_at',
          payload: { x: clickX, y: clickY }
        };

        this.websocketService.sendMessage(message);
      }
      // Jeśli to gest przeciągnięcia (swipe)
      else if (distance > this.SWIPE_THRESHOLD) {
        // Zapobiegaj domyślnym akcjom przeglądarki tylko dla swipe
        event.preventDefault();

        const startX = Math.round(this.touchStartX * scaleX);
        const startY = Math.round(this.touchStartY * scaleY);
        const endX = Math.round(touchEndX * scaleX);
        const endY = Math.round(touchEndY * scaleY);

        const message: WebSocketMessage = {
          type: 'scroll',
          payload: {
            startX: startX,
            startY: startY,
            endX: endX,
            endY: endY,
            duration: duration
          }
        };

        this.websocketService.sendMessage(message);
      }
    }
  }
}
