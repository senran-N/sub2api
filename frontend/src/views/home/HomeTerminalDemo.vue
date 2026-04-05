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
  background: linear-gradient(145deg, #1e293b 0%, #0f172a 100%);
  border-radius: 14px;
  box-shadow:
    0 25px 50px -12px rgba(0, 0, 0, 0.4),
    0 0 0 1px rgba(255, 255, 255, 0.1),
    inset 0 1px 0 rgba(255, 255, 255, 0.1);
  overflow: hidden;
  transform: perspective(1000px) rotateX(2deg) rotateY(-2deg);
  transition: transform 0.3s ease;
}

.terminal-window:hover {
  transform: perspective(1000px) rotateX(0deg) rotateY(0deg) translateY(-4px);
}

.terminal-header {
  display: flex;
  align-items: center;
  padding: 12px 16px;
  background: rgba(30, 41, 59, 0.8);
  border-bottom: 1px solid rgba(255, 255, 255, 0.05);
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
  background: #ef4444;
}

.btn-minimize {
  background: #eab308;
}

.btn-maximize {
  background: #22c55e;
}

.terminal-title {
  flex: 1;
  text-align: center;
  font-size: 12px;
  font-family: ui-monospace, monospace;
  color: #64748b;
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
  color: #22c55e;
  font-weight: bold;
}

.code-cmd {
  color: #38bdf8;
}

.code-flag {
  color: #a78bfa;
}

.code-url {
  color: #14b8a6;
}

.code-comment {
  color: #64748b;
  font-style: italic;
}

.code-success {
  color: #22c55e;
  background: rgba(34, 197, 94, 0.15);
  padding: 2px 8px;
  border-radius: 4px;
  font-weight: 600;
}

.code-response {
  color: #fbbf24;
}

.cursor {
  display: inline-block;
  width: 8px;
  height: 16px;
  background: #22c55e;
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

:deep(.dark) .terminal-window {
  box-shadow:
    0 25px 50px -12px rgba(0, 0, 0, 0.6),
    0 0 0 1px rgba(20, 184, 166, 0.2),
    0 0 40px rgba(20, 184, 166, 0.1),
    inset 0 1px 0 rgba(255, 255, 255, 0.1);
}
</style>
