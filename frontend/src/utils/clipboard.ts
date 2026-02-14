/**
 * 剪贴板相关工具函数
 */

import { message } from 'antd';

/**
 * 复制文本到剪贴板
 * @param text - 要复制的文本
 * @param successMsg - 成功提示消息
 */
export const copyToClipboard = async (
  text: string,
  successMsg: string = '已复制到剪贴板',
): Promise<void> => {
  /**
   * 降级复制方案（用于不支持 Clipboard API 的环境）
   */
  const fallbackCopy = (): boolean => {
    const textArea = document.createElement('textarea');
    textArea.value = text;
    textArea.style.position = 'fixed';
    textArea.style.left = '-999999px';
    textArea.style.top = '-999999px';
    document.body.appendChild(textArea);
    textArea.focus();
    textArea.select();

    const successful = document.execCommand('copy');
    document.body.removeChild(textArea);
    return successful;
  };

  try {
    // 检查是否在 iframe 中（VSCode Webview 环境）
    if (window.parent !== window) {
      window.parent.postMessage({
        command: 'copyToClipboard',
        text: text,
      }, '*');

      const result = await Promise.race([
        new Promise<{ success: boolean; error?: string }>((resolve) => {
          const handler = (event: MessageEvent) => {
            if (event.data && event.data.command === 'copyToClipboardResult') {
              window.removeEventListener('message', handler);
              resolve({ success: event.data.success, error: event.data.error });
            }
          };
          window.addEventListener('message', handler);
        }),
        new Promise<{ success: boolean }>((_, reject) =>
          setTimeout(() => reject(new Error('timeout')), 1000),
        ),
      ]);

      if (result.success) {
        message.success(successMsg);
      } else {
        throw new Error('复制失败');
      }
      return;
    }

    // 尝试使用降级方案
    if (fallbackCopy()) {
      message.success(successMsg);
    } else {
      message.error('复制失败');
    }
  } catch (err) {
    // 最终降级方案
    if (fallbackCopy()) {
      message.success(successMsg);
    } else {
      message.error('复制失败');
    }
  }
};
