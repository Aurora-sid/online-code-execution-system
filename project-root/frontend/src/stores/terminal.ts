import { defineStore } from 'pinia'
import { Terminal } from 'xterm'

export const useTerminalStore = defineStore('terminal', {
  state: () => ({
    terminal: null as Terminal | null,
    ws: null as WebSocket | null,
    isConnected: false,
    isRunning: false
  }),
  
  actions: {
    // 设置终端实例
    setTerminal(terminal: Terminal) {
      this.terminal = terminal
    },
    
    // 连接WebSocket
    async connect() {
      if (this.isConnected) return
      
      return new Promise<void>((resolve, reject) => {
        try {
          // 创建WebSocket连接
          const wsUrl = import.meta.env.DEV 
            ? 'ws://localhost:8080/ws/run' 
            : `wss://${window.location.host}/ws/run`
          
          this.ws = new WebSocket(wsUrl)
          
          // 连接打开
          this.ws.onopen = () => {
            this.isConnected = true
            this.writeLine('WebSocket连接已建立')
            resolve()
          }
          
          // 接收消息
          this.ws.onmessage = (event) => {
            this.writeLine(event.data)
          }
          
          // 连接关闭
          this.ws.onclose = () => {
            this.isConnected = false
            this.isRunning = false
            this.writeLine('WebSocket连接已关闭')
          }
          
          // 连接错误
          this.ws.onerror = (error) => {
            this.isConnected = false
            this.isRunning = false
            this.writeLine(`WebSocket错误: ${error}`)
            reject(error)
          }
        } catch (error) {
          this.writeLine(`连接失败: ${error}`)
          reject(error)
        }
      })
    },
    
    // 断开WebSocket连接
    disconnect() {
      if (this.ws) {
        this.ws.close()
        this.ws = null
      }
      this.isConnected = false
      this.isRunning = false
    },
    
    // 发送数据到WebSocket
    sendData(data: string) {
      if (this.ws && this.isConnected) {
        this.ws.send(data)
      }
    },
    
    // 运行代码
    async runCode(code: string, language: string) {
      if (!this.ws || !this.isConnected) {
        await this.connect()
      }
      
      if (this.ws && this.isConnected) {
        this.isRunning = true
        this.writeLine(`开始运行 ${language} 代码...`)
        
        // 发送运行代码命令
        const message = {
          type: 'run',
          code,
          language
        }
        
        this.ws.send(JSON.stringify(message))
      }
    },
    
    // 写入终端
    writeLine(text: string) {
      if (this.terminal) {
        this.terminal.writeln(text)
      }
    },
    
    // 写入终端（不换行）
    write(text: string) {
      if (this.terminal) {
        this.terminal.write(text)
      }
    },
    
    // 清空终端
    clear() {
      if (this.terminal) {
        this.terminal.clear()
      }
    }
  }
})