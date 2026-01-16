import { defineStore } from 'pinia'

export const useCodeStore = defineStore('code', {
  state: () => ({
    code: {
      python: 'print("Hello, World!")',
      go: 'package main\n\nimport "fmt"\n\nfunc main() {\n    fmt.Println("Hello, World!")\n}',
      java: 'public class Main {\n    public static void main(String[] args) {\n        System.out.println("Hello, World!");\n    }\n}'
    },
    language: 'python',
    history: [] as Array<{
      id: string
      code: string
      language: string
      timestamp: number
      result?: string
    }>
  }),
  
  getters: {
    currentCode: (state) => state.code[state.language as keyof typeof state.code],
    recentHistory: (state) => {
      return [...state.history].sort((a, b) => b.timestamp - a.timestamp).slice(0, 10)
    }
  },
  
  actions: {
    updateCode(language: string, code: string) {
      this.code[language as keyof typeof this.code] = code
    },
    changeLanguage(language: string) {
      this.language = language
    },
    saveToHistory(code: string, language: string, result?: string) {
      this.history.push({
        id: Date.now().toString(),
        code,
        language,
        timestamp: Date.now(),
        result
      })
      
      // 限制历史记录数量
      if (this.history.length > 100) {
        this.history.shift()
      }
    }
  }
})