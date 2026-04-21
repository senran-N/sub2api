<template>
  <div class="terminal-container">
    <div class="terminal-window">
      <div class="terminal-header">
        <div class="terminal-buttons">
          <span class="btn-close"></span>
          <span class="btn-minimize"></span>
          <span class="btn-maximize"></span>
        </div>
        <span class="terminal-title">terminal</span>
      </div>
      <div class="terminal-body">
        <div class="code-line line-1">
          <span class="code-prompt">$</span>
          <span class="code-cmd">curl</span>
          <span class="code-flag">-X POST</span>
          <span class="code-url">/v1/messages</span>
        </div>
        <div class="code-line line-2">
          <span class="code-comment"># Routing to upstream...</span>
        </div>
        <div class="code-line line-3">
          <span class="code-success">200 OK</span>
          <span class="code-response">{ "content": "Hello!" }</span>
        </div>
        <div class="code-line line-4">
          <span class="code-prompt">$</span>
          <span class="cursor"></span>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.terminal-container {
  position: relative;
  display: inline-block;
  max-width: 100%;
  width: 100%;
}

.terminal-window {
  width: min(420px, 100%);
  background: linear-gradient(
    145deg,
    color-mix(in srgb, var(--theme-surface-contrast) 88%, var(--theme-page-bg) 12%) 0%,
    color-mix(in srgb, var(--theme-surface-contrast) 82%, var(--theme-accent) 18%) 100%
  );
  border-radius: 14px;
  box-shadow:
    0 25px 50px -12px color-mix(in srgb, var(--theme-overlay-strong) 48%, transparent),
    0 0 0 1px color-mix(in srgb, var(--theme-filled-text) 10%, transparent),
    inset 0 1px 0 color-mix(in srgb, var(--theme-filled-text) 10%, transparent);
  overflow: hidden;
  transform: perspective(1000px) rotateX(2deg) rotateY(-2deg);
  transition: transform 0.3s ease, box-shadow 0.3s ease;
}

.terminal-window:hover {
  transform: perspective(1000px) rotateX(0deg) rotateY(0deg) translateY(-4px);
}

/* Factory: flatten the perspective and use a hard offset shadow instead of a soft drop. */
:root[data-brand-theme='factory'] .terminal-window {
  border-radius: 0;
  border: 2px solid var(--theme-page-text);
  box-shadow: 8px 8px 0 var(--theme-page-text);
  transform: none;
}

:root[data-brand-theme='factory'] .terminal-window:hover {
  transform: translate(-2px, -2px);
  box-shadow: 10px 10px 0 var(--theme-page-text);
}

.dark[data-brand-theme='factory'] .terminal-window {
  border-color: rgba(255, 255, 255, 0.35);
  box-shadow: 8px 8px 0 rgba(255, 255, 255, 0.2);
}

.dark[data-brand-theme='factory'] .terminal-window:hover {
  box-shadow: 10px 10px 0 rgba(255, 255, 255, 0.3);
}

/* Claude: softer perspective + warm glow ring. */
:root[data-brand-theme='claude'] .terminal-window {
  transform: perspective(1000px) rotateX(1deg) rotateY(-1deg);
  box-shadow:
    0 30px 60px -20px color-mix(in srgb, var(--theme-accent) 28%, transparent),
    0 0 0 1px color-mix(in srgb, var(--theme-accent) 24%, transparent);
}

:root[data-brand-theme='claude'] .terminal-window:hover {
  transform: perspective(1000px) rotateX(0deg) rotateY(0deg) translateY(-6px);
  box-shadow:
    0 40px 80px -24px color-mix(in srgb, var(--theme-accent) 38%, transparent),
    0 0 0 1px color-mix(in srgb, var(--theme-accent) 34%, transparent);
}

.terminal-header {
  display: flex;
  align-items: center;
  padding: 12px 16px;
  background: color-mix(in srgb, var(--theme-surface-contrast) 80%, transparent);
  border-bottom: 1px solid color-mix(in srgb, var(--theme-filled-text) 5%, transparent);
}

.terminal-buttons {
  display: flex;
  gap: 8px;
}

.terminal-buttons span {
  width: 12px;
  height: 12px;
  border-radius: 50%;
}

.btn-close {
  background: rgb(var(--theme-danger-rgb));
}

.btn-minimize {
  background: rgb(var(--theme-warning-rgb));
}

.btn-maximize {
  background: rgb(var(--theme-success-rgb));
}

.terminal-title {
  flex: 1;
  text-align: center;
  font-size: 12px;
  font-family: ui-monospace, monospace;
  color: color-mix(in srgb, var(--theme-filled-text) 42%, transparent);
  margin-right: 52px;
}

.terminal-body {
  padding: 20px 24px;
  font-family: ui-monospace, 'Fira Code', monospace;
  font-size: 14px;
  line-height: 2;
}

@media (max-width: 480px) {
  .terminal-body {
    padding: 14px 16px;
    font-size: 11px;
    line-height: 1.8;
  }

  .terminal-header {
    padding: 10px 12px;
  }

  .terminal-buttons span {
    width: 10px;
    height: 10px;
  }
}

.code-line {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
  opacity: 0;
  animation: line-appear 0.5s ease forwards;
}

.line-1 {
  animation-delay: 0.3s;
}

.line-2 {
  animation-delay: 1s;
}

.line-3 {
  animation-delay: 1.8s;
}

.line-4 {
  animation-delay: 2.5s;
}

@keyframes line-appear {
  from {
    opacity: 0;
    transform: translateY(5px);
  }

  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.code-prompt {
  color: rgb(var(--theme-success-rgb));
  font-weight: bold;
}

.code-cmd {
  color: rgb(var(--theme-info-rgb));
}

.code-flag {
  color: rgb(var(--theme-brand-purple-rgb));
}

.code-url {
  color: var(--theme-accent);
}

.code-comment {
  color: color-mix(in srgb, var(--theme-filled-text) 42%, transparent);
  font-style: italic;
}

.code-success {
  color: rgb(var(--theme-success-rgb));
  background: color-mix(in srgb, rgb(var(--theme-success-rgb)) 15%, transparent);
  padding: 2px 8px;
  border-radius: 4px;
  font-weight: 600;
}

.code-response {
  color: color-mix(in srgb, rgb(var(--theme-warning-rgb)) 78%, var(--theme-filled-text));
}

.cursor {
  display: inline-block;
  width: 8px;
  height: 16px;
  background: rgb(var(--theme-success-rgb));
  animation: blink 1s step-end infinite;
}

@keyframes blink {
  0%,
  50% {
    opacity: 1;
  }

  51%,
  100% {
    opacity: 0;
  }
}

.terminal-window:hover {
  box-shadow:
    0 25px 50px -12px color-mix(in srgb, var(--theme-overlay-strong) 56%, transparent),
    0 0 0 1px color-mix(in srgb, var(--theme-accent) 20%, transparent),
    0 0 40px color-mix(in srgb, var(--theme-accent) 10%, transparent),
    inset 0 1px 0 color-mix(in srgb, var(--theme-filled-text) 10%, transparent);
}
</style>
