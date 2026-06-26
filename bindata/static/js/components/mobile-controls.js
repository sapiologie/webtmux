// Mobile controls component
import { LitElement, html, css } from 'lit';

class WebtmuxMobileControls extends LitElement {
  static properties = {
    showPaneSelector: { type: Boolean },
    showSessionSelector: { type: Boolean },
    layout: { type: Object },
    recording: { type: Boolean },
    voiceStatus: { type: String },
    showHelp: { type: Boolean },
  };

  static styles = css`
    :host {
      display: block;
      position: relative;
      background: #16213e;
      border-top: 1px solid #0f3460;
      padding: 6px;
      padding-bottom: calc(6px + env(safe-area-inset-bottom));
      z-index: 1000;
    }

    .controls {
      display: flex;
      justify-content: space-around;
      gap: 8px;
    }

    .control-btn {
      flex: 1;
      max-width: 60px;
      background: #1a1a2e;
      border: 1px solid #0f3460;
      border-radius: 6px;
      color: #888;
      padding: 8px 4px;
      font-size: 9px;
      cursor: pointer;
      transition: all 0.2s;
      display: flex;
      flex-direction: column;
      align-items: center;
      gap: 2px;
      -webkit-tap-highlight-color: transparent;
    }

    .control-btn:active {
      background: #0f3460;
      border-color: #e94560;
      color: #fff;
      transform: scale(0.95);
    }

    .control-btn svg {
      width: 16px;
      height: 16px;
    }

    .control-btn.prefix {
      background: #e94560;
      border-color: #e94560;
      color: #fff;
    }

    .window-tabs {
      display: flex;
      gap: 4px;
      margin-bottom: 8px;
      overflow-x: auto;
      padding-bottom: 4px;
      -webkit-overflow-scrolling: touch;
    }

    .window-tab {
      flex-shrink: 0;
      background: #1a1a2e;
      border: 1px solid #0f3460;
      border-radius: 4px;
      color: #888;
      padding: 6px 12px;
      font-size: 11px;
      cursor: pointer;
      white-space: nowrap;
    }

    .window-tab:active, .window-tab.active {
      background: #e94560;
      border-color: #e94560;
      color: #fff;
    }

    .pane-selector {
      position: absolute;
      bottom: 100%;
      left: 0;
      right: 0;
      background: #16213e;
      border-top: 1px solid #0f3460;
      padding: 12px;
      display: none;
    }

    .pane-selector.open {
      display: block;
    }

    .pane-grid {
      display: grid;
      grid-template-columns: repeat(3, 1fr);
      gap: 8px;
    }

    .pane-btn {
      background: #1a1a2e;
      border: 1px solid #0f3460;
      border-radius: 4px;
      color: #888;
      padding: 12px;
      font-size: 12px;
      cursor: pointer;
    }

    .pane-btn.active {
      border-color: #e94560;
      color: #e94560;
    }

    .arrow-pad {
      display: grid;
      grid-template-columns: repeat(3, 1fr);
      grid-template-rows: repeat(3, 1fr);
      gap: 2px;
      width: 90px;
      height: 90px;
    }

    .arrow-btn {
      background: #1a1a2e;
      border: 1px solid #0f3460;
      border-radius: 4px;
      color: #888;
      display: flex;
      align-items: center;
      justify-content: center;
      cursor: pointer;
      font-size: 14px;
    }

    .arrow-btn:active {
      background: #0f3460;
      border-color: #e94560;
    }

    .arrow-btn.empty {
      visibility: hidden;
    }

    .session-overlay {
      position: fixed;
      top: 0;
      left: 0;
      right: 0;
      bottom: 0;
      background: rgba(0, 0, 0, 0.8);
      display: none;
      align-items: center;
      justify-content: center;
      z-index: 2000;
    }

    .session-overlay.open {
      display: flex;
    }

    .session-modal {
      background: #16213e;
      border: 1px solid #0f3460;
      border-radius: 12px;
      padding: 20px;
      min-width: 280px;
      max-width: 90%;
      max-height: 70vh;
      overflow-y: auto;
    }

    .session-modal h3 {
      color: #4a9eff;
      font-size: 14px;
      margin: 0 0 16px 0;
      text-align: center;
    }

    .session-list {
      display: flex;
      flex-direction: column;
      gap: 8px;
    }

    .session-item {
      background: #1a1a2e;
      border: 1px solid #0f3460;
      border-radius: 8px;
      padding: 12px 16px;
      color: #888;
      font-size: 14px;
      cursor: pointer;
      display: flex;
      justify-content: space-between;
      align-items: center;
    }

    .session-item:active {
      background: #0f3460;
      border-color: #4a9eff;
    }

    .session-item.active {
      border-color: #4a9eff;
      color: #fff;
      background: #1a3a5c;
    }

    .session-item .session-name {
      font-weight: 500;
    }

    .session-item .session-meta {
      font-size: 12px;
      opacity: 0.7;
    }

    .close-overlay {
      position: absolute;
      top: 20px;
      right: 20px;
      background: transparent;
      border: none;
      color: #888;
      font-size: 24px;
      cursor: pointer;
    }

    .session-btn {
      background: #4a9eff;
      border-color: #4a9eff;
      color: #fff;
    }

    .control-btn.mic {
      background: #2d6a4f;
      border-color: #2d6a4f;
      color: #fff;
      touch-action: none;
    }

    .control-btn.mic.recording {
      background: #e94560;
      border-color: #e94560;
      animation: mic-pulse 1s ease-in-out infinite;
    }

    @keyframes mic-pulse {
      0%, 100% { opacity: 1; }
      50% { opacity: 0.5; }
    }

    .voice-toast {
      position: absolute;
      bottom: 100%;
      left: 8px;
      right: 8px;
      margin-bottom: 8px;
      background: #16213e;
      border: 1px solid #0f3460;
      border-radius: 8px;
      color: #eaeaea;
      padding: 10px 14px;
      font-size: 12px;
      line-height: 1.4;
      box-shadow: 0 4px 16px rgba(0, 0, 0, 0.4);
      word-break: break-word;
    }

    .help-list {
      display: flex;
      flex-direction: column;
      gap: 8px;
    }

    .help-item {
      display: flex;
      flex-direction: column;
      gap: 2px;
      background: #1a1a2e;
      border: 1px solid #0f3460;
      border-radius: 8px;
      padding: 8px 12px;
    }

    .help-say {
      color: #fff;
      font-size: 13px;
    }

    .help-do {
      color: #888;
      font-size: 12px;
    }

    /* On touch devices the bar sits at the top of the screen (see index.html),
       so pop the toast and pane selector downward into the screen instead of
       upward off the top edge. */
    @media (pointer: coarse) {
      .voice-toast {
        top: 100%;
        bottom: auto;
        margin-top: 8px;
        margin-bottom: 0;
      }

      .pane-selector {
        top: 100%;
        bottom: auto;
        border-top: none;
        border-bottom: 1px solid #0f3460;
      }
    }
  `;

  constructor() {
    super();
    this.showPaneSelector = false;
    this.showSessionSelector = false;
    this.layout = null;
    this.recording = false;
    this.voiceStatus = '';
    this.showHelp = false;
    this._mediaRecorder = null;
    this._chunks = [];
    this._stream = null;
    this._statusTimer = null;

    window.addEventListener('tmux-layout-update', (e) => {
      this.layout = e.detail;
    });
  }

  render() {
    const sessions = this.layout?.sessions || [];
    const showSessionBtn = sessions.length > 1;

    return html`
      <!-- Session overlay -->
      <div class="session-overlay ${this.showSessionSelector ? 'open' : ''}" @click=${this.closeSessionSelector}>
        <div class="session-modal" @click=${(e) => e.stopPropagation()}>
          <h3>Switch Session</h3>
          <div class="session-list">
            ${sessions.map(sess => html`
              <button
                class="session-item ${sess.active ? 'active' : ''}"
                @click=${() => this.switchSession(sess.name)}
              >
                <span class="session-name">${sess.name}</span>
                <span class="session-meta">${sess.windows} window${sess.windows !== 1 ? 's' : ''}</span>
              </button>
            `)}
          </div>
        </div>
      </div>

      <!-- Help overlay: available voice commands -->
      <div class="session-overlay ${this.showHelp ? 'open' : ''}" @click=${this.closeHelp}>
        <div class="session-modal" @click=${(e) => e.stopPropagation()}>
          <h3>Voice commands</h3>
          <div class="help-list">
            <div class="help-item"><span class="help-say">Just speak</span><span class="help-do">types exactly what you say</span></div>
            <div class="help-item"><span class="help-say">"submit" / "press enter"</span><span class="help-do">presses Enter</span></div>
            <div class="help-item"><span class="help-say">"scroll up" / "scroll down"</span><span class="help-do">scrolls the history</span></div>
            <div class="help-item"><span class="help-say">"switch to session two"</span><span class="help-do">switches tmux session</span></div>
            <div class="help-item"><span class="help-say">"copy the screen" / "copy all"</span><span class="help-do">copies to the clipboard</span></div>
            <div class="help-item"><span class="help-say">"paste"</span><span class="help-do">pastes the clipboard</span></div>
            <div class="help-item"><span class="help-say">"press escape" / "interrupt"</span><span class="help-do">Esc / Ctrl-C</span></div>
            <div class="help-item"><span class="help-say">"press tab"</span><span class="help-do">Tab</span></div>
          </div>
        </div>
      </div>

      <div class="pane-selector ${this.showPaneSelector ? 'open' : ''}">
        <div class="pane-grid">
          ${this.layout?.windows?.find(w => w.active)?.panes?.map(pane => html`
            <button
              class="pane-btn ${pane.active ? 'active' : ''}"
              @click=${() => this.selectPane(pane.id)}
            >
              Pane ${pane.index}
            </button>
          `)}
        </div>
      </div>

      ${this.layout?.windows?.length > 0 ? html`
        <div class="window-tabs">
          ${this.layout.windows.map(win => html`
            <button
              class="window-tab ${win.active ? 'active' : ''}"
              @click=${() => this.selectWindow(win.id)}
            >
              ${win.index}: ${win.name || 'bash'}
            </button>
          `)}
          <button class="window-tab" @click=${this.newWindow}>+</button>
        </div>
      ` : ''}

      ${this.voiceStatus ? html`<div class="voice-toast">${this.voiceStatus}</div>` : ''}

      <div class="controls">
        <button
          class="control-btn mic ${this.recording ? 'recording' : ''}"
          @pointerdown=${this.startRecording}
          @pointerup=${this.stopRecording}
          @pointercancel=${this.stopRecording}
          @contextmenu=${(e) => e.preventDefault()}
        >
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <rect x="9" y="2" width="6" height="11" rx="3"/>
            <path d="M5 10v1a7 7 0 0 0 14 0v-1"/>
            <line x1="12" y1="19" x2="12" y2="22"/>
          </svg>
          ${this.recording ? 'Rec' : 'Voice'}
        </button>

        <button class="control-btn" @click=${this.toggleHelp}>
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <circle cx="12" cy="12" r="10"/>
            <line x1="12" y1="16" x2="12" y2="12"/>
            <line x1="12" y1="8" x2="12.01" y2="8"/>
          </svg>
          Help
        </button>

        ${showSessionBtn ? html`
          <button class="control-btn session-btn" @click=${this.toggleSessionSelector}>
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <rect x="2" y="3" width="20" height="14" rx="2"/>
              <line x1="8" y1="21" x2="16" y2="21"/>
              <line x1="12" y1="17" x2="12" y2="21"/>
            </svg>
            Sess
          </button>
        ` : ''}

        <button class="control-btn prefix" @click=${this.sendPrefix}>
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <rect x="3" y="3" width="18" height="18" rx="2"/>
            <text x="12" y="16" font-size="10" fill="currentColor" text-anchor="middle">^B</text>
          </svg>
          Prefix
        </button>

        <button class="control-btn" @click=${() => this.splitPane(true)}>
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <rect x="3" y="3" width="18" height="18" rx="2"/>
            <line x1="12" y1="3" x2="12" y2="21"/>
          </svg>
          Split H
        </button>

        <button class="control-btn" @click=${() => this.splitPane(false)}>
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <rect x="3" y="3" width="18" height="18" rx="2"/>
            <line x1="3" y1="12" x2="21" y2="12"/>
          </svg>
          Split V
        </button>

        <button class="control-btn" @click=${this.togglePaneSelector}>
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <rect x="3" y="3" width="7" height="7"/>
            <rect x="14" y="3" width="7" height="7"/>
            <rect x="3" y="14" width="7" height="7"/>
            <rect x="14" y="14" width="7" height="7"/>
          </svg>
          Panes
        </button>

        <button class="control-btn" @click=${this.newWindow}>
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <rect x="3" y="3" width="18" height="18" rx="2"/>
            <line x1="12" y1="8" x2="12" y2="16"/>
            <line x1="8" y1="12" x2="16" y2="12"/>
          </svg>
          New
        </button>
      </div>
    `;
  }

  sendPrefix() {
    // Send Ctrl+B (tmux prefix)
    // ASCII code for Ctrl+B is 0x02
    window.webtmux?.terminal?.input('\x02');
  }

  splitPane(horizontal) {
    window.webtmux?.splitPane(horizontal);
  }

  togglePaneSelector() {
    this.showPaneSelector = !this.showPaneSelector;
  }

  selectPane(paneId) {
    window.webtmux?.selectPane(paneId);
    this.showPaneSelector = false;
  }

  selectWindow(windowId) {
    window.webtmux?.selectWindow(windowId);
  }

  newWindow() {
    window.webtmux?.newWindow();
  }

  toggleSessionSelector() {
    this.showSessionSelector = !this.showSessionSelector;
  }

  closeSessionSelector() {
    this.showSessionSelector = false;
  }

  switchSession(sessionName) {
    window.webtmux?.switchSession(sessionName);
    this.showSessionSelector = false;
  }

  // --- Voice (push-to-talk) ---------------------------------------------------

  async startRecording(e) {
    if (e) e.preventDefault();
    if (this.recording) return;

    // Keep receiving pointer events even if the finger slides off the button.
    if (e && e.pointerId != null && e.target.setPointerCapture) {
      try { e.target.setPointerCapture(e.pointerId); } catch (_) {}
    }

    if (!navigator.mediaDevices || !navigator.mediaDevices.getUserMedia) {
      this.flashStatus('Mic unavailable (needs HTTPS)');
      return;
    }

    try {
      this._stream = await navigator.mediaDevices.getUserMedia({ audio: true });
    } catch (err) {
      this.flashStatus('Mic access denied');
      return;
    }

    this._chunks = [];
    let mimeType = '';
    const candidates = ['audio/webm;codecs=opus', 'audio/webm', 'audio/mp4', 'audio/aac'];
    for (const c of candidates) {
      if (window.MediaRecorder && MediaRecorder.isTypeSupported && MediaRecorder.isTypeSupported(c)) {
        mimeType = c;
        break;
      }
    }

    try {
      this._mediaRecorder = mimeType
        ? new MediaRecorder(this._stream, { mimeType })
        : new MediaRecorder(this._stream);
    } catch (err) {
      this._mediaRecorder = new MediaRecorder(this._stream);
    }

    this._mediaRecorder.ondataavailable = (ev) => {
      if (ev.data && ev.data.size > 0) this._chunks.push(ev.data);
    };
    this._mediaRecorder.onstop = () => this.uploadRecording();
    this._mediaRecorder.start();
    this.recording = true;
  }

  stopRecording(e) {
    if (e) e.preventDefault();
    if (!this.recording) return;
    this.recording = false;
    if (this._mediaRecorder && this._mediaRecorder.state !== 'inactive') {
      this._mediaRecorder.stop();
    }
  }

  async uploadRecording() {
    if (this._stream) {
      this._stream.getTracks().forEach(t => t.stop());
      this._stream = null;
    }
    if (!this._chunks.length) return;

    const type = (this._mediaRecorder && this._mediaRecorder.mimeType) || 'audio/webm';
    const blob = new Blob(this._chunks, { type });
    this._chunks = [];
    const ext = (type.includes('mp4') || type.includes('aac')) ? 'mp4' : 'webm';

    const form = new FormData();
    form.append('audio', blob, `voice.${ext}`);

    this.flashStatus('Transcribing...', 0);
    try {
      const resp = await fetch('./voice', { method: 'POST', body: form });
      if (!resp.ok) {
        this.flashStatus('Voice error: ' + resp.status);
        return;
      }
      const data = await resp.json();
      this.flashStatus(`"${data.transcript}" -> ${this.describeAction(data.action)}`);
      window.webtmux?.dispatchVoiceAction(data.action);
    } catch (err) {
      this.flashStatus('Voice request failed');
    }
  }

  describeAction(a) {
    if (!a) return 'nothing';
    switch (a.type) {
      case 'dictate': return 'type';
      case 'submit': return 'submit';
      case 'key': return 'key ' + a.name;
      case 'switch_session': return 'session ' + a.target;
      case 'select_window': return 'window ' + a.index;
      case 'scroll': return 'scroll ' + a.dir;
      case 'copy': return 'copy';
      case 'paste': return 'paste';
      default: return a.type;
    }
  }

  flashStatus(msg, timeout = 4000) {
    this.voiceStatus = msg;
    if (this._statusTimer) clearTimeout(this._statusTimer);
    if (timeout > 0) {
      this._statusTimer = setTimeout(() => { this.voiceStatus = ''; }, timeout);
    }
  }

  toggleHelp() {
    this.showHelp = !this.showHelp;
  }

  closeHelp() {
    this.showHelp = false;
  }
}

customElements.define('webtmux-mobile-controls', WebtmuxMobileControls);
