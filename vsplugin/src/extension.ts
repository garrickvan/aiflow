import * as vscode from 'vscode';
import * as http from 'http';

/**
 * 服务配置常量
 */
const SERVER_CONFIG = {
  /** 默认端口 */
  DEFAULT_PORT: 9990,
  /** 页面路径 */
  PAGE_PATH: '/web/static?page=job',
  /** 检测超时时间(毫秒) */
  TIMEOUT_MS: 3000,
} as const;

/**
 * 检测本地服务是否运行
 * @param port 端口号
 * @param timeout 超时时间(毫秒)
 * @returns Promise<boolean> 服务是否可用
 */
function checkServer(port: number, timeout: number): Promise<boolean> {
  return new Promise((resolve) => {
    const request = http.get(
      {
        hostname: 'localhost',
        port: port,
        path: '/',
        timeout: timeout
      },
      (response) => {
        resolve(response.statusCode !== undefined);
      }
    );

    request.on('error', () => {
      resolve(false);
    });

    request.on('timeout', () => {
      request.destroy();
      resolve(false);
    });
  });
}

/**
 * 侧边栏 WebView 提供者
 */
class AiFlowSidebarProvider implements vscode.WebviewViewProvider {
  public static readonly viewType = 'aiflow.sidebar';
  private _view?: vscode.WebviewView;
  /** 是否正在连接中 */
  private _isConnecting: boolean = false;
  /** 当前连接的端口 */
  private _currentPort?: number;

  constructor(private readonly _extensionUri: vscode.Uri) {}

  public resolveWebviewView(
    webviewView: vscode.WebviewView,
    context: vscode.WebviewViewResolveContext,
    _token: vscode.CancellationToken
  ): void {
    this._view = webviewView;

    webviewView.webview.options = {
      enableScripts: true,
      localResourceRoots: [this._extensionUri]
    };

    // 初始显示输入端口页面
    webviewView.webview.html = this._getInputHtml();

    // 监听WebView消息
    webviewView.webview.onDidReceiveMessage(
      (message) => {
        switch (message.command) {
          case 'connect':
            const port = parseInt(message.port, 10);
            if (isNaN(port) || port < 1 || port > 65535) {
              vscode.window.showErrorMessage('请输入有效的端口号(1-65535)');
              return;
            }
            this._connectToPort(port);
            return;
          case 'disconnect':
            this._disconnect();
            return;
          case 'openInBrowser':
            if (message.url) {
              vscode.env.openExternal(vscode.Uri.parse(message.url));
            }
            return;
          case 'openInTab':
            if (message.url) {
              this._openInTab(message.url);
            }
            return;
        }
      }
    );
  }

  /**
   * 连接到指定端口
   * @param port 端口号
   */
  private async _connectToPort(port: number): Promise<void> {
    if (this._isConnecting || !this._view) {
      return;
    }

    this._isConnecting = true;
    this._view.webview.html = this._getConnectingHtml(port);

    const { TIMEOUT_MS, PAGE_PATH } = SERVER_CONFIG;

    const isRunning = await checkServer(port, TIMEOUT_MS);

    this._isConnecting = false;

    if (isRunning) {
      this._currentPort = port;
      this._loadPage(port, PAGE_PATH);
    } else {
      if (this._view) {
        this._view.webview.html = this._getErrorHtml(port);
      }
    }
  }

  /**
   * 断开连接，返回输入页面
   */
  private _disconnect(): void {
    this._currentPort = undefined;
    if (this._view) {
      this._view.webview.html = this._getInputHtml();
    }
  }

  /**
   * 加载指定端口和路径的页面到WebView
   * @param port 端口号
   * @param path 页面路径
   */
  private _loadPage(port: number, path: string): void {
    const url = `http://localhost:${port}${path}`;
    if (this._view) {
      this._view.webview.html = this._getIframeHtml(url, port);
    }
  }

  /**
   * 在代码Tab区打开页面
   * @param url 要打开的页面URL
   */
  private _openInTab(url: string): void {
    const panel = vscode.window.createWebviewPanel(
      'aiflowTab',
      'AiFlow',
      vscode.ViewColumn.One,
      {
        enableScripts: true,
        retainContextWhenHidden: true
      }
    );

    panel.webview.html = this._getTabHtml(url);
  }

  /**
   * 生成Tab页面的HTML
   * @param url 要加载的页面URL
   */
  private _getTabHtml(url: string): string {
    return `<!DOCTYPE html>
<html lang="zh-CN">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <style>
    * {
      margin: 0;
      padding: 0;
      box-sizing: border-box;
    }
    html, body {
      width: 100%;
      height: 100%;
      overflow: hidden;
    }
    .iframe-container {
      width: 100%;
      height: 100%;
    }
    iframe {
      width: 100%;
      height: 100%;
      border: none;
    }
  </style>
</head>
<body>
  <div class="iframe-container">
    <iframe id="appFrame" src="${url}" sandbox="allow-scripts allow-same-origin allow-forms allow-popups"></iframe>
  </div>
  <script>
    const appFrame = document.getElementById('appFrame');

    // 监听来自 iframe 的消息并转发
    window.addEventListener('message', async (event) => {
      // 只处理来自 iframe 的消息
      if (event.source !== appFrame.contentWindow) {
        return;
      }

      const message = event.data;
      if (message && message.command === 'copyToClipboard') {
        try {
          await navigator.clipboard.writeText(message.text);
          // 将结果返回给 iframe
          appFrame.contentWindow.postMessage(
            { command: 'copyToClipboardResult', success: true },
            '*'
          );
        } catch (err) {
          appFrame.contentWindow.postMessage(
            { command: 'copyToClipboardResult', success: false, error: err.message },
            '*'
          );
        }
      }
    });
  </script>
</body>
</html>`;
  }

  /**
   * 生成包含iframe的HTML，用于嵌入外部页面
   * @param url 要加载的页面URL
   * @param port 端口号
   */
  private _getIframeHtml(url: string, port: number): string {
    return `<!DOCTYPE html>
<html lang="zh-CN">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <style>
    * {
      margin: 0;
      padding: 0;
      box-sizing: border-box;
    }
    html, body {
      width: 100%;
      height: 100%;
      overflow: hidden;
    }
    .header {
      height: 36px;
      display: flex;
      align-items: center;
      justify-content: space-between;
      padding: 0 12px;
      background: var(--vscode-sideBar-background);
      border-bottom: 1px solid var(--vscode-panel-border);
    }
    .header-info {
      display: flex;
      align-items: center;
      gap: 8px;
      font-size: 12px;
      color: var(--vscode-foreground);
    }
    .status-dot {
      width: 8px;
      height: 8px;
      border-radius: 50%;
      background: #4caf50;
      box-shadow: 0 0 4px #4caf50;
    }
    .port-tag {
      font-family: monospace;
      padding: 2px 6px;
      background: var(--vscode-badge-background);
      color: var(--vscode-badge-foreground);
      border-radius: 3px;
      font-size: 10px;
    }
    .header-actions {
      display: flex;
      align-items: center;
      gap: 8px;
    }
    .open-tab-btn {
      padding: 4px 10px;
      font-size: 11px;
      background: var(--vscode-button-secondaryBackground);
      color: var(--vscode-button-secondaryForeground);
      border: none;
      border-radius: 3px;
      cursor: pointer;
    }
    .open-tab-btn:hover {
      background: var(--vscode-button-secondaryHoverBackground);
    }
    .open-browser-btn {
      padding: 4px 10px;
      font-size: 11px;
      background: var(--vscode-button-secondaryBackground);
      color: var(--vscode-button-secondaryForeground);
      border: none;
      border-radius: 3px;
      cursor: pointer;
    }
    .open-browser-btn:hover {
      background: var(--vscode-button-secondaryHoverBackground);
    }
    .disconnect-btn {
      padding: 4px 10px;
      font-size: 11px;
      background: var(--vscode-button-secondaryBackground);
      color: var(--vscode-button-secondaryForeground);
      border: none;
      border-radius: 3px;
      cursor: pointer;
    }
    .disconnect-btn:hover {
      background: var(--vscode-button-secondaryHoverBackground);
    }
    .iframe-container {
      width: 100%;
      height: calc(100% - 36px);
    }
    iframe {
      width: 100%;
      height: 100%;
      border: none;
    }
  </style>
</head>
<body>
  <div class="header">
    <div class="header-info">
      <div class="status-dot"></div>
      <span>已连接</span>
      <span class="port-tag">${port}</span>
    </div>
    <div class="header-actions">
      <button class="open-tab-btn" onclick="openInTab()">Tab</button>
      <button class="open-browser-btn" onclick="openInBrowser()">浏览器</button>
      <button class="disconnect-btn" onclick="disconnect()">断开</button>
    </div>
  </div>
  <div class="iframe-container">
    <iframe id="appFrame" src="${url}" sandbox="allow-scripts allow-same-origin allow-forms allow-popups"></iframe>
  </div>
  <script>
    const vscode = acquireVsCodeApi();
    const appFrame = document.getElementById('appFrame');

    function disconnect() {
      vscode.postMessage({ command: 'disconnect' });
    }

    function openInTab() {
      vscode.postMessage({ command: 'openInTab', url: '${url}' });
    }

    function openInBrowser() {
      vscode.postMessage({ command: 'openInBrowser', url: '${url}' });
    }

    // 监听来自 iframe 的消息并转发给 VS Code
    window.addEventListener('message', async (event) => {
      // 只处理来自 iframe 的消息
      if (event.source !== appFrame.contentWindow) {
        return;
      }

      const message = event.data;
      if (message && message.command === 'copyToClipboard') {
        try {
          await vscode.env.clipboard.writeText(message.text);
          // 将结果返回给 iframe
          appFrame.contentWindow.postMessage(
            { command: 'copyToClipboardResult', success: true },
            '*'
          );
        } catch (err) {
          appFrame.contentWindow.postMessage(
            { command: 'copyToClipboardResult', success: false, error: err.message },
            '*'
          );
        }
      }
    });
  </script>
</body>
</html>`;
  }

  /**
   * 生成端口输入HTML
   */
  private _getInputHtml(): string {
    const defaultPort = SERVER_CONFIG.DEFAULT_PORT;
    return `<!DOCTYPE html>
<html lang="zh-CN">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <style>
    * {
      margin: 0;
      padding: 0;
      box-sizing: border-box;
    }
    body {
      padding: 16px;
      font-family: var(--vscode-font-family);
      color: var(--vscode-foreground);
      background: var(--vscode-sideBar-background);
    }
    .header {
      font-size: 14px;
      font-weight: bold;
      margin-bottom: 16px;
      padding-bottom: 8px;
      border-bottom: 1px solid var(--vscode-panel-border);
    }
    .input-section {
      margin-bottom: 16px;
    }
    .input-label {
      font-size: 12px;
      margin-bottom: 6px;
      color: var(--vscode-foreground);
    }
    .port-input {
      width: 100%;
      padding: 8px 10px;
      font-size: 13px;
      font-family: monospace;
      background: var(--vscode-input-background);
      color: var(--vscode-input-foreground);
      border: 1px solid var(--vscode-input-border);
      border-radius: 4px;
      outline: none;
    }
    .port-input:focus {
      border-color: var(--vscode-focusBorder);
    }
    .hint-text {
      font-size: 11px;
      color: var(--vscode-descriptionForeground);
      margin-top: 6px;
    }
    .connect-btn {
      width: 100%;
      padding: 10px;
      font-size: 13px;
      font-weight: 500;
      background: var(--vscode-button-background);
      color: var(--vscode-button-foreground);
      border: none;
      border-radius: 4px;
      cursor: pointer;
      transition: background 0.2s;
    }
    .connect-btn:hover {
      background: var(--vscode-button-hoverBackground);
    }
    .connect-btn:active {
      transform: translateY(1px);
    }
  </style>
</head>
<body>
  <div class="input-section">
    <div class="input-label">管理服务端口号，默认端口: ${defaultPort}</div>
    <input type="number" class="port-input" id="portInput" value="${defaultPort}" placeholder="请输入端口号" min="1" max="65535">
  </div>

  <button class="connect-btn" onclick="connect()">连接服务</button>

  <script>
    const vscode = acquireVsCodeApi();
    const portInput = document.getElementById('portInput');

    // 支持回车键连接
    portInput.addEventListener('keypress', (e) => {
      if (e.key === 'Enter') {
        connect();
      }
    });

    function connect() {
      const port = portInput.value.trim();
      if (!port) {
        portInput.focus();
        return;
      }
      vscode.postMessage({ command: 'connect', port: port });
    }
  </script>
</body>
</html>`;
  }

  /**
   * 生成连接中HTML
   * @param port 正在连接的端口号
   */
  private _getConnectingHtml(port: number): string {
    return `<!DOCTYPE html>
<html lang="zh-CN">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <style>
    * {
      margin: 0;
      padding: 0;
      box-sizing: border-box;
    }
    body {
      padding: 16px;
      font-family: var(--vscode-font-family);
      color: var(--vscode-foreground);
      background: var(--vscode-sideBar-background);
      display: flex;
      flex-direction: column;
      align-items: center;
      justify-content: center;
      min-height: 200px;
    }
    .spinner {
      width: 32px;
      height: 32px;
      border: 3px solid var(--vscode-panel-border);
      border-top-color: var(--vscode-textLink-foreground);
      border-radius: 50%;
      animation: spin 1s linear infinite;
      margin-bottom: 16px;
    }
    @keyframes spin {
      to { transform: rotate(360deg); }
    }
    .loading-text {
      font-size: 13px;
      color: var(--vscode-descriptionForeground);
    }
    .loading-detail {
      font-size: 11px;
      color: var(--vscode-descriptionForeground);
      margin-top: 8px;
      opacity: 0.7;
    }
  </style>
</head>
<body>
  <div class="spinner"></div>
  <div class="loading-text">正在连接服务...</div>
  <div class="loading-detail">检测端口 ${port}</div>
</body>
</html>`;
  }

  /**
   * 生成连接失败HTML
   * @param port 连接失败的端口号
   */
  private _getErrorHtml(port: number): string {
    return `<!DOCTYPE html>
<html lang="zh-CN">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <style>
    * {
      margin: 0;
      padding: 0;
      box-sizing: border-box;
    }
    body {
      padding: 16px;
      font-family: var(--vscode-font-family);
      color: var(--vscode-foreground);
      background: var(--vscode-sideBar-background);
    }
    .header {
      font-size: 14px;
      font-weight: bold;
      margin-bottom: 16px;
      padding-bottom: 8px;
      border-bottom: 1px solid var(--vscode-panel-border);
    }
    .error-section {
      margin-bottom: 16px;
      padding: 12px;
      border-radius: 6px;
      border-left: 3px solid #f44336;
      background: rgba(244, 67, 54, 0.1);
    }
    .error-title {
      font-size: 13px;
      font-weight: bold;
      color: #f44336;
      margin-bottom: 8px;
      display: flex;
      align-items: center;
      gap: 6px;
    }
    .error-detail {
      font-size: 12px;
      color: var(--vscode-descriptionForeground);
      line-height: 1.5;
    }
    .port-tag {
      font-family: monospace;
      padding: 2px 6px;
      background: var(--vscode-badge-background);
      color: var(--vscode-badge-foreground);
      border-radius: 3px;
      font-size: 10px;
    }
    .action-section {
      display: flex;
      gap: 8px;
    }
    .back-btn {
      flex: 1;
      padding: 10px;
      font-size: 13px;
      font-weight: 500;
      background: var(--vscode-button-secondaryBackground);
      color: var(--vscode-button-secondaryForeground);
      border: none;
      border-radius: 4px;
      cursor: pointer;
    }
    .back-btn:hover {
      background: var(--vscode-button-secondaryHoverBackground);
    }
    .retry-btn {
      flex: 1;
      padding: 10px;
      font-size: 13px;
      font-weight: 500;
      background: var(--vscode-button-background);
      color: var(--vscode-button-foreground);
      border: none;
      border-radius: 4px;
      cursor: pointer;
    }
    .retry-btn:hover {
      background: var(--vscode-button-hoverBackground);
    }
    .hint-text {
      font-size: 11px;
      color: var(--vscode-descriptionForeground);
      margin-top: 12px;
      padding: 8px;
      background: var(--vscode-textBlockQuote-background);
      border-radius: 3px;
      line-height: 1.5;
    }
  </style>
</head>
<body>
  <div class="error-section">
    <div class="error-title">✗ 连接失败</div>
    <div class="error-detail">
      无法连接到端口 <span class="port-tag">${port}</span>
    </div>
  </div>

  <div class="action-section">
    <button class="back-btn" onclick="back()">返回</button>
    <button class="retry-btn" onclick="retry()">重试</button>
  </div>

  <div class="hint-text">
    请确保Aiflow的Go服务进程已启动并监听该端口。<br>
  </div>

  <script>
    const vscode = acquireVsCodeApi();
    const port = ${port};

    function back() {
      vscode.postMessage({ command: 'disconnect' });
    }

    function retry() {
      vscode.postMessage({ command: 'connect', port: port.toString() });
    }
  </script>
</body>
</html>`;
  }
}

/**
 * 激活插件
 * @param context 插件上下文
 */
export function activate(context: vscode.ExtensionContext): void {
  // 注册侧边栏 WebView 提供者
  const sidebarProvider = new AiFlowSidebarProvider(context.extensionUri);

  context.subscriptions.push(
    vscode.window.registerWebviewViewProvider(
      AiFlowSidebarProvider.viewType,
      sidebarProvider
    )
  );

  // 注册刷新命令
  const refreshCommand = vscode.commands.registerCommand(
    'aiflow.refreshServers',
    () => {
      vscode.window.showInformationMessage('正在刷新服务状态...');
    }
  );

  context.subscriptions.push(refreshCommand);
}

/**
 * 停用插件
 */
export function deactivate(): void {
  // 清理资源
}
