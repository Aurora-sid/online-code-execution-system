const WebSocket = require('ws');

// 创建WebSocket连接
const ws = new WebSocket('ws://localhost:8082/ws/run');

// 连接打开事件
ws.on('open', () => {
  console.log('WebSocket连接已建立');
  
  // 测试代码执行
  const codeRunRequest = {
    type: 'run',
    code: 'print("Hello, Docker!")',
    language: 'python'
  };
  
  console.log('发送代码执行请求:', codeRunRequest);
  ws.send(JSON.stringify(codeRunRequest));
});

// 接收消息事件
ws.on('message', (data) => {
  console.log(`收到消息: ${data}`);
});

// 连接关闭事件
ws.on('close', () => {
  console.log('WebSocket连接已关闭');
});

// 错误事件
ws.on('error', (error) => {
  console.error(`WebSocket错误: ${error}`);
});